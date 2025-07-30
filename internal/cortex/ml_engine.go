package cortex

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/config"
	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/ml"
)

// MLCortexEngine represents the enhanced cortex engine with real ML capabilities
type MLCortexEngine struct {
	// Core ML engine
	mlEngine *ml.MLEngine
	
	// Configuration
	config config.MLConfig
	
	// Statistics
	stats *MLCortexStatistics
	
	// State management
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// MLCortexStatistics holds enhanced statistics for the ML cortex engine
type MLCortexStatistics struct {
	TotalInferences   int64     `json:"total_inferences"`
	BotDetections     int64     `json:"bot_detections"`
	HumanDetections   int64     `json:"human_detections"`
	AverageConfidence float64   `json:"average_confidence"`
	ModelAccuracy     float64   `json:"model_accuracy"`
	TrainingTime      time.Duration `json:"training_time"`
	LastInference     time.Time `json:"last_inference"`
	ModelType         string    `json:"model_type"`
	mu                sync.RWMutex
}

// NewMLCortexEngine creates a new ML-enhanced cortex engine
func NewMLCortexEngine(cfg config.MLConfig) (*MLCortexEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Convert config.MLConfig to ml.MLConfig
	mlConfig := ml.MLConfig{
		ModelType:          cfg.ModelType,
		DetectionThreshold: cfg.DetectionThreshold,
		BatchSize:          cfg.BatchSize,
		TrainingEpochs:     cfg.TrainingEpochs,
		LearningRate:       cfg.LearningRate,
		FeatureSize:        cfg.FeatureSize,
		GenerateFakeData:   cfg.GenerateFakeData,
		FakeDataSize:       cfg.FakeDataSize,
	}
	
	// Initialize ML engine
	mlEngine, err := ml.NewMLEngine(mlConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize ML engine: %w", err)
	}
	
	engine := &MLCortexEngine{
		mlEngine: mlEngine,
		config:   cfg,
		stats:    &MLCortexStatistics{},
		ctx:      ctx,
		cancel:   cancel,
	}
	
	// Initialize statistics
	engine.stats.ModelType = cfg.ModelType
	
	slog.Info("ML Cortex engine initialized",
		"model_type", cfg.ModelType,
		"threshold", cfg.DetectionThreshold,
		"feature_size", cfg.FeatureSize,
		"fake_data", cfg.GenerateFakeData)
	
	return engine, nil
}

// Analyze performs bot detection analysis using the ML engine
func (e *MLCortexEngine) Analyze(ctx context.Context, features []float64, flowID string) (*DetectionResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// Validate input features
	if len(features) != e.config.FeatureSize {
		return nil, fmt.Errorf("invalid feature vector size: got %d, expected %d",
			len(features), e.config.FeatureSize)
	}
	
	// Perform ML-based prediction
	mlResult, err := e.mlEngine.Predict(ctx, features, flowID)
	if err != nil {
		return nil, fmt.Errorf("ML prediction failed: %w", err)
	}
	
	// Convert ML result to cortex result
	result := &DetectionResult{
		IsBot:      mlResult.IsBot,
		Confidence: mlResult.Confidence,
		Features:   mlResult.Features,
		Reasoning:  mlResult.Reasoning,
		Timestamp:  mlResult.Timestamp,
		FlowID:     mlResult.FlowID,
	}
	
	// Update statistics
	e.updateStats(result)
	
	slog.Debug("ML bot detection analysis completed",
		"flow_id", flowID,
		"is_bot", result.IsBot,
		"confidence", result.Confidence,
		"model_used", mlResult.ModelUsed,
		"reasoning", result.Reasoning)
	
	return result, nil
}

// GetStatistics returns the current ML cortex engine statistics
func (e *MLCortexEngine) GetStatistics() *MLCortexStatistics {
	e.stats.mu.RLock()
	defer e.stats.mu.RUnlock()
	
	// Get ML engine statistics
	mlStats := e.mlEngine.GetStatistics()
	
	// Update our statistics with ML engine data
	e.stats.mu.Lock()
	e.stats.TotalInferences = mlStats.TotalPredictions
	e.stats.BotDetections = mlStats.BotDetections
	e.stats.HumanDetections = mlStats.HumanDetections
	e.stats.AverageConfidence = mlStats.AverageConfidence
	e.stats.ModelAccuracy = mlStats.ModelAccuracy
	e.stats.TrainingTime = mlStats.TrainingTime
	e.stats.LastInference = mlStats.LastPrediction
	e.stats.mu.Unlock()
	
	stats := *e.stats // Copy to avoid race conditions
	return &stats
}

// GetMLStatistics returns the raw ML engine statistics
func (e *MLCortexEngine) GetMLStatistics() *ml.MLStatistics {
	return e.mlEngine.GetStatistics()
}

// RetrainModel retrains the ML model with new data
func (e *MLCortexEngine) RetrainModel(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	slog.Info("Retraining ML model")
	
	// Generate new fake data and retrain
	if err := e.mlEngine.TrainOnFakeData(); err != nil {
		return fmt.Errorf("failed to retrain model: %w", err)
	}
	
	slog.Info("ML model retraining completed")
	return nil
}

// UpdateConfig updates the ML engine configuration
func (e *MLCortexEngine) UpdateConfig(newConfig config.MLConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// Validate new configuration
	if err := config.ValidateMLConfig(newConfig); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	e.config = newConfig
	slog.Info("ML Cortex engine configuration updated", "model_type", newConfig.ModelType)
	
	return nil
}

// GetConfig returns the current configuration
func (e *MLCortexEngine) GetConfig() config.MLConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

// updateStats updates the ML cortex engine statistics
func (e *MLCortexEngine) updateStats(result *DetectionResult) {
	e.stats.mu.Lock()
	defer e.stats.mu.Unlock()
	
	e.stats.TotalInferences++
	e.stats.LastInference = result.Timestamp
	
	if result.IsBot {
		e.stats.BotDetections++
	} else {
		e.stats.HumanDetections++
	}
	
	// Update average confidence
	total := float64(e.stats.TotalInferences)
	e.stats.AverageConfidence = (e.stats.AverageConfidence*(total-1) + result.Confidence) / total
}

// Close cleans up resources
func (e *MLCortexEngine) Close() error {
	e.cancel()
	
	if e.mlEngine != nil {
		return e.mlEngine.Close()
	}
	
	return nil
}

// HealthCheck performs a health check on the ML engine
func (e *MLCortexEngine) HealthCheck() error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if e.mlEngine == nil {
		return fmt.Errorf("ML engine not initialized")
	}
	
	// Perform a simple prediction test
	testFeatures := make([]float64, e.config.FeatureSize)
	for i := range testFeatures {
		testFeatures[i] = 0.5 // Neutral test values
	}
	
	_, err := e.mlEngine.Predict(e.ctx, testFeatures, "health_check")
	return err
}

// GetModelInfo returns information about the current ML model
func (e *MLCortexEngine) GetModelInfo() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	info := map[string]interface{}{
		"model_type":          e.config.ModelType,
		"detection_threshold": e.config.DetectionThreshold,
		"feature_size":        e.config.FeatureSize,
		"batch_size":          e.config.BatchSize,
		"learning_rate":       e.config.LearningRate,
		"training_epochs":     e.config.TrainingEpochs,
		"generate_fake_data":  e.config.GenerateFakeData,
		"fake_data_size":      e.config.FakeDataSize,
		"model_path":          e.config.ModelPath,
		"save_model":          e.config.SaveModel,
		"load_model":          e.config.LoadModel,
		"enable_gpu":          e.config.EnableGPU,
		"max_concurrency":     e.config.MaxConcurrency,
		"enable_metrics":      e.config.EnableMetrics,
		"log_predictions":     e.config.LogPredictions,
	}
	
	return info
} 