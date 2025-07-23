package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"langlite-ingestion/internal/database"
)

type Worker struct {
	id         string
	client     *Client
	processors map[JobType]JobProcessor
	jobTypes   []JobType
	running    bool
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

func NewWorker(id string, client *Client, db database.Service) *Worker {
	processors := make(map[JobType]JobProcessor)

	enrichProcessor := NewEnrichTraceProcessor(db)
	storeProcessor := NewStoreRawProcessor(db)
	analyticsProcessor := NewAnalyticsExportProcessor()

	processors[JobTypeEnrichTrace] = enrichProcessor
	processors[JobTypeStoreRaw] = storeProcessor
	processors[JobTypeAnalyticsExport] = analyticsProcessor

	jobTypes := []JobType{
		JobTypeEnrichTrace,
		JobTypeStoreRaw,
		JobTypeAnalyticsExport,
	}

	return &Worker{
		id:         id,
		client:     client,
		processors: processors,
		jobTypes:   jobTypes,
		stopCh:     make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	if w.running {
		return
	}

	w.running = true
	log.Printf("Worker %s starting...", w.id)

	w.wg.Add(1)
	go w.processLoop(ctx)

	w.wg.Add(1)
	go w.delayedJobLoop(ctx)
}

func (w *Worker) Stop() {
	if !w.running {
		return
	}

	log.Printf("Worker %s stopping...", w.id)
	w.running = false
	close(w.stopCh)
	w.wg.Wait()
	log.Printf("Worker %s stopped", w.id)
}

func (w *Worker) processLoop(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			w.processNextJob(ctx)
		}
	}
}

func (w *Worker) processNextJob(ctx context.Context) {
	job, err := w.client.Dequeue(ctx, w.jobTypes, 5*time.Second)
	if err != nil {
		log.Printf("Worker %s: error dequeuing job: %v", w.id, err)
		return
	}

	if job == nil {
		return
	}

	log.Printf("Worker %s: processing job %s (type: %s, attempt: %d)",
		w.id, job.ID, job.Type, job.Attempts)

	processor, exists := w.processors[job.Type]
	if !exists {
		log.Printf("Worker %s: no processor found for job type %s", w.id, job.Type)
		w.client.FailJob(ctx, job, fmt.Sprintf("no processor for job type %s", job.Type))
		return
	}

	result, err := processor.Process(ctx, job)
	if err != nil {
		log.Printf("Worker %s: processor error for job %s: %v", w.id, job.ID, err)
		w.client.FailJob(ctx, job, err.Error())
		return
	}

	if result.Success {
		log.Printf("Worker %s: job %s completed successfully (duration: %v)",
			w.id, job.ID, result.Duration)
		w.client.CompleteJob(ctx, job, result)
	} else {
		log.Printf("Worker %s: job %s failed: %s (duration: %v)",
			w.id, job.ID, result.Error, result.Duration)
		w.client.FailJob(ctx, job, result.Error)
	}
}

func (w *Worker) delayedJobLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := w.client.ProcessDelayedJobs(ctx)
			if err != nil {
				log.Printf("Worker %s: error processing delayed jobs: %v", w.id, err)
			}
		}
	}
}

type WorkerPool struct {
	workers []*Worker
	client  *Client
	db      database.Service
}

func NewWorkerPool(client *Client, db database.Service, workerCount int) *WorkerPool {
	workers := make([]*Worker, workerCount)

	for i := 0; i < workerCount; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		workers[i] = NewWorker(workerID, client, db)
	}

	return &WorkerPool{
		workers: workers,
		client:  client,
		db:      db,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	log.Printf("Starting worker pool with %d workers", len(wp.workers))

	for _, worker := range wp.workers {
		worker.Start(ctx)
	}
}

func (wp *WorkerPool) Stop() {
	log.Printf("Stopping worker pool...")

	for _, worker := range wp.workers {
		worker.Stop()
	}
}

func (wp *WorkerPool) GetStats(ctx context.Context) (map[string]interface{}, error) {
	queueStats, err := wp.client.GetQueueStats(ctx)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"worker_count": len(wp.workers),
		"queue_stats":  queueStats,
		"workers":      make([]map[string]interface{}, len(wp.workers)),
	}

	for i, worker := range wp.workers {
		workerStats := map[string]interface{}{
			"id":      worker.id,
			"running": worker.running,
		}
		stats["workers"].([]map[string]interface{})[i] = workerStats
	}

	return stats, nil
}
