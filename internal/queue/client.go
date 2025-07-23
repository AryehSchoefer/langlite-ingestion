package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	redis *redis.Client
}

func NewClient(redisClient *redis.Client) *Client {
	return &Client{
		redis: redisClient,
	}
}

func (c *Client) Enqueue(ctx context.Context, jobType JobType, priority QueuePriority, payload map[string]interface{}) (*Job, error) {
	job := &Job{
		ID:          uuid.New().String(),
		Type:        jobType,
		Priority:    priority,
		Payload:     payload,
		CreatedAt:   time.Now().UTC(),
		Attempts:    0,
		MaxAttempts: 3, // Default max attempts
	}

	jobJSON, err := job.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize job: %w", err)
	}

	queueName := GetQueueName(jobType, priority)

	err = c.redis.LPush(ctx, queueName, jobJSON).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	trackingKey := fmt.Sprintf("jobs:tracking:%s", job.ID)
	err = c.redis.Set(ctx, trackingKey, jobJSON, 24*time.Hour).Err()
	if err != nil {
		// log error brt don't fail the enqueue operation
		fmt.Printf("Warning: failed to add job to tracking: %v\n", err)
	}

	return job, nil
}

func (c *Client) Dequeue(ctx context.Context, jobTypes []JobType, timeout time.Duration) (*Job, error) {
	// build queue names in priority order (high -> medium -> low)
	var queueNames []string
	priorities := []QueuePriority{QueueHigh, QueueMedium, QueueLow}

	for _, priority := range priorities {
		for _, jobType := range jobTypes {
			queueNames = append(queueNames, GetQueueName(jobType, priority))
		}
	}

	// use BRPOP to block until a job is available
	result, err := c.redis.BRPop(ctx, timeout, queueNames...).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No job available within timeout
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	// result[0] is the queue name, result[1] is the job JSON
	jobJSON := result[1]
	job, err := FromJSON(jobJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}

	job.Attempts++

	return job, nil
}

func (c *Client) CompleteJob(ctx context.Context, job *Job, result *JobResult) error {
	job.ProcessedAt = &result.ProcessedAt

	trackingKey := fmt.Sprintf("jobs:tracking:%s", job.ID)
	err := c.redis.Del(ctx, trackingKey).Err()
	if err != nil {
		fmt.Printf("Warning: failed to remove job from tracking: %v\n", err)
	}

	completedKey := fmt.Sprintf("jobs:completed:%s", job.ID)
	jobJSON, _ := job.ToJSON()
	err = c.redis.Set(ctx, completedKey, jobJSON, time.Hour).Err()
	if err != nil {
		fmt.Printf("Warning: failed to add job to completed: %v\n", err)
	}

	return nil
}

func (c *Client) FailJob(ctx context.Context, job *Job, errorMsg string) error {
	job.Error = errorMsg

	if job.Attempts < job.MaxAttempts {
		return c.retryJob(ctx, job)
	}

	return c.moveToDeadLetter(ctx, job)
}

func (c *Client) retryJob(ctx context.Context, job *Job) error {
	delay := time.Duration(job.Attempts*job.Attempts) * time.Second

	jobJSON, err := job.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize job for retry: %w", err)
	}

	delayedKey := "jobs:delayed"
	score := float64(time.Now().Add(delay).Unix())

	err = c.redis.ZAdd(ctx, delayedKey, redis.Z{
		Score:  score,
		Member: jobJSON,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to schedule job retry: %w", err)
	}

	return nil
}

func (c *Client) moveToDeadLetter(ctx context.Context, job *Job) error {
	jobJSON, err := job.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize job for dead letter: %w", err)
	}

	deadLetterKey := "jobs:dead_letter"
	err = c.redis.LPush(ctx, deadLetterKey, jobJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to move job to dead letter queue: %w", err)
	}

	trackingKey := fmt.Sprintf("jobs:tracking:%s", job.ID)
	c.redis.Del(ctx, trackingKey)

	return nil
}

func (c *Client) ProcessDelayedJobs(ctx context.Context) error {
	delayedKey := "jobs:delayed"
	now := float64(time.Now().Unix())

	jobs, err := c.redis.ZRangeByScore(ctx, delayedKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to get delayed jobs: %w", err)
	}

	for _, jobJSON := range jobs {
		job, err := FromJSON(jobJSON)
		if err != nil {
			continue
		}

		queueName := GetQueueName(job.Type, job.Priority)
		err = c.redis.LPush(ctx, queueName, jobJSON).Err()
		if err != nil {
			continue
		}

		c.redis.ZRem(ctx, delayedKey, jobJSON)
	}

	return nil
}

func (c *Client) GetQueueStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	jobTypes := []JobType{JobTypeEnrichTrace, JobTypeStoreRaw, JobTypeAnalyticsExport}
	priorities := []QueuePriority{QueueHigh, QueueMedium, QueueLow}

	for _, priority := range priorities {
		for _, jobType := range jobTypes {
			queueName := GetQueueName(jobType, priority)
			length, err := c.redis.LLen(ctx, queueName).Result()
			if err != nil {
				continue
			}
			stats[queueName] = length
		}
	}

	deadLetterLength, _ := c.redis.LLen(ctx, "jobs:dead_letter").Result()
	stats["dead_letter"] = deadLetterLength

	delayedLength, _ := c.redis.ZCard(ctx, "jobs:delayed").Result()
	stats["delayed"] = delayedLength

	return stats, nil
}
