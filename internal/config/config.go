package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	App       AppConfig       `json:"app"`
	Log       LogConfig       `json:"log"`
	Auth      AuthConfig      `json:"auth"`
	RateLimit RateLimitConfig `json:"rate_limit"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Address        string `json:"address"`
	ReadTimeout    int    `json:"read_timeout"`
	WriteTimeout   int    `json:"write_timeout"`
	IdleTimeout    int    `json:"idle_timeout"`
	MaxBodySize    int64  `json:"max_body_size"`
	RequestTimeout int    `json:"request_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path         string `json:"path"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Environment  string `json:"environment"`
	ArtifactsDir string `json:"artifacts_dir"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level    string `json:"level"`
	Format   string `json:"format"`
	Output   string `json:"output"`
	FilePath string `json:"file_path"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret      string            `json:"jwt_secret"`
	JWTExpiryHours int               `json:"jwt_expiry_hours"`
	JWTIssuer      string            `json:"jwt_issuer"`
	APIKeys        map[string]string `json:"api_keys"`
	TrustedProxies []string          `json:"trusted_proxies"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
	BurstSize         int  `json:"burst_size"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address:        getEnv("HAL_PROXY_ADDR", ":8080"),
			ReadTimeout:    30,
			WriteTimeout:   30,
			IdleTimeout:    120,
			MaxBodySize:    1 << 20,
			RequestTimeout: 30,
		},
		Database: DatabaseConfig{
			Path:         getEnv("HAL_PROXY_DB", "hal-proxy.db"),
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		App: AppConfig{
			Name:         "HAL-Proxy",
			Version:      "1.0.0",
			Environment:  getEnv("HAL_PROXY_ENV", "development"),
			ArtifactsDir: getEnv("HAL_PROXY_ARTIFACTS", "artifacts"),
		},
		Log: LogConfig{
			Level:  getEnv("HAL_PROXY_LOG_LEVEL", "info"),
			Format: "json",
			Output: "stdout",
		},
		Auth: AuthConfig{
			JWTSecret:      getEnv("HAL_PROXY_JWT_SECRET", "UNCONFIGURED-REQUIRED-SET-HAL_PROXY_JWT_SECRET-32CHARS"),
			JWTExpiryHours: 24,
			JWTIssuer:      getEnv("HAL_PROXY_JWT_ISSUER", "hal-proxy"),
			APIKeys:        parseAPIKeys(getEnv("HAL_PROXY_API_KEYS", "")),
		},
		RateLimit: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 60,
			BurstSize:         10,
		},
	}
}

func parseAPIKeys(env string) map[string]string {
	keys := make(map[string]string)
	if env == "" {
		return keys
	}
	pairs := strings.Split(env, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			keys[kv[0]] = kv[1]
		}
	}
	return keys
}

// Load loads configuration from a simple key=value file
func Load(path string) (*Config, error) {
	config := DefaultConfig()

	if path == "" {
		return config, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse simple key=value format
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parseConfigLine(config, line)
	}

	// Override with environment variables
	config.applyEnvOverrides()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// parseConfigLine parses a single config line
func parseConfigLine(config *Config, line string) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Server config
	switch key {
	case "server.address":
		config.Server.Address = value
	case "server.read_timeout":
		if v, err := strconv.Atoi(value); err == nil {
			config.Server.ReadTimeout = v
		}
	case "server.write_timeout":
		if v, err := strconv.Atoi(value); err == nil {
			config.Server.WriteTimeout = v
		}
	case "server.idle_timeout":
		if v, err := strconv.Atoi(value); err == nil {
			config.Server.IdleTimeout = v
		}
	case "server.max_body_size":
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			config.Server.MaxBodySize = v
		}
	case "server.request_timeout":
		if v, err := strconv.Atoi(value); err == nil {
			config.Server.RequestTimeout = v
		}
	// Database config
	case "database.path":
		config.Database.Path = value
	case "database.max_open_conns":
		if v, err := strconv.Atoi(value); err == nil {
			config.Database.MaxOpenConns = v
		}
	case "database.max_idle_conns":
		if v, err := strconv.Atoi(value); err == nil {
			config.Database.MaxIdleConns = v
		}
	// App config
	case "app.name":
		config.App.Name = value
	case "app.version":
		config.App.Version = value
	case "app.environment":
		config.App.Environment = value
	case "app.artifacts_dir":
		config.App.ArtifactsDir = value
	// Log config
	case "log.level":
		config.Log.Level = value
	case "log.format":
		config.Log.Format = value
	case "log.output":
		config.Log.Output = value
	case "log.file_path":
		config.Log.FilePath = value
	// Auth config
	case "auth.jwt_secret":
		config.Auth.JWTSecret = value
	case "auth.jwt_expiry_hours":
		if v, err := strconv.Atoi(value); err == nil {
			config.Auth.JWTExpiryHours = v
		}
	case "auth.jwt_issuer":
		config.Auth.JWTIssuer = value
	// RateLimit config
	case "rate_limit.enabled":
		if v, err := strconv.ParseBool(value); err == nil {
			config.RateLimit.Enabled = v
		}
	case "rate_limit.requests_per_minute":
		if v, err := strconv.Atoi(value); err == nil {
			config.RateLimit.RequestsPerMinute = v
		}
	case "rate_limit.burst_size":
		if v, err := strconv.Atoi(value); err == nil {
			config.RateLimit.BurstSize = v
		}
	}
}

// applyEnvOverrides applies environment variable overrides
func (c *Config) applyEnvOverrides() {
	if addr := os.Getenv("HAL_PROXY_ADDR"); addr != "" {
		c.Server.Address = addr
	}
	if db := os.Getenv("HAL_PROXY_DB"); db != "" {
		c.Database.Path = db
	}
	if artifacts := os.Getenv("HAL_PROXY_ARTIFACTS"); artifacts != "" {
		c.App.ArtifactsDir = artifacts
	}
	if env := os.Getenv("HAL_PROXY_ENV"); env != "" {
		c.App.Environment = env
	}
	if logLevel := os.Getenv("HAL_PROXY_LOG_LEVEL"); logLevel != "" {
		c.Log.Level = logLevel
	}
	if jwtSecret := os.Getenv("HAL_PROXY_JWT_SECRET"); jwtSecret != "" {
		c.Auth.JWTSecret = jwtSecret
	}
	if jwtIssuer := os.Getenv("HAL_PROXY_JWT_ISSUER"); jwtIssuer != "" {
		c.Auth.JWTIssuer = jwtIssuer
	}
	if apiKeys := os.Getenv("HAL_PROXY_API_KEYS"); apiKeys != "" {
		c.Auth.APIKeys = parseAPIKeys(apiKeys)
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server address is required")
	}

	if c.Database.Path == "" {
		return fmt.Errorf("database path is required")
	}

	if c.Server.MaxBodySize <= 0 {
		return fmt.Errorf("max body size must be positive")
	}

	if c.Server.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	// JWT secret validation - security critical
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("jwt_secret must be at least 32 characters long")
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.Log.Level] {
		return fmt.Errorf("invalid log level: %s", c.Log.Level)
	}

	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[c.Log.Format] {
		return fmt.Errorf("invalid log format: %s", c.Log.Format)
	}

	return nil
}

// GetReadTimeout returns the read timeout as a duration
func (c *ServerConfig) GetReadTimeout() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout returns the write timeout as a duration
func (c *ServerConfig) GetWriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeout) * time.Second
}

// GetIdleTimeout returns the idle timeout as a duration
func (c *ServerConfig) GetIdleTimeout() time.Duration {
	return time.Duration(c.IdleTimeout) * time.Second
}

// GetRequestTimeout returns the request timeout as a duration
func (c *ServerConfig) GetRequestTimeout() time.Duration {
	return time.Duration(c.RequestTimeout) * time.Second
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
