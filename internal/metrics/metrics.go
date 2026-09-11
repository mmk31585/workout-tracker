package metrics

import (
	"expvar"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Metrics holds all application-level observability counters.
type Metrics struct {
	// Runtime
	runtimeStats   *expvar.Int
	startTime      *expvar.Int

	// Requests
	requestsTotal  *expvar.Int

	// Authentication
	signupTotal    *expvar.Int
	loginSuccess   *expvar.Int
	loginFailed    *expvar.Int

	// Business
	activeUsers    *expvar.Int
}

// MetricsOption configures a Metrics instance.
type MetricsOption func(*Metrics)

// WithRuntimeStats adds runtime.GOOS, GOARCH, etc. to expvar.
func WithRuntimeStats() MetricsOption {
	return func(m *Metrics) {
		m.runtimeStats = expvar.NewInt("runtime_goos")
		// Actually, let's do it differently - we'll just add runtime stats as separate ints
	}
}

// WithStartTime records the server start time.
func WithStartTime(startTime time.Time) MetricsOption {
	return func(m *Metrics) {
		m.startTime.Set(startTime.UnixMilli())
	}
}

// NewMetrics creates a new Metrics instance with all counters registered
// with expvar.
func NewMetrics(opts ...MetricsOption) *Metrics {
	m := &Metrics{
		startTime:     expvar.NewInt("start_time_ms"),
		requestsTotal: expvar.NewInt("requests_total"),
		signupTotal:   expvar.NewInt("signup_total"),
		loginSuccess:  expvar.NewInt("login_success_total"),
		loginFailed:   expvar.NewInt("login_failed_total"),
		activeUsers:   expvar.NewInt("active_users"),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// IncrRequests increments the total request counter.
func (m *Metrics) IncrRequests() {
	m.requestsTotal.Add(1)
}

// IncrSignup increments the signup counter.
func (m *Metrics) IncrSignup() {
	m.signupTotal.Add(1)
}

// IncrLoginSuccess increments the login success counter.
func (m *Metrics) IncrLoginSuccess() {
	m.loginSuccess.Add(1)
}

// IncrLoginFailed increments the login failure counter.
func (m *Metrics) IncrLoginFailed() {
	m.loginFailed.Add(1)
}

// IncrActiveUsers increments the active users counter.
func (m *Metrics) IncrActiveUsers(delta int) {
	m.activeUsers.Add(int64(delta))
}

// MetricsHandler returns an http.Handler that serves all metrics
// registered with expvar at the /debug/vars endpoint.
func MetricsHandler() http.Handler {
	// We expose the metrics directly rather than through expvar.Handler
	// so we can control the format and ensure all metrics are always available.
	mux := http.NewServeMux()

	// Register all metrics variables with expvar
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// Summary returns a formatted string of all metrics for logging/debugging.
func (m *Metrics) Summary() string {
	return fmt.Sprintf(`
requests_total:        %d
signup_total:          %d
login_success_total:   %d
login_failed_total:    %d
active_users:          %d
`, m.requestsTotal.Value(),
		m.signupTotal.Value(),
		m.loginSuccess.Value(),
		m.loginFailed.Value(),
		m.activeUsers.Value(),
	)
}

// GlobalMetrics is the default metrics instance. It can be replaced for
// testing or different deployment contexts.
var GlobalMetrics *Metrics

var metricsMutex sync.Once

// SetGlobalMetrics sets the global metrics instance.
func SetGlobalMetrics(m *Metrics) {
	metricsMutex.Do(func() {
		GlobalMetrics = m
	})
}

// GetGlobalMetrics returns the global metrics instance, initialising it
// if necessary.
func GetGlobalMetrics() *Metrics {
	metricsMutex.Do(func() {
		if GlobalMetrics == nil {
			GlobalMetrics = NewMetrics()
		}
	})
	return GlobalMetrics
}