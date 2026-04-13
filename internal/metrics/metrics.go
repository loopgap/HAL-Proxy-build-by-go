package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP requests metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hal_proxy_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hal_proxy_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Saturation metrics - Golden Signals
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "hal_proxy_http_requests_in_flight_current",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "hal_proxy_db_connections_active_current",
			Help: "Number of active database connections",
		},
	)

	GoroutineCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "hal_proxy_goroutine_count_current",
			Help: "Number of running goroutines",
		},
	)

	// Error metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hal_proxy_errors_total",
			Help: "Total number of errors by type",
		},
		[]string{"type"},
	)

	// Case metrics
	CasesCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "hal_proxy_cases_created_total",
			Help: "Total number of cases created",
		},
	)

	CasesCompletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "hal_proxy_cases_completed_total",
			Help: "Total number of cases completed",
		},
	)

	CasesRunningGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "hal_proxy_cases_running",
			Help: "Number of cases currently running",
		},
	)

	// Approval metrics
	ApprovalsRequestedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "hal_proxy_approvals_requested_total",
			Help: "Total number of approvals requested",
		},
	)

	ApprovalsApprovedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "hal_proxy_approvals_approved_total",
			Help: "Total number of approvals approved",
		},
	)

	ApprovalsRejectedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "hal_proxy_approvals_rejected_total",
			Help: "Total number of approvals rejected",
		},
	)

	// Command execution metrics
	CommandsExecutedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hal_proxy_commands_executed_total",
			Help: "Total number of commands executed",
		},
		[]string{"risk_class"},
	)

	CommandExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hal_proxy_command_duration_seconds",
			Help:    "Command execution duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		},
		[]string{"risk_class"},
	)

	// Database metrics
	DBOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hal_proxy_db_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "status"},
	)

	DBOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hal_proxy_db_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// StartGoroutinePolling starts a background goroutine that periodically
// updates the GoroutineCount metric with the current number of running goroutines.
func StartGoroutinePolling(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			GoroutineCount.Set(float64(runtime.NumGoroutine()))
		}
	}()
}
