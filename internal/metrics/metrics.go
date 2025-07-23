package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	// HTTP Request metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight *prometheus.GaugeVec

	// Queue metrics
	QueueDepth        *prometheus.GaugeVec
	QueueJobsTotal    *prometheus.CounterVec
	QueueJobDuration  *prometheus.HistogramVec
	QueueJobsFailures *prometheus.CounterVec

	// Worker metrics
	WorkerJobsProcessed *prometheus.CounterVec
	WorkerJobsActive    *prometheus.GaugeVec
	WorkerStatus        *prometheus.GaugeVec

	// Rate limiting metrics
	RateLimitHits    *prometheus.CounterVec
	RateLimitCurrent *prometheus.GaugeVec

	// Database metrics
	DatabaseConnections   *prometheus.GaugeVec
	DatabaseQueries       *prometheus.CounterVec
	DatabaseQueryDuration *prometheus.HistogramVec

	// Redis metrics
	RedisOperations        *prometheus.CounterVec
	RedisOperationDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),

		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),

		HTTPRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "endpoint"},
		),

		QueueDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "queue_depth",
				Help: "Current depth of job queues",
			},
			[]string{"queue_name", "priority", "job_type"},
		),

		QueueJobsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "queue_jobs_total",
				Help: "Total number of jobs processed by queues",
			},
			[]string{"queue_name", "priority", "job_type", "status"},
		),

		QueueJobDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "queue_job_duration_seconds",
				Help:    "Duration of queue job processing in seconds",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"queue_name", "priority", "job_type"},
		),

		QueueJobsFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "queue_jobs_failures_total",
				Help: "Total number of failed queue jobs",
			},
			[]string{"queue_name", "priority", "job_type", "error_type"},
		),

		WorkerJobsProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "worker_jobs_processed_total",
				Help: "Total number of jobs processed by workers",
			},
			[]string{"worker_id", "job_type", "status"},
		),

		WorkerJobsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "worker_jobs_active",
				Help: "Number of jobs currently being processed by workers",
			},
			[]string{"worker_id"},
		),

		WorkerStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "worker_status",
				Help: "Status of workers (1 = running, 0 = stopped)",
			},
			[]string{"worker_id"},
		),

		RateLimitHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rate_limit_hits_total",
				Help: "Total number of rate limit hits",
			},
			[]string{"api_key_id", "limit_type"},
		),

		RateLimitCurrent: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "rate_limit_current",
				Help: "Current rate limit usage",
			},
			[]string{"api_key_id", "limit_type"},
		),

		DatabaseConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections",
				Help: "Number of database connections",
			},
			[]string{"status"}, // open, idle, in_use
		),

		DatabaseQueries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),

		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5},
			},
			[]string{"operation", "table"},
		),

		RedisOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_operations_total",
				Help: "Total number of Redis operations",
			},
			[]string{"operation", "status"},
		),

		RedisOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "redis_operation_duration_seconds",
				Help:    "Duration of Redis operations in seconds",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1},
			},
			[]string{"operation"},
		),
	}
}

func (m *Metrics) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	status := strconv.Itoa(statusCode)
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
}

func (m *Metrics) RecordHTTPRequestInFlight(method, endpoint string, delta float64) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Add(delta)
}

func (m *Metrics) UpdateQueueDepth(queueName, priority, jobType string, depth float64) {
	m.QueueDepth.WithLabelValues(queueName, priority, jobType).Set(depth)
}

func (m *Metrics) RecordQueueJob(queueName, priority, jobType, status string, duration time.Duration) {
	m.QueueJobsTotal.WithLabelValues(queueName, priority, jobType, status).Inc()
	m.QueueJobDuration.WithLabelValues(queueName, priority, jobType).Observe(duration.Seconds())
}

func (m *Metrics) RecordQueueJobFailure(queueName, priority, jobType, errorType string) {
	m.QueueJobsFailures.WithLabelValues(queueName, priority, jobType, errorType).Inc()
}

func (m *Metrics) RecordWorkerJob(workerID, jobType, status string) {
	m.WorkerJobsProcessed.WithLabelValues(workerID, jobType, status).Inc()
}

func (m *Metrics) UpdateWorkerJobsActive(workerID string, delta float64) {
	m.WorkerJobsActive.WithLabelValues(workerID).Add(delta)
}

func (m *Metrics) UpdateWorkerStatus(workerID string, running bool) {
	status := 0.0
	if running {
		status = 1.0
	}
	m.WorkerStatus.WithLabelValues(workerID).Set(status)
}

func (m *Metrics) RecordRateLimitHit(apiKeyID, limitType string) {
	m.RateLimitHits.WithLabelValues(apiKeyID, limitType).Inc()
}

func (m *Metrics) UpdateRateLimitCurrent(apiKeyID, limitType string, current float64) {
	m.RateLimitCurrent.WithLabelValues(apiKeyID, limitType).Set(current)
}

func (m *Metrics) UpdateDatabaseConnections(open, idle, inUse float64) {
	m.DatabaseConnections.WithLabelValues("open").Set(open)
	m.DatabaseConnections.WithLabelValues("idle").Set(idle)
	m.DatabaseConnections.WithLabelValues("in_use").Set(inUse)
}

func (m *Metrics) RecordDatabaseQuery(operation, table, status string, duration time.Duration) {
	m.DatabaseQueries.WithLabelValues(operation, table, status).Inc()
	m.DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

func (m *Metrics) RecordRedisOperation(operation, status string, duration time.Duration) {
	m.RedisOperations.WithLabelValues(operation, status).Inc()
	m.RedisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}
