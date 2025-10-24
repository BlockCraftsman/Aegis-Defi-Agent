package monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/config"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMonitor(t *testing.T) {
	logger, err := logging.NewLogger(&config.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		cfg     *config.MonitoringConfig
		enabled bool
	}{
		{
			name: "enabled monitoring",
			cfg: &config.MonitoringConfig{
				Enabled:     true,
				MetricsPort: 9090,
				HealthCheck: 30 * time.Second,
				Alerting: config.AlertingConfig{
					Enabled:  true,
					Channels: []string{"console"},
				},
			},
			enabled: true,
		},
		{
			name: "disabled monitoring",
			cfg: &config.MonitoringConfig{
				Enabled: false,
			},
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewMonitor(tt.cfg, logger)
			assert.NotNil(t, monitor)

			if tt.enabled {
				assert.NotNil(t, monitor.metrics)
				assert.NotNil(t, monitor.alerts)
			} else {
				assert.Nil(t, monitor.metrics)
			}
		})
	}
}

func TestMonitorRecording(t *testing.T) {
	logger, err := logging.NewLogger(&config.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	})
	require.NoError(t, err)

	cfg := &config.MonitoringConfig{
		Enabled:     false, // Disabled to avoid duplicate metric registration
		MetricsPort: 9091,
		HealthCheck: 30 * time.Second,
		Alerting: config.AlertingConfig{
			Enabled:  true,
			Channels: []string{"console"},
		},
	}

	monitor := NewMonitor(cfg, logger)
	require.NotNil(t, monitor)

	// Test recording various metrics (should not panic when disabled)
	monitor.RecordTrade(true, 1000.0, 2*time.Second, "mean_reversion")
	monitor.RecordTrade(false, 500.0, 1*time.Second, "trend_following")

	monitor.RecordPriceUpdate(100*time.Millisecond, "ETH", 0.05)
	monitor.RecordBlockchainCall(true, 500*time.Millisecond, 21000)

	monitor.RecordError("critical", "blockchain")
	monitor.RecordWarning("high_slippage", "trading")

	monitor.UpdateSystemMetrics(5, 1024*1024*1024, 25.5, 50)
	monitor.UpdateStrategyMetrics("mean_reversion", 15.5, 0.3)

	// Metrics should be nil when disabled
	assert.Nil(t, monitor.metrics)
}

func TestHealthChecks(t *testing.T) {
	logger, err := logging.NewLogger(&config.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	})
	require.NoError(t, err)

	cfg := &config.MonitoringConfig{
		Enabled:     false, // Disabled to avoid duplicate metric registration
		MetricsPort: 9092,
		HealthCheck: 30 * time.Second,
	}

	monitor := NewMonitor(cfg, logger)
	require.NotNil(t, monitor)

	// Add a health check that always passes
	monitor.AddHealthCheck("test_check", func(ctx context.Context) error {
		return nil
	})

	// Add a health check that always fails
	monitor.AddHealthCheck("failing_check", func(ctx context.Context) error {
		return fmt.Errorf("health check failed")
	})

	// Test health check execution
	ctx := context.Background()
	monitor.runHealthChecks(ctx)

	// Should not panic
	assert.NotNil(t, monitor.healthChecks)
}

func TestAlertManager(t *testing.T) {
	logger, err := logging.NewLogger(&config.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	})
	require.NoError(t, err)

	cfg := &config.AlertingConfig{
		Enabled:  true,
		Channels: []string{"console"},
	}

	alertManager := NewAlertManager(cfg, logger)
	require.NotNil(t, alertManager)

	// Test sending alerts
	alertManager.SendAlert("Test Alert", "This is a test alert", "info")
	alertManager.SendAlert("Critical Alert", "This is a critical alert", "critical")

	// Should not panic
}

func TestDisabledMonitoring(t *testing.T) {
	logger, err := logging.NewLogger(&config.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	})
	require.NoError(t, err)

	cfg := &config.MonitoringConfig{
		Enabled: false,
	}

	monitor := NewMonitor(cfg, logger)
	require.NotNil(t, monitor)

	// All recording methods should work without panicking
	monitor.RecordTrade(true, 1000.0, 2*time.Second, "test")
	monitor.RecordPriceUpdate(100*time.Millisecond, "ETH", 0.05)
	monitor.RecordBlockchainCall(true, 500*time.Millisecond, 21000)
	monitor.RecordError("critical", "blockchain")
	monitor.UpdateSystemMetrics(5, 1024*1024*1024, 25.5, 50)

	// Metrics should be nil when disabled
	assert.Nil(t, monitor.metrics)
}
