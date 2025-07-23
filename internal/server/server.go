package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"

	"langlite-ingestion/internal/database"
	"langlite-ingestion/internal/metrics"
	"langlite-ingestion/internal/queue"
)

type Server struct {
	port int

	db          database.Service
	redis       *redis.Client
	rateLimiter *RateLimiter
	queueClient *queue.Client
	workerPool  *queue.WorkerPool
	metrics     *metrics.Metrics
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v. Rate limiting will be disabled.", err)
		redisClient = nil
	}

	var rateLimiter *RateLimiter
	var queueClient *queue.Client
	var workerPool *queue.WorkerPool

	if redisClient != nil {
		rateLimiter = NewRateLimiter(redisClient)
		queueClient = queue.NewClient(redisClient)

		workerPool = queue.NewWorkerPool(queueClient, database.New(), 3)

		go func() {
			ctx := context.Background()
			workerPool.Start(ctx)
		}()
	}

	metricsInstance := metrics.NewMetrics()

	NewServer := &Server{
		port:        port,
		db:          database.New(),
		redis:       redisClient,
		rateLimiter: rateLimiter,
		queueClient: queueClient,
		workerPool:  workerPool,
		metrics:     metricsInstance,
	}

	go NewServer.DatabaseMetricsCollector()
	if queueClient != nil {
		go NewServer.QueueMetricsCollector()
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
