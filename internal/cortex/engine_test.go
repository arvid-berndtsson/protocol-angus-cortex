package cortex

import (
	"context"
	"testing"
	"time"

	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/config"
)

func TestNewEngine(t *testing.T) {
	cfg := config.CortexConfig{
		ModelPath:          "./test_model.onnx",
		DetectionThreshold: 0.85,
		BatchSize:          32,
		InferenceTimeout:   1000,
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	if engine == nil {
		t.Fatal("Engine should not be nil")
	}

	if engine.config != cfg {
		t.Errorf("Expected config %v, got %v", cfg, engine.config)
	}

	if engine.model == nil {
		t.Fatal("Model should not be nil")
	}

	if !engine.model.loaded {
		t.Error("Model should be loaded")
	}
}

func TestAnalyze(t *testing.T) {
	cfg := config.CortexConfig{
		ModelPath:          "./test_model.onnx",
		DetectionThreshold: 0.85,
		BatchSize:          32,
		InferenceTimeout:   1000,
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Create test features
	features := make([]float64, 128)
	for i := range features {
		features[i] = float64(i) / 128.0
	}

	ctx := context.Background()
	flowID := "test-flow-123"

	result, err := engine.Analyze(ctx, features, flowID)
	if err != nil {
		t.Fatalf("Failed to analyze: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.FlowID != flowID {
		t.Errorf("Expected flow ID %s, got %s", flowID, result.FlowID)
	}

	if len(result.Features) != 128 {
		t.Errorf("Expected 128 features, got %d", len(result.Features))
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", result.Confidence)
	}

	if result.Reasoning == "" {
		t.Error("Reasoning should not be empty")
	}
}

func TestSimulateInference(t *testing.T) {
	engine := &Engine{}

	// Test with different feature sets
	testCases := []struct {
		name     string
		features []float64
	}{
		{
			name:     "normal features",
			features: make([]float64, 128),
		},
		{
			name: "bot-like features",
			features: func() []float64 {
				features := make([]float64, 128)
				features[0] = 1500  // Large packet size
				features[10] = 0.05 // Low timing variance
				features[20] = 0.3  // Low header count
				return features
			}(),
		},
		{
			name: "human-like features",
			features: func() []float64 {
				features := make([]float64, 128)
				features[0] = 800  // Normal packet size
				features[10] = 0.5 // Normal timing variance
				features[20] = 0.8 // Normal header count
				return features
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			confidence, reasoning := engine.simulateInference(tc.features)

			if confidence < 0 || confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
			}

			if reasoning == "" {
				t.Error("Reasoning should not be empty")
			}
		})
	}
}

func TestGetStatistics(t *testing.T) {
	cfg := config.CortexConfig{
		ModelPath:          "./test_model.onnx",
		DetectionThreshold: 0.85,
		BatchSize:          32,
		InferenceTimeout:   1000,
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Perform some analyses
	features := make([]float64, 128)
	for i := range features {
		features[i] = float64(i) / 128.0
	}

	ctx := context.Background()
	engine.Analyze(ctx, features, "flow-1")
	engine.Analyze(ctx, features, "flow-2")
	engine.Analyze(ctx, features, "flow-3")

	stats := engine.GetStatistics()

	if stats.TotalInferences != 3 {
		t.Errorf("Expected 3 total inferences, got %d", stats.TotalInferences)
	}

	if stats.BotDetections < 0 {
		t.Errorf("Bot detections should be non-negative, got %d", stats.BotDetections)
	}

	if stats.HumanDetections < 0 {
		t.Errorf("Human detections should be non-negative, got %d", stats.HumanDetections)
	}

	if stats.AverageConfidence < 0 || stats.AverageConfidence > 1 {
		t.Errorf("Average confidence should be between 0 and 1, got %f", stats.AverageConfidence)
	}

	if stats.LastInference.IsZero() {
		t.Error("Last inference time should not be zero")
	}
}

func TestUpdateStats(t *testing.T) {
	engine := &Engine{
		stats: &Statistics{},
	}

	// Create test results
	results := []*DetectionResult{
		{
			IsBot:      true,
			Confidence: 0.9,
			Timestamp:  time.Now(),
		},
		{
			IsBot:      false,
			Confidence: 0.3,
			Timestamp:  time.Now(),
		},
		{
			IsBot:      true,
			Confidence: 0.8,
			Timestamp:  time.Now(),
		},
	}

	// Update stats with results
	for _, result := range results {
		engine.updateStats(result)
	}

	stats := engine.GetStatistics()

	if stats.TotalInferences != 3 {
		t.Errorf("Expected 3 total inferences, got %d", stats.TotalInferences)
	}

	if stats.BotDetections != 2 {
		t.Errorf("Expected 2 bot detections, got %d", stats.BotDetections)
	}

	if stats.HumanDetections != 1 {
		t.Errorf("Expected 1 human detection, got %d", stats.HumanDetections)
	}

	expectedAvg := (0.9 + 0.3 + 0.8) / 3.0
	if stats.AverageConfidence != expectedAvg {
		t.Errorf("Expected average confidence %f, got %f", expectedAvg, stats.AverageConfidence)
	}
}
