package errors

import (
	"errors"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Error count metrics by module, code, and category
	errorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "virtengine_errors_total",
			Help: "Total number of errors by module, code, and category",
		},
		[]string{"module", "code", "category", "severity"},
	)

	// Panic recovery metrics
	panicCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "virtengine_panics_recovered_total",
			Help: "Total number of panics recovered by context",
		},
		[]string{"context"},
	)

	// Retryable error metrics
	retryableErrorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "virtengine_retryable_errors_total",
			Help: "Total number of retryable errors by module and category",
		},
		[]string{"module", "category"},
	)

	metricsOnce sync.Once
)

// InitMetrics initializes error metrics.
// This is called automatically on first use, but can be called explicitly
// to register metrics with a custom registry.
func InitMetrics() {
	metricsOnce.Do(func() {
		// Metrics are auto-registered via promauto
	})
}

// RecordError records an error in metrics.
func RecordError(err error) {
	if err == nil {
		return
	}

	InitMetrics()

	var codedErr *CodedError
	if !errors.As(err, &codedErr) {
		// Non-coded error, record as internal
		errorCount.WithLabelValues("unknown", "0", string(CategoryInternal), string(SeverityError)).Inc()
		return
	}

	errorCount.WithLabelValues(
		codedErr.Module,
		formatCode(codedErr.Code),
		string(codedErr.Category),
		string(codedErr.Severity),
	).Inc()

	if codedErr.Retryable {
		retryableErrorCount.WithLabelValues(codedErr.Module, string(codedErr.Category)).Inc()
	}
}

// RecordPanic records a recovered panic in metrics.
func RecordPanic(context string) {
	InitMetrics()
	panicCount.WithLabelValues(context).Inc()
}

// MetricsCollector is a panic handler that records metrics.
type MetricsCollector struct {
	nextHandler PanicHandler //nolint:unused // Reserved for handler chaining
}

// NewMetricsCollector creates a panic handler that records metrics.
func NewMetricsCollector(next PanicHandler) PanicHandler {
	return func(recovered interface{}, stack []byte) {
		RecordPanic("unknown")
		if next != nil {
			next(recovered, stack)
		}
	}
}

// AsWithMetrics is a wrapper for errors.As that also records the error.
func AsWithMetrics(err error, target interface{}) bool {
	if err == nil {
		return false
	}

	result := errors.As(err, target)
	if result {
		RecordError(err)
	}
	return result
}

// formatCode formats an error code as a string for metric labels.
func formatCode(code uint32) string {
	return fmt.Sprintf("%d", code)
}
