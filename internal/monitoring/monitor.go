package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/config"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Monitor represents the monitoring system
type Monitor struct {
	cfg          *config.MonitoringConfig
	logger       logging.Logger
	server       *http.Server
	metrics      *Metrics
	healthChecks map[string]HealthCheck
	alerts       *AlertManager
	mu           sync.RWMutex
	isRunning    bool
}

// Metrics contains all Prometheus metrics
type Metrics struct {
	// Trading metrics
	TradesTotal      prometheus.Counter
	TradesSuccessful prometheus.Counter
	TradesFailed     prometheus.Counter
	TradeVolume      prometheus.Gauge
	TradeLatency     prometheus.Histogram

	// Strategy metrics
	StrategyExecutions  *prometheus.CounterVec
	StrategyPerformance *prometheus.GaugeVec
	StrategyRiskScore   *prometheus.GaugeVec

	// Market data metrics
	PriceUpdates     prometheus.Counter
	PriceLatency     prometheus.Histogram
	MarketVolatility *prometheus.GaugeVec

	// Blockchain metrics
	BlockchainCalls   prometheus.Counter
	BlockchainLatency prometheus.Histogram
	GasUsed           prometheus.Gauge

	// System metrics
	ActiveAgents prometheus.Gauge
	MemoryUsage  prometheus.Gauge
	CPUUsage     prometheus.Gauge
	Goroutines   prometheus.Gauge

	// Error metrics
	ErrorsTotal   *prometheus.CounterVec
	WarningsTotal *prometheus.CounterVec
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// AlertManager handles alerting
type AlertManager struct {
	cfg    *config.AlertingConfig
	logger logging.Logger
}

// NewMonitor creates a new monitoring system
func NewMonitor(cfg *config.MonitoringConfig, logger logging.Logger) *Monitor {
	monitor := &Monitor{
		cfg:          cfg,
		logger:       logger,
		healthChecks: make(map[string]HealthCheck),
		alerts:       NewAlertManager(&cfg.Alerting, logger),
	}

	if !cfg.Enabled {
		return monitor
	}

	monitor.metrics = registerMetrics()

	return monitor
}

// registerMetrics registers all Prometheus metrics
func registerMetrics() *Metrics {
	return &Metrics{
		TradesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "aegis_trades_total",
			Help: "Total number of trades executed",
		}),
		TradesSuccessful: promauto.NewCounter(prometheus.CounterOpts{
			Name: "aegis_trades_successful_total",
			Help: "Total number of successful trades",
		}),
		TradesFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "aegis_trades_failed_total",
			Help: "Total number of failed trades",
		}),
		TradeVolume: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "aegis_trade_volume_usd",
			Help: "Current trade volume in USD",
		}),
		TradeLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "aegis_trade_latency_seconds",
			Help:    "Trade execution latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		StrategyExecutions: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "aegis_strategy_executions_total",
			Help: "Total number of strategy executions",
		}, []string{"strategy", "result"}),
		StrategyPerformance: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "aegis_strategy_performance_pct",
			Help: "Strategy performance percentage",
		}, []string{"strategy"}),
		StrategyRiskScore: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "aegis_strategy_risk_score",
			Help: "Strategy risk score",
		}, []string{"strategy"}),
		PriceUpdates: promauto.NewCounter(prometheus.CounterOpts{
			Name: "aegis_price_updates_total",
			Help: "Total number of price updates",
		}),
		PriceLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "aegis_price_latency_seconds",
			Help:    "Price update latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		MarketVolatility: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "aegis_market_volatility_pct",
			Help: "Market volatility percentage",
		}, []string{"asset"}),
		BlockchainCalls: promauto.NewCounter(prometheus.CounterOpts{
			Name: "aegis_blockchain_calls_total",
			Help: "Total number of blockchain calls",
		}),
		BlockchainLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "aegis_blockchain_latency_seconds",
			Help:    "Blockchain call latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		GasUsed: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "aegis_gas_used",
			Help: "Gas used in transactions",
		}),
		ActiveAgents: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "aegis_active_agents",
			Help: "Number of active trading agents",
		}),
		MemoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "aegis_memory_usage_bytes",
			Help: "Memory usage in bytes",
		}),
		CPUUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "aegis_cpu_usage_percent",
			Help: "CPU usage percentage",
		}),
		Goroutines: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "aegis_goroutines",
			Help: "Number of goroutines",
		}),
		ErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "aegis_errors_total",
			Help: "Total number of errors by type",
		}, []string{"type", "component"}),
		WarningsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "aegis_warnings_total",
			Help: "Total number of warnings by type",
		}, []string{"type", "component"}),
	}
}

// Start starts the monitoring system
func (m *Monitor) Start(ctx context.Context) error {
	if !m.cfg.Enabled {
		m.logger.Info("Monitoring disabled, skipping startup")
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("monitoring system already running")
	}

	// Start metrics server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", m.healthHandler)
	mux.HandleFunc("/ready", m.readyHandler)

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.cfg.MetricsPort),
		Handler: mux,
	}

	go func() {
		m.logger.Info("Starting metrics server", logging.WithString("port", fmt.Sprintf("%d", m.cfg.MetricsPort)))
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Error("Metrics server failed", logging.WithError(err))
		}
	}()

	// Start health check loop
	go m.healthCheckLoop(ctx)

	m.isRunning = true
	m.logger.Info("Monitoring system started")

	return nil
}

// Stop stops the monitoring system
func (m *Monitor) Stop(ctx context.Context) error {
	if !m.cfg.Enabled || !m.isRunning {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.server != nil {
		if err := m.server.Shutdown(ctx); err != nil {
			m.logger.Error("Failed to shutdown metrics server", logging.WithError(err))
			return err
		}
	}

	m.isRunning = false
	m.logger.Info("Monitoring system stopped")
	return nil
}

// AddHealthCheck adds a health check
func (m *Monitor) AddHealthCheck(name string, check HealthCheck) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.healthChecks[name] = check
	m.logger.Debug("Added health check", logging.WithString("name", name))
}

// RecordTrade records trade metrics
func (m *Monitor) RecordTrade(success bool, volume float64, latency time.Duration, strategy string) {
	if !m.cfg.Enabled || m.metrics == nil {
		return
	}

	m.metrics.TradesTotal.Inc()
	if success {
		m.metrics.TradesSuccessful.Inc()
		m.metrics.StrategyExecutions.WithLabelValues(strategy, "success").Inc()
	} else {
		m.metrics.TradesFailed.Inc()
		m.metrics.StrategyExecutions.WithLabelValues(strategy, "failed").Inc()
	}

	m.metrics.TradeVolume.Set(volume)
	m.metrics.TradeLatency.Observe(latency.Seconds())
}

// RecordPriceUpdate records price update metrics
func (m *Monitor) RecordPriceUpdate(latency time.Duration, asset string, volatility float64) {
	if !m.cfg.Enabled || m.metrics == nil {
		return
	}

	m.metrics.PriceUpdates.Inc()
	m.metrics.PriceLatency.Observe(latency.Seconds())
	m.metrics.MarketVolatility.WithLabelValues(asset).Set(volatility)
}

// RecordBlockchainCall records blockchain call metrics
func (m *Monitor) RecordBlockchainCall(success bool, latency time.Duration, gasUsed uint64) {
	if !m.cfg.Enabled || m.metrics == nil {
		return
	}

	m.metrics.BlockchainCalls.Inc()
	m.metrics.BlockchainLatency.Observe(latency.Seconds())
	m.metrics.GasUsed.Set(float64(gasUsed))
}

// RecordError records error metrics
func (m *Monitor) RecordError(errorType, component string) {
	if !m.cfg.Enabled || m.metrics == nil {
		return
	}

	m.metrics.ErrorsTotal.WithLabelValues(errorType, component).Inc()

	// Trigger alert for critical errors
	if errorType == "critical" {
		m.alerts.SendAlert("Critical Error", fmt.Sprintf("Critical error in %s: %s", component, errorType), "error")
	}
}

// RecordWarning records warning metrics
func (m *Monitor) RecordWarning(warningType, component string) {
	if !m.cfg.Enabled || m.metrics == nil {
		return
	}

	m.metrics.WarningsTotal.WithLabelValues(warningType, component).Inc()
}

// UpdateSystemMetrics updates system-level metrics
func (m *Monitor) UpdateSystemMetrics(activeAgents int, memoryUsage, cpuUsage float64, goroutines int) {
	if !m.cfg.Enabled || m.metrics == nil {
		return
	}

	m.metrics.ActiveAgents.Set(float64(activeAgents))
	m.metrics.MemoryUsage.Set(memoryUsage)
	m.metrics.CPUUsage.Set(cpuUsage)
	m.metrics.Goroutines.Set(float64(goroutines))
}

// UpdateStrategyMetrics updates strategy performance metrics
func (m *Monitor) UpdateStrategyMetrics(strategy string, performance, riskScore float64) {
	if !m.cfg.Enabled || m.metrics == nil {
		return
	}

	m.metrics.StrategyPerformance.WithLabelValues(strategy).Set(performance)
	m.metrics.StrategyRiskScore.WithLabelValues(strategy).Set(riskScore)
}

// healthHandler handles health check requests
func (m *Monitor) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Run all health checks
	for name, check := range m.healthChecks {
		if err := check(ctx); err != nil {
			m.logger.Error("Health check failed", logging.WithString("check", name), logging.WithError(err))
			http.Error(w, fmt.Sprintf("Health check %s failed: %v", name, err), http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// readyHandler handles readiness check requests
func (m *Monitor) readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}

// healthCheckLoop runs health checks periodically
func (m *Monitor) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(m.cfg.HealthCheck)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.runHealthChecks(ctx)
		}
	}
}

// runHealthChecks executes all registered health checks
func (m *Monitor) runHealthChecks(ctx context.Context) {
	for name, check := range m.healthChecks {
		if err := check(ctx); err != nil {
			m.logger.Warn("Health check failed", logging.WithString("check", name), logging.WithError(err))
			m.alerts.SendAlert("Health Check Failed", fmt.Sprintf("Health check %s failed: %v", name, err), "warning")
		}
	}
}

// NewAlertManager creates a new alert manager
func NewAlertManager(cfg *config.AlertingConfig, logger logging.Logger) *AlertManager {
	return &AlertManager{
		cfg:    cfg,
		logger: logger,
	}
}

// SendAlert sends an alert
func (am *AlertManager) SendAlert(title, message, severity string) {
	if !am.cfg.Enabled {
		return
	}

	am.logger.Warn("Alert triggered",
		logging.WithString("title", title),
		logging.WithString("message", message),
		logging.WithString("severity", severity),
	)

	// TODO: Implement webhook and channel integrations
	// For now, just log the alert
}
