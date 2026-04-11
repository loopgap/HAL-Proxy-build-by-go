package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Address != ":8080" {
		t.Errorf("Expected default address :8080, got %s", cfg.Server.Address)
	}
	if cfg.Database.Path == "" {
		t.Error("Database path should have default value")
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Expected default log level info, got %s", cfg.Log.Level)
	}
}

func TestConfig_EnvOverrides(t *testing.T) {
	os.Setenv("BRIDGEOS_ADDR", ":9090")
	os.Setenv("BRIDGEOS_DB", "custom.db")
	defer os.Unsetenv("BRIDGEOS_ADDR")
	defer os.Unsetenv("BRIDGEOS_DB")
	cfg := DefaultConfig()
	if cfg.Server.Address != ":9090" {
		t.Errorf("Expected :9090, got %s", cfg.Server.Address)
	}
	if cfg.Database.Path != "custom.db" {
		t.Errorf("Expected custom.db, got %s", cfg.Database.Path)
	}
}

func TestConfig_Validate(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Address: ":8080", MaxBodySize: 1 << 20, RequestTimeout: 30},
		Database: DatabaseConfig{Path: "test.db"},
		Log: LogConfig{Level: "info", Format: "json"},
	}
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

func TestConfig_ValidateInvalidAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Address: "", MaxBodySize: 1 << 20, RequestTimeout: 30},
		Database: DatabaseConfig{Path: "test.db"},
		Log: LogConfig{Level: "info", Format: "json"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for empty server address")
	}
}

func TestConfig_ValidateInvalidLogLevel(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Address: ":8080", MaxBodySize: 1 << 20, RequestTimeout: 30},
		Database: DatabaseConfig{Path: "test.db"},
		Log: LogConfig{Level: "invalid", Format: "json"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for invalid log level")
	}
}

func TestServerConfig_GetTimeouts(t *testing.T) {
	cfg := &ServerConfig{
		ReadTimeout:    10,
		WriteTimeout:   20,
		IdleTimeout:    30,
		RequestTimeout: 40,
	}
	
	if cfg.GetReadTimeout() != 10*1e9 {
		t.Errorf("Expected 10s, got %v", cfg.GetReadTimeout())
	}
	if cfg.GetWriteTimeout() != 20*1e9 {
		t.Errorf("Expected 20s, got %v", cfg.GetWriteTimeout())
	}
	if cfg.GetIdleTimeout() != 30*1e9 {
		t.Errorf("Expected 30s, got %v", cfg.GetIdleTimeout())
	}
	if cfg.GetRequestTimeout() != 40*1e9 {
		t.Errorf("Expected 40s, got %v", cfg.GetRequestTimeout())
	}
}
