package metrics

import (
	"expvar"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Metrics struct {
	runtimeStats  *expvar.Int
	startTime     *expvar.Int
	requestsTotal *expvar.Int
	signupTotal   *expvar.Int
	loginSuccess  *expvar.Int
	loginFailed   *expvar.Int
	activeUsers   *expvar.Int
}
type MetricsOption func(*Metrics)

func WithRuntimeStats() MetricsOption {
	return func(m *Metrics) {
		m.runtimeStats = expvar.NewInt("runtime_goos")
	}
}
func WithStartTime(startTime time.Time) MetricsOption {
	return func(m *Metrics) {
		m.startTime.Set(startTime.UnixMilli())
	}
}
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
func (m *Metrics) IncrRequests() {
	m.requestsTotal.Add(1)
}
func (m *Metrics) IncrSignup() {
	m.signupTotal.Add(1)
}
func (m *Metrics) IncrLoginSuccess() {
	m.loginSuccess.Add(1)
}
func (m *Metrics) IncrLoginFailed() {
	m.loginFailed.Add(1)
}
func (m *Metrics) IncrActiveUsers(delta int) {
	m.activeUsers.Add(int64(delta))
}
func MetricsHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}
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

var GlobalMetrics *Metrics

var metricsMutex sync.Once

func SetGlobalMetrics(m *Metrics) {
	metricsMutex.Do(func() {
		GlobalMetrics = m
	})
}
func GetGlobalMetrics() *Metrics {
	metricsMutex.Do(func() {
		if GlobalMetrics == nil {
			GlobalMetrics = NewMetrics()
		}
	})
	return GlobalMetrics
}