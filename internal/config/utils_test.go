package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigManager(t *testing.T) {
	config := DefaultConfig
	manager := NewConfigManager(&config)

	// Test environment detection
	if !manager.IsDevelopment() {
		t.Error("Expected development environment")
	}

	// Test network URL retrieval
	rpcURL, err := manager.GetNetworkRPCURL("ethereum")
	if err != nil {
		t.Fatalf("Failed to get Ethereum RPC URL: %v", err)
	}
	if rpcURL == "" {
		t.Error("Expected non-empty RPC URL")
	}

	wsURL, err := manager.GetNetworkWSURL("ethereum")
	if err != nil {
		t.Fatalf("Failed to get Ethereum WS URL: %v", err)
	}
	if wsURL == "" {
		t.Error("Expected non-empty WS URL")
	}

	// Test enabled strategies
	enabledStrategies := manager.GetEnabledStrategies()
	if len(enabledStrategies) == 0 {
		t.Error("Expected at least one enabled strategy")
	}

	// Test enabled data sources
	enabledSources := manager.GetEnabledDataSources()
	if len(enabledSources) == 0 {
		t.Error("Expected at least one enabled data source")
	}

	// Test risk parameter validation
	if err := manager.ValidateRiskParameters(); err != nil {
		t.Errorf("Valid risk parameters should not fail validation: %v", err)
	}

	// Test default network
	defaultNetwork, err := manager.GetDefaultNetwork()
	if err != nil {
		t.Fatalf("Failed to get default network: %v", err)
	}
	if defaultNetwork.Name != "ethereum" {
		t.Errorf("Expected default network 'ethereum', got '%s'", defaultNetwork.Name)
	}
}

func TestStrategyParameters(t *testing.T) {
	config := DefaultConfig
	manager := NewConfigManager(&config)

	// Test getting strategy parameters
	lookback, err := manager.GetStrategyParameterInt("mean_reversion", "lookback_period")
	if err != nil {
		t.Fatalf("Failed to get lookback period: %v", err)
	}
	if lookback != 20 {
		t.Errorf("Expected lookback period 20, got %d", lookback)
	}

	threshold, err := manager.GetStrategyParameterFloat64("mean_reversion", "threshold")
	if err != nil {
		t.Fatalf("Failed to get threshold: %v", err)
	}
	if threshold != 2.0 {
		t.Errorf("Expected threshold 2.0, got %f", threshold)
	}

	positionSize, err := manager.GetStrategyParameterFloat64("mean_reversion", "position_size")
	if err != nil {
		t.Fatalf("Failed to get position size: %v", err)
	}
	if positionSize != 0.1 {
		t.Errorf("Expected position size 0.1, got %f", positionSize)
	}

	// Test non-existent parameter
	_, err = manager.GetStrategyParameter("mean_reversion", "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent parameter")
	}

	// Test non-existent strategy
	_, err = manager.GetStrategyParameter("nonexistent", "lookback_period")
	if err == nil {
		t.Error("Expected error for non-existent strategy")
	}
}

func TestEnvironmentHelpers(t *testing.T) {
	// Test GetEnvOrDefault
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	value := GetEnvOrDefault("TEST_VAR", "default")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	defaultValue := GetEnvOrDefault("NONEXISTENT_VAR", "default")
	if defaultValue != "default" {
		t.Errorf("Expected 'default', got '%s'", defaultValue)
	}

	// Test GetEnvIntOrDefault
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	intValue := GetEnvIntOrDefault("TEST_INT", 0)
	if intValue != 42 {
		t.Errorf("Expected 42, got %d", intValue)
	}

	defaultInt := GetEnvIntOrDefault("NONEXISTENT_INT", 100)
	if defaultInt != 100 {
		t.Errorf("Expected 100, got %d", defaultInt)
	}

	// Test GetEnvDurationOrDefault
	os.Setenv("TEST_DURATION", "30s")
	defer os.Unsetenv("TEST_DURATION")

	durationValue := GetEnvDurationOrDefault("TEST_DURATION", 10*time.Second)
	if durationValue != 30*time.Second {
		t.Errorf("Expected 30s, got %v", durationValue)
	}

	defaultDuration := GetEnvDurationOrDefault("NONEXISTENT_DURATION", 60*time.Second)
	if defaultDuration != 60*time.Second {
		t.Errorf("Expected 60s, got %v", defaultDuration)
	}

	// Test GetEnvFloat64OrDefault
	os.Setenv("TEST_FLOAT", "3.14")
	defer os.Unsetenv("TEST_FLOAT")

	floatValue := GetEnvFloat64OrDefault("TEST_FLOAT", 0.0)
	if floatValue != 3.14 {
		t.Errorf("Expected 3.14, got %f", floatValue)
	}

	defaultFloat := GetEnvFloat64OrDefault("NONEXISTENT_FLOAT", 1.0)
	if defaultFloat != 1.0 {
		t.Errorf("Expected 1.0, got %f", defaultFloat)
	}

	// Test GetEnvBoolOrDefault
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	boolValue := GetEnvBoolOrDefault("TEST_BOOL", false)
	if !boolValue {
		t.Error("Expected true, got false")
	}

	defaultBool := GetEnvBoolOrDefault("NONEXISTENT_BOOL", true)
	if !defaultBool {
		t.Error("Expected true, got false")
	}
}

func TestParameterTypeConversions(t *testing.T) {
	config := DefaultConfig
	manager := NewConfigManager(&config)

	// Test string parameter
	strategy, err := manager.GetStrategyByName("mean_reversion")
	if err != nil {
		t.Fatalf("Failed to get strategy: %v", err)
	}

	// Add a string parameter for testing
	strategy.Parameters["test_string"] = "hello"
	strategy.Parameters["test_bool"] = true
	strategy.Parameters["test_slice"] = []string{"a", "b", "c"}

	// Test string parameter
	stringValue, err := manager.GetStrategyParameterString("mean_reversion", "test_string")
	if err != nil {
		t.Fatalf("Failed to get string parameter: %v", err)
	}
	if stringValue != "hello" {
		t.Errorf("Expected 'hello', got '%s'", stringValue)
	}

	// Test bool parameter
	boolValue, err := manager.GetStrategyParameterBool("mean_reversion", "test_bool")
	if err != nil {
		t.Fatalf("Failed to get bool parameter: %v", err)
	}
	if !boolValue {
		t.Error("Expected true, got false")
	}

	// Test string slice parameter
	sliceValue, err := manager.GetStrategyParameterStringSlice("mean_reversion", "test_slice")
	if err != nil {
		t.Fatalf("Failed to get string slice parameter: %v", err)
	}
	if len(sliceValue) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(sliceValue))
	}
	if sliceValue[0] != "a" {
		t.Errorf("Expected 'a', got '%s'", sliceValue[0])
	}
}

func TestInvalidParameterTypes(t *testing.T) {
	config := DefaultConfig
	manager := NewConfigManager(&config)

	// Add an invalid parameter type
	strategy, err := manager.GetStrategyByName("mean_reversion")
	if err != nil {
		t.Fatalf("Failed to get strategy: %v", err)
	}

	strategy.Parameters["invalid_type"] = struct{}{}

	// Test invalid numeric type
	_, err = manager.GetStrategyParameterFloat64("mean_reversion", "invalid_type")
	if err == nil {
		t.Error("Expected error for invalid numeric type")
	}

	// Test invalid integer type
	_, err = manager.GetStrategyParameterInt("mean_reversion", "invalid_type")
	if err == nil {
		t.Error("Expected error for invalid integer type")
	}

	// Test invalid boolean type
	_, err = manager.GetStrategyParameterBool("mean_reversion", "invalid_type")
	if err == nil {
		t.Error("Expected error for invalid boolean type")
	}

	// Test invalid string slice type
	_, err = manager.GetStrategyParameterStringSlice("mean_reversion", "invalid_type")
	if err == nil {
		t.Error("Expected error for invalid string slice type")
	}
}
