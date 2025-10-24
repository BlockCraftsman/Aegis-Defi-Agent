package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the main application configuration
type Config struct {
	Environment string           `json:"environment" yaml:"environment" env:"ENVIRONMENT"`
	Server      ServerConfig     `json:"server" yaml:"server"`
	Database    DatabaseConfig   `json:"database" yaml:"database"`
	Blockchain  BlockchainConfig `json:"blockchain" yaml:"blockchain"`
	MarketData  MarketDataConfig `json:"market_data" yaml:"market_data"`
	Agents      AgentsConfig     `json:"agents" yaml:"agents"`
	Logging     LoggingConfig    `json:"logging" yaml:"logging"`
	Monitoring  MonitoringConfig `json:"monitoring" yaml:"monitoring"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Host           string        `json:"host" yaml:"host" env:"SERVER_HOST"`
	Port           int           `json:"port" yaml:"port" env:"SERVER_PORT"`
	MaxConnections int           `json:"max_connections" yaml:"max_connections" env:"MAX_CONNECTIONS"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout" env:"SERVER_TIMEOUT"`
	CORS           CORSConfig    `json:"cors" yaml:"cors"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins" env:"CORS_ALLOWED_ORIGINS"`
	AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods" env:"CORS_ALLOWED_METHODS"`
	AllowedHeaders []string `json:"allowed_headers" yaml:"allowed_headers" env:"CORS_ALLOWED_HEADERS"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Driver   string `json:"driver" yaml:"driver" env:"DB_DRIVER"`
	Host     string `json:"host" yaml:"host" env:"DB_HOST"`
	Port     int    `json:"port" yaml:"port" env:"DB_PORT"`
	Name     string `json:"name" yaml:"name" env:"DB_NAME"`
	Username string `json:"username" yaml:"username" env:"DB_USERNAME"`
	Password string `json:"password" yaml:"password" env:"DB_PASSWORD"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode" env:"DB_SSL_MODE"`
}

// BlockchainConfig contains blockchain-related configuration
type BlockchainConfig struct {
	Networks       []NetworkConfig `json:"networks" yaml:"networks"`
	DefaultNetwork string          `json:"default_network" yaml:"default_network" env:"DEFAULT_NETWORK"`
	GasPrice       int64           `json:"gas_price" yaml:"gas_price" env:"GAS_PRICE"`
	GasLimit       uint64          `json:"gas_limit" yaml:"gas_limit" env:"GAS_LIMIT"`
	Confirmations  int             `json:"confirmations" yaml:"confirmations" env:"CONFIRMATIONS"`
	PrivateKey     string          `json:"private_key" yaml:"private_key" env:"PRIVATE_KEY"`
}

// NetworkConfig contains configuration for a specific blockchain network
type NetworkConfig struct {
	Name     string `json:"name" yaml:"name"`
	ChainID  int64  `json:"chain_id" yaml:"chain_id"`
	RPCURL   string `json:"rpc_url" yaml:"rpc_url"`
	WSURL    string `json:"ws_url" yaml:"ws_url"`
	Explorer string `json:"explorer" yaml:"explorer"`
}

// MarketDataConfig contains market data configuration
type MarketDataConfig struct {
	Sources        []DataSourceConfig `json:"sources" yaml:"sources"`
	UpdateInterval time.Duration      `json:"update_interval" yaml:"update_interval" env:"MARKET_UPDATE_INTERVAL"`
	PriceFeeds     []string           `json:"price_feeds" yaml:"price_feeds" env:"PRICE_FEEDS"`
	WebSocket      WebSocketConfig    `json:"websocket" yaml:"websocket"`
}

// DataSourceConfig contains configuration for a data source
type DataSourceConfig struct {
	Name     string `json:"name" yaml:"name"`
	Type     string `json:"type" yaml:"type"`
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	APIKey   string `json:"api_key" yaml:"api_key" env:"API_KEY"`
	Enabled  bool   `json:"enabled" yaml:"enabled"`
}

// WebSocketConfig contains WebSocket configuration
type WebSocketConfig struct {
	Enabled       bool          `json:"enabled" yaml:"enabled" env:"WS_ENABLED"`
	Port          int           `json:"port" yaml:"port" env:"WS_PORT"`
	BroadcastRate time.Duration `json:"broadcast_rate" yaml:"broadcast_rate" env:"WS_BROADCAST_RATE"`
	MaxClients    int           `json:"max_clients" yaml:"max_clients" env:"WS_MAX_CLIENTS"`
}

// AgentsConfig contains agent-related configuration
type AgentsConfig struct {
	MaxConcurrent int              `json:"max_concurrent" yaml:"max_concurrent" env:"MAX_CONCURRENT_AGENTS"`
	Strategies    []StrategyConfig `json:"strategies" yaml:"strategies"`
	Risk          RiskConfig       `json:"risk" yaml:"risk"`
}

// StrategyConfig contains configuration for a trading strategy
type StrategyConfig struct {
	Name        string         `json:"name" yaml:"name"`
	Type        string         `json:"type" yaml:"type"`
	Enabled     bool           `json:"enabled" yaml:"enabled"`
	Parameters  map[string]any `json:"parameters" yaml:"parameters"`
	Assets      []string       `json:"assets" yaml:"assets"`
	RiskProfile string         `json:"risk_profile" yaml:"risk_profile"`
}

// RiskConfig contains risk management configuration
type RiskConfig struct {
	MaxPositionSize   float64 `json:"max_position_size" yaml:"max_position_size" env:"MAX_POSITION_SIZE"`
	MaxSlippage       float64 `json:"max_slippage" yaml:"max_slippage" env:"MAX_SLIPPAGE"`
	StopLossPercent   float64 `json:"stop_loss_percent" yaml:"stop_loss_percent" env:"STOP_LOSS_PERCENT"`
	TakeProfitPercent float64 `json:"take_profit_percent" yaml:"take_profit_percent" env:"TAKE_PROFIT_PERCENT"`
	MaxDrawdown       float64 `json:"max_drawdown" yaml:"max_drawdown" env:"MAX_DRAWDOWN"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level    string `json:"level" yaml:"level" env:"LOG_LEVEL"`
	Format   string `json:"format" yaml:"format" env:"LOG_FORMAT"`
	Output   string `json:"output" yaml:"output" env:"LOG_OUTPUT"`
	FilePath string `json:"file_path" yaml:"file_path" env:"LOG_FILE_PATH"`
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	Enabled     bool           `json:"enabled" yaml:"enabled" env:"MONITORING_ENABLED"`
	MetricsPort int            `json:"metrics_port" yaml:"metrics_port" env:"METRICS_PORT"`
	HealthCheck time.Duration  `json:"health_check" yaml:"health_check" env:"HEALTH_CHECK_INTERVAL"`
	Alerting    AlertingConfig `json:"alerting" yaml:"alerting"`
}

// AlertingConfig contains alerting configuration
type AlertingConfig struct {
	Enabled    bool     `json:"enabled" yaml:"enabled" env:"ALERTING_ENABLED"`
	WebhookURL string   `json:"webhook_url" yaml:"webhook_url" env:"ALERT_WEBHOOK_URL"`
	Channels   []string `json:"channels" yaml:"channels" env:"ALERT_CHANNELS"`
}

// DefaultConfig returns a default configuration
var DefaultConfig = Config{
	Environment: "development",
	Server: ServerConfig{
		Host:           "0.0.0.0",
		Port:           8080,
		MaxConnections: 100,
		Timeout:        30 * time.Second,
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"*"},
		},
	},
	Database: DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Name:     "aegis_defi",
		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",
	},
	Blockchain: BlockchainConfig{
		DefaultNetwork: "ethereum",
		GasPrice:       25,
		GasLimit:       21000,
		Confirmations:  3,
		Networks: []NetworkConfig{
			{
				Name:     "ethereum",
				ChainID:  1,
				RPCURL:   "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
				WSURL:    "wss://mainnet.infura.io/ws/v3/YOUR_PROJECT_ID",
				Explorer: "https://etherscan.io",
			},
			{
				Name:     "polygon",
				ChainID:  137,
				RPCURL:   "https://polygon-rpc.com",
				WSURL:    "wss://polygon-rpc.com",
				Explorer: "https://polygonscan.com",
			},
		},
	},
	MarketData: MarketDataConfig{
		UpdateInterval: 30 * time.Second,
		PriceFeeds:     []string{"ETH/USD", "BTC/USD", "USDC/USD"},
		Sources: []DataSourceConfig{
			{
				Name:     "pyth",
				Type:     "oracle",
				Endpoint: "https://hermes.pyth.network",
				Enabled:  true,
			},
			{
				Name:     "coingecko",
				Type:     "api",
				Endpoint: "https://api.coingecko.com/api/v3",
				Enabled:  true,
			},
		},
		WebSocket: WebSocketConfig{
			Enabled:       true,
			Port:          8081,
			BroadcastRate: 5 * time.Second,
			MaxClients:    100,
		},
	},
	Agents: AgentsConfig{
		MaxConcurrent: 10,
		Strategies: []StrategyConfig{
			{
				Name:        "mean_reversion",
				Type:        "mean_reversion",
				Enabled:     true,
				Assets:      []string{"ETH", "BTC"},
				RiskProfile: "medium",
				Parameters: map[string]any{
					"lookback_period": 20,
					"threshold":       2.0,
					"position_size":   0.1,
				},
			},
			{
				Name:        "trend_following",
				Type:        "trend_following",
				Enabled:     true,
				Assets:      []string{"BTC", "ETH"},
				RiskProfile: "high",
				Parameters: map[string]any{
					"short_window":  10,
					"long_window":   50,
					"position_size": 0.05,
				},
			},
		},
		Risk: RiskConfig{
			MaxPositionSize:   0.1,
			MaxSlippage:       0.005,
			StopLossPercent:   0.05,
			TakeProfitPercent: 0.1,
			MaxDrawdown:       0.15,
		},
	},
	Logging: LoggingConfig{
		Level:    "info",
		Format:   "json",
		Output:   "stdout",
		FilePath: "/var/log/aegis-defi-agent.log",
	},
	Monitoring: MonitoringConfig{
		Enabled:     true,
		MetricsPort: 9090,
		HealthCheck: 300 * time.Second,
		Alerting: AlertingConfig{
			Enabled:    true,
			WebhookURL: "",
			Channels:   []string{"console"},
		},
	},
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	var config Config

	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// If no config path provided, use default locations
	if configPath == "" {
		configPath = findConfigFile()
	}

	// Load configuration from file if it exists
	if configPath != "" {
		if err := loadFromFile(configPath, &config); err != nil {
			return nil, fmt.Errorf("error loading config file: %w", err)
		}
	} else {
		// Use default configuration
		config = DefaultConfig
	}

	// Override with environment variables
	if err := loadFromEnv(&config); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// findConfigFile searches for configuration files in common locations
func findConfigFile() string {
	configPaths := []string{
		"config.yaml",
		"config.yml",
		"config.json",
		"config/config.yaml",
		"config/config.yml",
		"config/config.json",
		"/etc/aegis-defi-agent/config.yaml",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// loadFromFile loads configuration from a file
func loadFromFile(filePath string, config *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	ext := filepath.Ext(filePath)
	switch strings.ToLower(ext) {
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, config)
	case ".json":
		return json.Unmarshal(data, config)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) error {
	// This is a simplified implementation
	// In a real implementation, you would use a library like envconfig
	// or implement proper environment variable parsing

	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	// Add more environment variable parsing as needed

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Environment == "" {
		return fmt.Errorf("environment must be set")
	}

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Agents.MaxConcurrent <= 0 {
		return fmt.Errorf("max concurrent agents must be positive")
	}

	if c.Blockchain.GasPrice <= 0 {
		return fmt.Errorf("gas price must be positive")
	}

	if c.Blockchain.GasLimit <= 0 {
		return fmt.Errorf("gas limit must be positive")
	}

	// Validate risk parameters
	if c.Agents.Risk.MaxPositionSize <= 0 || c.Agents.Risk.MaxPositionSize > 1 {
		return fmt.Errorf("max position size must be between 0 and 1")
	}

	if c.Agents.Risk.MaxSlippage < 0 || c.Agents.Risk.MaxSlippage > 1 {
		return fmt.Errorf("max slippage must be between 0 and 1")
	}

	if c.Agents.Risk.StopLossPercent < 0 || c.Agents.Risk.StopLossPercent > 1 {
		return fmt.Errorf("stop loss percent must be between 0 and 1")
	}

	if c.Agents.Risk.TakeProfitPercent < 0 || c.Agents.Risk.TakeProfitPercent > 1 {
		return fmt.Errorf("take profit percent must be between 0 and 1")
	}

	return nil
}

// Save saves the configuration to a file
func (c *Config) Save(filePath string) error {
	var data []byte
	var err error

	ext := filepath.Ext(filePath)
	switch strings.ToLower(ext) {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(c)
	case ".json":
		data, err = json.MarshalIndent(c, "", "  ")
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// GetNetworkConfig returns the configuration for a specific network
func (c *Config) GetNetworkConfig(networkName string) (*NetworkConfig, error) {
	for _, network := range c.Blockchain.Networks {
		if network.Name == networkName {
			return &network, nil
		}
	}
	return nil, fmt.Errorf("network not found: %s", networkName)
}

// GetDataSourceConfig returns the configuration for a specific data source
func (c *Config) GetDataSourceConfig(sourceName string) (*DataSourceConfig, error) {
	for _, source := range c.MarketData.Sources {
		if source.Name == sourceName {
			return &source, nil
		}
	}
	return nil, fmt.Errorf("data source not found: %s", sourceName)
}

// GetStrategyConfig returns the configuration for a specific strategy
func (c *Config) GetStrategyConfig(strategyName string) (*StrategyConfig, error) {
	for _, strategy := range c.Agents.Strategies {
		if strategy.Name == strategyName {
			return &strategy, nil
		}
	}
	return nil, fmt.Errorf("strategy not found: %s", strategyName)
}
