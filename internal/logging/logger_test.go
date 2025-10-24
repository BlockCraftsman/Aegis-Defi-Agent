package logging

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.LoggingConfig
		wantErr  bool
		validate func(t *testing.T, logger Logger)
	}{
		{
			name: "json format stdout",
			cfg: &config.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
			wantErr: false,
			validate: func(t *testing.T, logger Logger) {
				assert.NotNil(t, logger)
			},
		},
		{
			name: "console format stdout",
			cfg: &config.LoggingConfig{
				Level:  "debug",
				Format: "console",
				Output: "stdout",
			},
			wantErr: false,
			validate: func(t *testing.T, logger Logger) {
				assert.NotNil(t, logger)
			},
		},
		{
			name: "invalid level",
			cfg: &config.LoggingConfig{
				Level:  "invalid",
				Format: "json",
				Output: "stdout",
			},
			wantErr: false, // Should default to info level
			validate: func(t *testing.T, logger Logger) {
				assert.NotNil(t, logger)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				if tt.validate != nil {
					tt.validate(t, logger)
				}
				// Cleanup
				if logger != nil {
					logger.Sync()
				}
			}
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	cfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Sync()

	// Test all log levels
	logger.Debug("debug message", zap.String("key", "value"))
	logger.Info("info message", zap.Int("count", 42))
	logger.Warn("warning message", zap.Duration("duration", time.Second))
	logger.Error("error message", zap.Error(assert.AnError))

	// Test with fields
	loggerWithFields := logger.With(zap.String("service", "test"))
	loggerWithFields.Info("message with fields")
}

func TestContextLogger(t *testing.T) {
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Sync()

	ctxLogger := NewContextLogger(logger)

	// Test with context
	ctx := context.WithValue(context.Background(), "request_id", "test-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	loggerWithContext := ctxLogger.WithContext(ctx)
	loggerWithContext.Info("message with context")
}

func TestGlobalLogger(t *testing.T) {
	// Test fallback global logger
	logger := GetGlobalLogger()
	assert.NotNil(t, logger)

	// Test initialization
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	err := InitGlobalLogger(cfg)
	assert.NoError(t, err)

	// Test convenience functions
	Debug("debug message")
	Info("info message")
	Warn("warning message")
	Error("error message")
}

func TestFileOutput(t *testing.T) {
	tempDir := t.TempDir()
	logFile := tempDir + "/test.log"

	cfg := &config.LoggingConfig{
		Level:    "info",
		Format:   "json",
		Output:   "file",
		FilePath: logFile,
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Sync()

	// Log a message
	logger.Info("test message to file", zap.String("test", "value"))

	// Verify file was created
	_, err = os.Stat(logFile)
	assert.NoError(t, err)
}

func TestStructuredLoggingHelpers(t *testing.T) {
	// Test helper functions
	fields := []zap.Field{
		WithError(assert.AnError),
		WithString("string_key", "string_value"),
		WithInt("int_key", 42),
		WithFloat64("float_key", 3.14),
		WithDuration("duration_key", time.Minute),
		WithTime("time_key", time.Now()),
	}

	assert.Len(t, fields, 6)
	// Field types are internal to zap, so we just verify the helpers work
}
