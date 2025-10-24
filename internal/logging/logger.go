package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the main logger interface
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sync() error
}

// zapLogger implements Logger using zap
type zapLogger struct {
	logger *zap.Logger
}

// NewLogger creates a new logger based on configuration
func NewLogger(cfg *config.LoggingConfig) (Logger, error) {
	var zapConfig zap.Config

	switch cfg.Format {
	case "json":
		zapConfig = zap.NewProductionConfig()
	case "console":
		zapConfig = zap.NewDevelopmentConfig()
	default:
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Configure output
	var outputs []string
	switch cfg.Output {
	case "file":
		if cfg.FilePath == "" {
			cfg.FilePath = "/var/log/aegis-defi-agent.log"
		}
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		outputs = []string{cfg.FilePath}
	case "both":
		if cfg.FilePath == "" {
			cfg.FilePath = "/var/log/aegis-defi-agent.log"
		}
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		outputs = []string{"stdout", cfg.FilePath}
	default:
		outputs = []string{"stdout"}
	}
	zapConfig.OutputPaths = outputs
	zapConfig.ErrorOutputPaths = outputs

	// Add custom fields
	zapConfig.InitialFields = map[string]interface{}{
		"service": "aegis-defi-agent",
		"version": "1.0.0",
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &zapLogger{logger: logger}, nil
}

func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// ContextLogger provides logging with context support
type ContextLogger struct {
	Logger
}

// NewContextLogger creates a new context logger
func NewContextLogger(logger Logger) *ContextLogger {
	return &ContextLogger{Logger: logger}
}

// WithContext adds context fields to the logger
func (cl *ContextLogger) WithContext(ctx context.Context) Logger {
	fields := []zap.Field{
		zap.String("request_id", getRequestID(ctx)),
		zap.String("user_id", getUserID(ctx)),
	}
	return cl.With(fields...)
}

// Helper functions for context extraction
func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}

func getUserID(ctx context.Context) string {
	if id, ok := ctx.Value("user_id").(string); ok {
		return id
	}
	return ""
}

// Global logger instance
var globalLogger Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(cfg *config.LoggingConfig) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() Logger {
	if globalLogger == nil {
		// Fallback to basic logger
		cfg := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		}
		logger, _ := NewLogger(cfg)
		return logger
	}
	return globalLogger
}

// Convenience functions for global logger
func Debug(msg string, fields ...zap.Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetGlobalLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetGlobalLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}

// Structured logging helpers
func WithError(err error) zap.Field {
	return zap.Error(err)
}

func WithString(key, value string) zap.Field {
	return zap.String(key, value)
}

func WithInt(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func WithFloat64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

func WithDuration(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

func WithTime(key string, value time.Time) zap.Field {
	return zap.Time(key, value)
}
