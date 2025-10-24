package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ConfigManager provides utility functions for configuration management
type ConfigManager struct {
	config *Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(config *Config) *ConfigManager {
	return &ConfigManager{
		config: config,
	}
}

// GetEnvOrDefault returns environment variable value or default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvIntOrDefault returns environment variable as int or default
func GetEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvDurationOrDefault returns environment variable as duration or default
func GetEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetEnvFloat64OrDefault returns environment variable as float64 or default
func GetEnvFloat64OrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// GetEnvBoolOrDefault returns environment variable as bool or default
func GetEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

// IsProduction returns true if environment is production
func (cm *ConfigManager) IsProduction() bool {
	return cm.config.Environment == "production"
}

// IsDevelopment returns true if environment is development
func (cm *ConfigManager) IsDevelopment() bool {
	return cm.config.Environment == "development"
}

// IsStaging returns true if environment is staging
func (cm *ConfigManager) IsStaging() bool {
	return cm.config.Environment == "staging"
}

// GetNetworkRPCURL returns the RPC URL for a specific network
func (cm *ConfigManager) GetNetworkRPCURL(networkName string) (string, error) {
	network, err := cm.config.GetNetworkConfig(networkName)
	if err != nil {
		return "", err
	}
	return network.RPCURL, nil
}

// GetNetworkWSURL returns the WebSocket URL for a specific network
func (cm *ConfigManager) GetNetworkWSURL(networkName string) (string, error) {
	network, err := cm.config.GetNetworkConfig(networkName)
	if err != nil {
		return "", err
	}
	return network.WSURL, nil
}

// GetEnabledStrategies returns all enabled strategies
func (cm *ConfigManager) GetEnabledStrategies() []StrategyConfig {
	var enabled []StrategyConfig
	for _, strategy := range cm.config.Agents.Strategies {
		if strategy.Enabled {
			enabled = append(enabled, strategy)
		}
	}
	return enabled
}

// GetStrategyByName returns a strategy by name
func (cm *ConfigManager) GetStrategyByName(name string) (*StrategyConfig, error) {
	return cm.config.GetStrategyConfig(name)
}

// GetEnabledDataSources returns all enabled data sources
func (cm *ConfigManager) GetEnabledDataSources() []DataSourceConfig {
	var enabled []DataSourceConfig
	for _, source := range cm.config.MarketData.Sources {
		if source.Enabled {
			enabled = append(enabled, source)
		}
	}
	return enabled
}

// GetDataSourceByName returns a data source by name
func (cm *ConfigManager) GetDataSourceByName(name string) (*DataSourceConfig, error) {
	return cm.config.GetDataSourceConfig(name)
}

// ValidateRiskParameters validates all risk parameters
func (cm *ConfigManager) ValidateRiskParameters() error {
	risk := cm.config.Agents.Risk

	if risk.MaxPositionSize <= 0 || risk.MaxPositionSize > 1 {
		return fmt.Errorf("max position size must be between 0 and 1")
	}

	if risk.MaxSlippage < 0 || risk.MaxSlippage > 1 {
		return fmt.Errorf("max slippage must be between 0 and 1")
	}

	if risk.StopLossPercent < 0 || risk.StopLossPercent > 1 {
		return fmt.Errorf("stop loss percent must be between 0 and 1")
	}

	if risk.TakeProfitPercent < 0 || risk.TakeProfitPercent > 1 {
		return fmt.Errorf("take profit percent must be between 0 and 1")
	}

	if risk.MaxDrawdown < 0 || risk.MaxDrawdown > 1 {
		return fmt.Errorf("max drawdown must be between 0 and 1")
	}

	return nil
}

// GetDefaultNetwork returns the default network configuration
func (cm *ConfigManager) GetDefaultNetwork() (*NetworkConfig, error) {
	return cm.config.GetNetworkConfig(cm.config.Blockchain.DefaultNetwork)
}

// GetStrategyParameter returns a strategy parameter by key
func (cm *ConfigManager) GetStrategyParameter(strategyName, paramKey string) (any, error) {
	strategy, err := cm.GetStrategyByName(strategyName)
	if err != nil {
		return nil, err
	}

	value, exists := strategy.Parameters[paramKey]
	if !exists {
		return nil, fmt.Errorf("parameter '%s' not found in strategy '%s'", paramKey, strategyName)
	}

	return value, nil
}

// GetStrategyParameterFloat64 returns a strategy parameter as float64
func (cm *ConfigManager) GetStrategyParameterFloat64(strategyName, paramKey string) (float64, error) {
	value, err := cm.GetStrategyParameter(strategyName, paramKey)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("parameter '%s' is not a numeric type", paramKey)
	}
}

// GetStrategyParameterInt returns a strategy parameter as int
func (cm *ConfigManager) GetStrategyParameterInt(strategyName, paramKey string) (int, error) {
	value, err := cm.GetStrategyParameter(strategyName, paramKey)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	default:
		return 0, fmt.Errorf("parameter '%s' is not an integer type", paramKey)
	}
}

// GetStrategyParameterString returns a strategy parameter as string
func (cm *ConfigManager) GetStrategyParameterString(strategyName, paramKey string) (string, error) {
	value, err := cm.GetStrategyParameter(strategyName, paramKey)
	if err != nil {
		return "", err
	}

	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// GetStrategyParameterBool returns a strategy parameter as bool
func (cm *ConfigManager) GetStrategyParameterBool(strategyName, paramKey string) (bool, error) {
	value, err := cm.GetStrategyParameter(strategyName, paramKey)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strings.ToLower(v) == "true", nil
	default:
		return false, fmt.Errorf("parameter '%s' is not a boolean type", paramKey)
	}
}

// GetStrategyParameterStringSlice returns a strategy parameter as string slice
func (cm *ConfigManager) GetStrategyParameterStringSlice(strategyName, paramKey string) ([]string, error) {
	value, err := cm.GetStrategyParameter(strategyName, paramKey)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case []any:
		var result []string
		for _, item := range v {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result, nil
	default:
		return nil, fmt.Errorf("parameter '%s' is not a string slice type", paramKey)
	}
}
