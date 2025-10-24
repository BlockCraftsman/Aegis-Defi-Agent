package logging

import (
	"context"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/config"
	"go.uber.org/zap"
)

// ExampleUsage demonstrates how to use the logging system
func ExampleUsage() {
	// Initialize logger with configuration
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Basic logging
	logger.Info("Application started")
	logger.Debug("Debug information", zap.String("component", "example"))

	// Structured logging with fields
	logger.Info("Trade executed",
		zap.String("strategy", "mean_reversion"),
		zap.Float64("amount", 1000.0),
		zap.String("asset", "ETH"),
		zap.Duration("latency", 2*time.Second),
	)

	// Error logging
	logger.Error("Failed to execute trade",
		zap.Error(err),
		zap.String("component", "trading_engine"),
	)

	// Context-aware logging
	ctxLogger := NewContextLogger(logger)
	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	ctxLogger.WithContext(ctx).Info("Request processed")

	// Global logger usage
	if err := InitGlobalLogger(cfg); err != nil {
		panic(err)
	}

	Info("Using global logger")
	Error("Global error", WithError(err))
}

// TradingExample demonstrates logging in trading scenarios
func TradingExample(logger Logger) {
	// Trade execution
	startTime := time.Now()

	// Simulate trade processing
	time.Sleep(100 * time.Millisecond)

	latency := time.Since(startTime)

	logger.Info("Trade completed",
		WithString("strategy", "arbitrage"),
		WithFloat64("profit", 150.25),
		WithDuration("latency", latency),
		WithString("status", "success"),
	)

	// Market data updates
	logger.Debug("Price update received",
		WithString("symbol", "ETH/USD"),
		WithFloat64("price", 3500.75),
		WithTime("timestamp", time.Now()),
	)

	// Risk management alerts
	logger.Warn("High volatility detected",
		WithString("asset", "BTC"),
		WithFloat64("volatility", 0.15),
		WithString("action", "reduce_position"),
	)

	// Blockchain interactions
	logger.Info("Transaction submitted",
		WithString("tx_hash", "0x1234..."),
		WithInt("gas_used", 21000),
		WithFloat64("gas_price", 25.5),
	)
}
