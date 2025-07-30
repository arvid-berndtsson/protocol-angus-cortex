package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// MLConfig holds configuration for the machine learning engine
type MLConfig struct {
	// Model selection
	ModelType string `mapstructure:"model_type" yaml:"model_type"`

	// Detection parameters
	DetectionThreshold float64 `mapstructure:"detection_threshold" yaml:"detection_threshold"`

	// Training parameters
	BatchSize      int     `mapstructure:"batch_size" yaml:"batch_size"`
	TrainingEpochs int     `mapstructure:"training_epochs" yaml:"training_epochs"`
	LearningRate   float64 `mapstructure:"learning_rate" yaml:"learning_rate"`
	FeatureSize    int     `mapstructure:"feature_size" yaml:"feature_size"`

	// Data generation
	GenerateFakeData bool `mapstructure:"generate_fake_data" yaml:"generate_fake_data"`
	FakeDataSize     int  `mapstructure:"fake_data_size" yaml:"fake_data_size"`

	// Model persistence
	ModelPath string `mapstructure:"model_path" yaml:"model_path"`
	SaveModel bool   `mapstructure:"save_model" yaml:"save_model"`
	LoadModel bool   `mapstructure:"load_model" yaml:"load_model"`

	// Performance settings
	EnableGPU      bool `mapstructure:"enable_gpu" yaml:"enable_gpu"`
	MaxConcurrency int  `mapstructure:"max_concurrency" yaml:"max_concurrency"`

	// Monitoring
	EnableMetrics  bool `mapstructure:"enable_metrics" yaml:"enable_metrics"`
	LogPredictions bool `mapstructure:"log_predictions" yaml:"log_predictions"`
}

// DefaultMLConfig returns default ML configuration
func DefaultMLConfig() MLConfig {
	return MLConfig{
		ModelType:          "ensemble",
		DetectionThreshold: 0.6,
		BatchSize:          32,
		TrainingEpochs:     100,
		LearningRate:       0.001,
		FeatureSize:        128,
		GenerateFakeData:   true,
		FakeDataSize:       1000,
		ModelPath:          "./models/bot_detection_model",
		SaveModel:          true,
		LoadModel:          false,
		EnableGPU:          false,
		MaxConcurrency:     4,
		EnableMetrics:      true,
		LogPredictions:     false,
	}
}

// LoadMLConfig loads ML configuration from viper
func LoadMLConfig(v *viper.Viper) MLConfig {
	config := DefaultMLConfig()

	// Load from viper if available
	if v != nil {
		if err := v.UnmarshalKey("ml", &config); err != nil {
			// Use defaults if unmarshaling fails
			return config
		}
	}

	return config
}

// ValidateMLConfig validates ML configuration
func ValidateMLConfig(config MLConfig) error {
	// Validate model type
	validModels := map[string]bool{
		"neural_network": true,
		"random_forest":  true,
		"knn":            true,
		"svm":            true,
		"ensemble":       true,
	}

	if !validModels[config.ModelType] {
		return fmt.Errorf("invalid model type: %s", config.ModelType)
	}

	// Validate thresholds
	if config.DetectionThreshold < 0 || config.DetectionThreshold > 1 {
		return fmt.Errorf("detection threshold must be between 0 and 1")
	}

	if config.LearningRate <= 0 {
		return fmt.Errorf("learning rate must be positive")
	}

	// Validate sizes
	if config.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}

	if config.FeatureSize <= 0 {
		return fmt.Errorf("feature size must be positive")
	}

	if config.FakeDataSize <= 0 {
		return fmt.Errorf("fake data size must be positive")
	}

	if config.TrainingEpochs <= 0 {
		return fmt.Errorf("training epochs must be positive")
	}

	if config.MaxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be positive")
	}

	return nil
}
