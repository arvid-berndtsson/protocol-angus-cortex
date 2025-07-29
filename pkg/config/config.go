package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Capture CaptureConfig `mapstructure:"capture"`
	Cortex  CortexConfig  `mapstructure:"cortex"`
}

// ServerConfig holds API and metrics server configuration
type ServerConfig struct {
	APIPort     int `mapstructure:"api_port"`
	MetricsPort int `mapstructure:"metrics_port"`
}

// CaptureConfig holds packet capture configuration
type CaptureConfig struct {
	Interface  string `mapstructure:"interface"`
	BPFFilter  string `mapstructure:"bpf_filter"`
	BufferSize int    `mapstructure:"buffer_size"`
}

// CortexConfig holds neural network model configuration
type CortexConfig struct {
	ModelPath          string  `mapstructure:"model_path"`
	DetectionThreshold float64 `mapstructure:"detection_threshold"`
	BatchSize          int     `mapstructure:"batch_size"`
	InferenceTimeout   int     `mapstructure:"inference_timeout"`
}

// Load reads configuration from the specified file
func Load(configPath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults
	if config.Server.APIPort == 0 {
		config.Server.APIPort = 8080
	}
	if config.Server.MetricsPort == 0 {
		config.Server.MetricsPort = 9090
	}
	if config.Capture.BufferSize == 0 {
		config.Capture.BufferSize = 1024 * 1024 // 1MB
	}
	if config.Cortex.DetectionThreshold == 0 {
		config.Cortex.DetectionThreshold = 0.85
	}
	if config.Cortex.BatchSize == 0 {
		config.Cortex.BatchSize = 32
	}
	if config.Cortex.InferenceTimeout == 0 {
		config.Cortex.InferenceTimeout = 1000 // milliseconds
	}

	return &config, nil
}
