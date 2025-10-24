package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig

	if config.Environment != "development" {
		t.Errorf("Expected environment 'development', got '%s'", config.Environment)
	}

	if config.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", config.Server.Port)
	}

	if config.Agents.MaxConcurrent != 10 {
		t.Errorf("Expected max concurrent agents 10, got %d", config.Agents.MaxConcurrent)
	}

	if config.Blockchain.GasPrice != 25 {
		t.Errorf("Expected gas price 25, got %d", config.Blockchain.GasPrice)
	}

	if config.Agents.Risk.MaxPositionSize != 0.1 {
		t.Errorf("Expected max position size 0.1, got %f", config.Agents.Risk.MaxPositionSize)
	}
}

func TestLoadConfig(t *testing.T) {
	// Test loading default config
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Test loading from non-existent file
	_, err = LoadConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Expected error when loading from non-existent file")
	}
}

func TestConfigValidation(t *testing.T) {
	config := DefaultConfig

	// Test valid config
	if err := config.Validate(); err != nil {
		t.Errorf("Valid config should not fail validation: %v", err)
	}

	// Test invalid environment
	config.Environment = ""
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for empty environment")
	}
	config.Environment = "development"

	// Test invalid server port
	config.Server.Port = 0
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for invalid port")
	}
	config.Server.Port = 8080

	// Test invalid max concurrent agents
	config.Agents.MaxConcurrent = 0
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for zero max concurrent agents")
	}
	config.Agents.MaxConcurrent = 10

	// Test invalid gas price
	config.Blockchain.GasPrice = 0
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for zero gas price")
	}
	config.Blockchain.GasPrice = 25

	// Test invalid risk parameters
	config.Agents.Risk.MaxPositionSize = 1.5
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for max position size > 1")
	}
	config.Agents.Risk.MaxPositionSize = 0.1

	config.Agents.Risk.MaxSlippage = 1.5
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for max slippage > 1")
	}
	config.Agents.Risk.MaxSlippage = 0.005

	config.Agents.Risk.StopLossPercent = 1.5
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for stop loss percent > 1")
	}
	config.Agents.Risk.StopLossPercent = 0.05

	config.Agents.Risk.TakeProfitPercent = 1.5
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for take profit percent > 1")
	}
	config.Agents.Risk.TakeProfitPercent = 0.1
}

func TestEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("SERVER_PORT", "9090")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("SERVER_PORT")
	}()

	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config with environment variables: %v", err)
	}

	if config.Environment != "production" {
		t.Errorf("Expected environment 'production', got '%s'", config.Environment)
	}

	if config.Server.Port != 9090 {
		t.Errorf("Expected server port 9090, got %d", config.Server.Port)
	}
}

func TestGetNetworkConfig(t *testing.T) {
	config := DefaultConfig

	// Test existing network
	network, err := config.GetNetworkConfig("ethereum")
	if err != nil {
		t.Fatalf("Failed to get ethereum network config: %v", err)
	}

	if network.Name != "ethereum" {
		t.Errorf("Expected network name 'ethereum', got '%s'", network.Name)
	}

	if network.ChainID != 1 {
		t.Errorf("Expected chain ID 1, got %d", network.ChainID)
	}

	// Test non-existent network
	_, err = config.GetNetworkConfig("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent network")
	}
}

func TestGetDataSourceConfig(t *testing.T) {
	config := DefaultConfig

	// Test existing data source
	source, err := config.GetDataSourceConfig("pyth")
	if err != nil {
		t.Fatalf("Failed to get pyth data source config: %v", err)
	}

	if source.Name != "pyth" {
		t.Errorf("Expected data source name 'pyth', got '%s'", source.Name)
	}

	if !source.Enabled {
		t.Error("Expected pyth data source to be enabled")
	}

	// Test non-existent data source
	_, err = config.GetDataSourceConfig("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent data source")
	}
}

func TestGetStrategyConfig(t *testing.T) {
	config := DefaultConfig

	// Test existing strategy
	strategy, err := config.GetStrategyConfig("mean_reversion")
	if err != nil {
		t.Fatalf("Failed to get mean reversion strategy config: %v", err)
	}

	if strategy.Name != "mean_reversion" {
		t.Errorf("Expected strategy name 'mean_reversion', got '%s'", strategy.Name)
	}

	if strategy.RiskProfile != "medium" {
		t.Errorf("Expected risk profile 'medium', got '%s'", strategy.RiskProfile)
	}

	// Test non-existent strategy
	_, err = config.GetStrategyConfig("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent strategy")
	}
}

func TestSaveConfig(t *testing.T) {
	config := DefaultConfig

	// Test saving to YAML
	tempFile := "/tmp/test_config.yaml"
	defer os.Remove(tempFile)

	if err := config.Save(tempFile); err != nil {
		t.Fatalf("Failed to save config to YAML: %v", err)
	}

	// Test loading saved config
	loadedConfig, err := LoadConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Environment != config.Environment {
		t.Errorf("Loaded environment '%s' doesn't match saved '%s'", loadedConfig.Environment, config.Environment)
	}

	if loadedConfig.Server.Port != config.Server.Port {
		t.Errorf("Loaded port %d doesn't match saved %d", loadedConfig.Server.Port, config.Server.Port)
	}

	// Test saving to JSON
	tempJSONFile := "/tmp/test_config.json"
	defer os.Remove(tempJSONFile)

	if err := config.Save(tempJSONFile); err != nil {
		t.Fatalf("Failed to save config to JSON: %v", err)
	}

	// Test unsupported format
	unsupportedFile := "/tmp/test_config.txt"
	if err := config.Save(unsupportedFile); err == nil {
		t.Error("Expected error when saving to unsupported format")
	}
}

func TestConfigStructure(t *testing.T) {
	config := DefaultConfig

	// Test server config
	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected server host '0.0.0.0', got '%s'", config.Server.Host)
	}

	if config.Server.Timeout != 30*time.Second {
		t.Errorf("Expected server timeout 30s, got %v", config.Server.Timeout)
	}

	// Test market data config
	if config.MarketData.UpdateInterval != 30*time.Second {
		t.Errorf("Expected market data update interval 30s, got %v", config.MarketData.UpdateInterval)
	}

	if !config.MarketData.WebSocket.Enabled {
		t.Error("Expected WebSocket to be enabled")
	}

	if config.MarketData.WebSocket.Port != 8081 {
		t.Errorf("Expected WebSocket port 8081, got %d", config.MarketData.WebSocket.Port)
	}

	// Test logging config
	if config.Logging.Level != "info" {
		t.Errorf("Expected log level 'info', got '%s'", config.Logging.Level)
	}

	if config.Logging.Format != "json" {
		t.Errorf("Expected log format 'json', got '%s'", config.Logging.Format)
	}

	// Test monitoring config
	if !config.Monitoring.Enabled {
		t.Error("Expected monitoring to be enabled")
	}

	if config.Monitoring.MetricsPort != 9090 {
		t.Errorf("Expected metrics port 9090, got %d", config.Monitoring.MetricsPort)
	}

	if config.Monitoring.HealthCheck != 300*time.Second {
		t.Errorf("Expected health check interval 300s, got %v", config.Monitoring.HealthCheck)
	}
}
