package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Trading     TradingConfig     `yaml:"trading"`
	Binance     BinanceConfig     `yaml:"binance"`
	Risk        RiskConfig        `yaml:"risk"`
	Indicators  IndicatorConfig   `yaml:"indicators"`
	Strategies  StrategiesConfig  `yaml:"strategies"`
	Database    DatabaseConfig    `yaml:"database"`
	Postgres    PostgresConfig    `yaml:"postgres"`
	Auth        AuthConfig        `yaml:"auth"`
	DataService DataServiceConfig `yaml:"dataService"`
	API         APIConfig         `yaml:"api"`
}

// TradingConfig represents trading configuration
type TradingConfig struct {
	Mode             string   `yaml:"mode"`             // "paper" or "live"
	Symbol           string   `yaml:"symbol"`           // e.g., "ETHUSDT"
	Timeframes       []string `yaml:"timeframes"`       // e.g., ["1m", "5m", "15m", "1h", "4h", "1d"]
	PrimaryTimeframe string   `yaml:"primaryTimeframe"` // e.g., "1h"
	InitialBalance   float64  `yaml:"initialBalance"`   // Paper trading initial balance
	Commission       float64  `yaml:"commission"`       // Commission rate (0.001 = 0.1%)
	Slippage         float64  `yaml:"slippage"`         // Slippage rate
}

// BinanceConfig represents Binance API configuration
type BinanceConfig struct {
	APIKey    string `yaml:"apiKey"`
	SecretKey string `yaml:"secretKey"`
	Testnet   bool   `yaml:"testnet"`
}

// RiskConfig represents risk management configuration
type RiskConfig struct {
	MaxPositionSize      float64 `yaml:"maxPositionSize"`      // Max position as % of equity (0.1 = 10%)
	MaxRiskPerTrade      float64 `yaml:"maxRiskPerTrade"`      // Max risk per trade (0.02 = 2%)
	MaxDailyLoss         float64 `yaml:"maxDailyLoss"`         // Max daily loss (0.05 = 5%)
	MaxWeeklyLoss        float64 `yaml:"maxWeeklyLoss"`        // Max weekly loss (0.1 = 10%)
	MaxDrawdown          float64 `yaml:"maxDrawdown"`          // Max total drawdown (0.2 = 20%)
	MaxOpenPositions     int     `yaml:"maxOpenPositions"`     // Max concurrent positions
	MaxLeverage          float64 `yaml:"maxLeverage"`          // Max leverage (1.0 = no leverage)
	MinRiskRewardRatio   float64 `yaml:"minRiskRewardRatio"`   // Minimum R/R ratio
	EnableCircuitBreaker bool    `yaml:"enableCircuitBreaker"` // Enable circuit breaker
	ConsecutiveLossLimit int     `yaml:"consecutiveLossLimit"` // Halt after N losses
	HaltDurationHours    int     `yaml:"haltDurationHours"`    // Circuit breaker halt duration
}

// IndicatorConfig represents indicator configuration
type IndicatorConfig struct {
	RSIPeriod       int     `yaml:"rsiPeriod"`
	RSIOversold     float64 `yaml:"rsiOversold"`
	RSIOverbought   float64 `yaml:"rsiOverbought"`
	MACDFast        int     `yaml:"macdFast"`
	MACDSlow        int     `yaml:"macdSlow"`
	MACDSignal      int     `yaml:"macdSignal"`
	BBPeriod        int     `yaml:"bbPeriod"`
	BBStdDev        float64 `yaml:"bbStdDev"`
	ADXPeriod       int     `yaml:"adxPeriod"`
	ADXThreshold    float64 `yaml:"adxThreshold"`
	ATRPeriod       int     `yaml:"atrPeriod"`
	ATRMultiplierSL float64 `yaml:"atrMultiplierSL"`
	ATRMultiplierTP float64 `yaml:"atrMultiplierTP"`
}

// StrategiesConfig represents strategies configuration
type StrategiesConfig struct {
	Enabled []string `yaml:"enabled"` // List of enabled strategy names
}

// DatabaseConfig represents database configuration (SQLite - deprecated, use Postgres)
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// PostgresConfig represents PostgreSQL configuration
type PostgresConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	SSLMode         string        `yaml:"sslmode"`
	MaxConns        int           `yaml:"maxConns"`
	MaxIdle         int           `yaml:"maxIdle"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	JWTSecret          string        `yaml:"jwtSecret"`
	TokenExpiry        time.Duration `yaml:"tokenExpiry"`
	RefreshTokenExpiry time.Duration `yaml:"refreshTokenExpiry"`
}

// DataServiceConfig represents data service configuration
type DataServiceConfig struct {
	CircularQueueSize int           `yaml:"circularQueueSize"`
	CacheExpiry       time.Duration `yaml:"cacheExpiry"`
}

// APIConfig represents API server configuration
type APIConfig struct {
	Port        string   `yaml:"port"`
	CORSOrigins []string `yaml:"corsOrigins"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults for any missing values
	applyDefaults(&cfg)

	return &cfg, nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	cfg := &Config{}
	applyDefaults(cfg)
	return cfg
}

// applyDefaults applies default values to missing config fields
func applyDefaults(cfg *Config) {
	// Trading defaults
	if cfg.Trading.Mode == "" {
		cfg.Trading.Mode = "paper"
	}
	if cfg.Trading.Symbol == "" {
		cfg.Trading.Symbol = "ETHUSDT"
	}
	if len(cfg.Trading.Timeframes) == 0 {
		cfg.Trading.Timeframes = []string{"1m", "5m", "15m", "1h", "4h", "1d"}
	}
	if cfg.Trading.PrimaryTimeframe == "" {
		cfg.Trading.PrimaryTimeframe = "1m" // Use 1m for faster signal generation
	}
	if cfg.Trading.InitialBalance == 0 {
		cfg.Trading.InitialBalance = 100000
	}
	if cfg.Trading.Commission == 0 {
		cfg.Trading.Commission = 0.001
	}
	if cfg.Trading.Slippage == 0 {
		cfg.Trading.Slippage = 0.0005
	}

	// Binance defaults - use production for real live data
	// Testnet is explicitly set only via config file

	// Risk defaults
	if cfg.Risk.MaxPositionSize == 0 {
		cfg.Risk.MaxPositionSize = 0.10
	}
	if cfg.Risk.MaxRiskPerTrade == 0 {
		cfg.Risk.MaxRiskPerTrade = 0.02
	}
	if cfg.Risk.MaxDailyLoss == 0 {
		cfg.Risk.MaxDailyLoss = 0.05
	}
	if cfg.Risk.MaxWeeklyLoss == 0 {
		cfg.Risk.MaxWeeklyLoss = 0.10
	}
	if cfg.Risk.MaxDrawdown == 0 {
		cfg.Risk.MaxDrawdown = 0.20
	}
	if cfg.Risk.MaxOpenPositions == 0 {
		cfg.Risk.MaxOpenPositions = 5
	}
	if cfg.Risk.MaxLeverage == 0 {
		cfg.Risk.MaxLeverage = 1.0
	}
	if cfg.Risk.MinRiskRewardRatio == 0 {
		cfg.Risk.MinRiskRewardRatio = 1.5
	}
	if cfg.Risk.ConsecutiveLossLimit == 0 {
		cfg.Risk.ConsecutiveLossLimit = 5
	}
	if cfg.Risk.HaltDurationHours == 0 {
		cfg.Risk.HaltDurationHours = 24
	}

	// Indicator defaults
	if cfg.Indicators.RSIPeriod == 0 {
		cfg.Indicators.RSIPeriod = 14
	}
	if cfg.Indicators.RSIOversold == 0 {
		cfg.Indicators.RSIOversold = 30
	}
	if cfg.Indicators.RSIOverbought == 0 {
		cfg.Indicators.RSIOverbought = 70
	}
	if cfg.Indicators.MACDFast == 0 {
		cfg.Indicators.MACDFast = 12
	}
	if cfg.Indicators.MACDSlow == 0 {
		cfg.Indicators.MACDSlow = 26
	}
	if cfg.Indicators.MACDSignal == 0 {
		cfg.Indicators.MACDSignal = 9
	}
	if cfg.Indicators.BBPeriod == 0 {
		cfg.Indicators.BBPeriod = 20
	}
	if cfg.Indicators.BBStdDev == 0 {
		cfg.Indicators.BBStdDev = 2.0
	}
	if cfg.Indicators.ADXPeriod == 0 {
		cfg.Indicators.ADXPeriod = 14
	}
	if cfg.Indicators.ADXThreshold == 0 {
		cfg.Indicators.ADXThreshold = 25
	}
	if cfg.Indicators.ATRPeriod == 0 {
		cfg.Indicators.ATRPeriod = 14
	}
	if cfg.Indicators.ATRMultiplierSL == 0 {
		cfg.Indicators.ATRMultiplierSL = 2.0
	}
	if cfg.Indicators.ATRMultiplierTP == 0 {
		cfg.Indicators.ATRMultiplierTP = 3.0
	}

	// Strategies defaults
	if len(cfg.Strategies.Enabled) == 0 {
		cfg.Strategies.Enabled = []string{
			"TrendFollowing",
			"MeanReversion",
			"Breakout",
			"Volatility",
			"StatArb",
		}
	}

	// Database defaults (SQLite - deprecated)
	if cfg.Database.Path == "" {
		cfg.Database.Path = "data/trading.db"
	}

	// PostgreSQL defaults
	if cfg.Postgres.Host == "" {
		cfg.Postgres.Host = "localhost"
	}
	if cfg.Postgres.Port == 0 {
		cfg.Postgres.Port = 5432
	}
	if cfg.Postgres.User == "" {
		cfg.Postgres.User = "postgres"
	}
	if cfg.Postgres.Password == "" {
		cfg.Postgres.Password = "postgres"
	}
	if cfg.Postgres.DBName == "" {
		cfg.Postgres.DBName = "eth_trading"
	}
	if cfg.Postgres.SSLMode == "" {
		cfg.Postgres.SSLMode = "disable"
	}
	if cfg.Postgres.MaxConns == 0 {
		cfg.Postgres.MaxConns = 25
	}
	if cfg.Postgres.MaxIdle == 0 {
		cfg.Postgres.MaxIdle = 5
	}
	if cfg.Postgres.ConnMaxLifetime == 0 {
		cfg.Postgres.ConnMaxLifetime = 5 * time.Minute
	}

	// Auth defaults
	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = "change-me-in-production-to-a-secure-random-string"
	}
	if cfg.Auth.TokenExpiry == 0 {
		cfg.Auth.TokenExpiry = 15 * time.Minute
	}
	if cfg.Auth.RefreshTokenExpiry == 0 {
		cfg.Auth.RefreshTokenExpiry = 7 * 24 * time.Hour
	}

	// DataService defaults
	if cfg.DataService.CircularQueueSize == 0 {
		cfg.DataService.CircularQueueSize = 1000
	}
	if cfg.DataService.CacheExpiry == 0 {
		cfg.DataService.CacheExpiry = 5 * time.Minute
	}

	// API defaults
	if cfg.API.Port == "" {
		cfg.API.Port = ":8080"
	}
	if len(cfg.API.CORSOrigins) == 0 {
		cfg.API.CORSOrigins = []string{"*"}
	}
}

// Save saves configuration to a YAML file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
