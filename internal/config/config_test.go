package config

import (
	"os"
	"testing"

	"github.com/restartfu/gophig"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ListenAddr == "" {
		t.Error("Default listen address should not be empty")
	}

	if cfg.ListenAddr != ":8080" {
		t.Errorf("Expected default listen address ':8080', got '%s'", cfg.ListenAddr)
	}

	if len(cfg.APIKeys) == 0 {
		t.Error("Default config should have at least one API key")
	}

	if cfg.APIKeys[0] != "default" {
		t.Errorf("Expected default API key 'default', got '%s'", cfg.APIKeys[0])
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		isValid bool
	}{
		{
			name: "Valid config",
			config: Config{
				ListenAddr: ":8080",
				APIKeys:    []string{"valid-key"},
			},
			isValid: true,
		},
		{
			name: "Valid config with port only",
			config: Config{
				ListenAddr: ":9090",
				APIKeys:    []string{"key1", "key2"},
			},
			isValid: true,
		},
		{
			name: "Valid config with host and port",
			config: Config{
				ListenAddr: "localhost:8080",
				APIKeys:    []string{"secret"},
			},
			isValid: true,
		},
		{
			name: "Empty listen address",
			config: Config{
				ListenAddr: "",
				APIKeys:    []string{"key"},
			},
			isValid: false,
		},
		{
			name: "No API keys",
			config: Config{
				ListenAddr: ":8080",
				APIKeys:    []string{},
			},
			isValid: false,
		},
		{
			name: "Nil API keys",
			config: Config{
				ListenAddr: ":8080",
				APIKeys:    nil,
			},
			isValid: false,
		},
		{
			name: "Empty API key",
			config: Config{
				ListenAddr: ":8080",
				APIKeys:    []string{""},
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidConfig(tt.config)
			if valid != tt.isValid {
				t.Errorf("Expected config validity %v, got %v", tt.isValid, valid)
			}
		})
	}
}

func TestConfigSerialization(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config_test_*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Test config
	originalConfig := Config{
		ListenAddr: ":9999",
		APIKeys:    []string{"test-key-1", "test-key-2"},
	}

	// Save config
	g := gophig.NewGophig[Config](tmpFile.Name(), gophig.TOMLMarshaler{}, os.ModePerm)
	err = g.SaveConf(originalConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedConfig, err := g.LoadConf()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Compare configs
	if loadedConfig.ListenAddr != originalConfig.ListenAddr {
		t.Errorf("Expected listen address '%s', got '%s'", originalConfig.ListenAddr, loadedConfig.ListenAddr)
	}

	if len(loadedConfig.APIKeys) != len(originalConfig.APIKeys) {
		t.Errorf("Expected %d API keys, got %d", len(originalConfig.APIKeys), len(loadedConfig.APIKeys))
	}

	for i, key := range originalConfig.APIKeys {
		if i >= len(loadedConfig.APIKeys) || loadedConfig.APIKeys[i] != key {
			t.Errorf("Expected API key '%s' at index %d, got '%s'", key, i, loadedConfig.APIKeys[i])
		}
	}
}

func TestConfigWithMultipleAPIKeys(t *testing.T) {
	cfg := Config{
		ListenAddr: ":8080",
		APIKeys:    []string{"key1", "key2", "key3", "very-long-api-key-with-special-chars-123!@#"},
	}

	if !isValidConfig(cfg) {
		t.Error("Config with multiple API keys should be valid")
	}

	// Test that all keys are preserved
	for i, expectedKey := range []string{"key1", "key2", "key3", "very-long-api-key-with-special-chars-123!@#"} {
		if cfg.APIKeys[i] != expectedKey {
			t.Errorf("Expected API key '%s' at index %d, got '%s'", expectedKey, i, cfg.APIKeys[i])
		}
	}
}

func TestConfigListenAddressFormats(t *testing.T) {
	validAddresses := []string{
		":8080",
		":0",
		":65535",
		"localhost:8080",
		"127.0.0.1:8080",
		"0.0.0.0:8080",
		"[::1]:8080",
		"example.com:8080",
	}

	for _, addr := range validAddresses {
		cfg := Config{
			ListenAddr: addr,
			APIKeys:    []string{"test-key"},
		}

		if !isValidConfig(cfg) {
			t.Errorf("Address '%s' should be valid", addr)
		}
	}
}

// Helper function to validate config
func isValidConfig(cfg Config) bool {
	if cfg.ListenAddr == "" {
		return false
	}

	if len(cfg.APIKeys) == 0 {
		return false
	}

	for _, key := range cfg.APIKeys {
		if key == "" {
			return false
		}
	}

	return true
}

func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

func BenchmarkConfigValidation(b *testing.B) {
	cfg := DefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = isValidConfig(cfg)
	}
}
