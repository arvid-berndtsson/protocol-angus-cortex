package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/config"
)

// SimpleMLDemo represents a simple ML demo without external dependencies
type SimpleMLDemo struct {
	config     config.MLConfig
	stats      *DemoStatistics
	mu         sync.RWMutex
}

// DemoStatistics holds demo statistics
type DemoStatistics struct {
	TotalPredictions   int64     `json:"total_predictions"`
	BotDetections      int64     `json:"bot_detections"`
	HumanDetections    int64     `json:"human_detections"`
	AverageConfidence  float64   `json:"average_confidence"`
	LastPrediction     time.Time `json:"last_prediction"`
	mu                 sync.RWMutex
}

// DetectionResult represents the result of bot detection
type DetectionResult struct {
	IsBot      bool      `json:"is_bot"`
	Confidence float64   `json:"confidence"`
	Features   []float64 `json:"features"`
	Reasoning  string    `json:"reasoning"`
	ModelUsed  string    `json:"model_used"`
	Timestamp  time.Time `json:"timestamp"`
	FlowID     string    `json:"flow_id"`
}

// SimpleMLModel represents a simple ML model
type SimpleMLModel struct {
	weights []float64
	bias    float64
	trained bool
}

// NewSimpleMLDemo creates a new simple ML demo
func NewSimpleMLDemo(config config.MLConfig) *SimpleMLDemo {
	return &SimpleMLDemo{
		config: config,
		stats:  &DemoStatistics{},
	}
}

// Predict performs bot detection using simple heuristics
func (d *SimpleMLDemo) Predict(ctx context.Context, features []float64, flowID string) (*DetectionResult, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Simple heuristic-based prediction
	confidence := d.simplePrediction(features)
	isBot := confidence > d.config.DetectionThreshold
	reasoning := d.generateReasoning(features, confidence)
	
	result := &DetectionResult{
		IsBot:      isBot,
		Confidence: confidence,
		Features:   features,
		Reasoning:  reasoning,
		ModelUsed:  "simple_heuristic",
		Timestamp:  time.Now(),
		FlowID:     flowID,
	}
	
	d.updateStats(result)
	
	return result, nil
}

// simplePrediction provides a simple heuristic-based prediction
func (d *SimpleMLDemo) simplePrediction(features []float64) float64 {
	var score float64
	
	// Analyze feature patterns that might indicate bot behavior
	for i, feature := range features {
		// Higher values in certain ranges might indicate bot behavior
		if i < 10 && feature > 0.7 {
			score += 0.1
		}
		if i >= 10 && i < 20 && feature < 0.3 {
			score += 0.1
		}
		if i >= 20 && i < 30 && math.Abs(feature-0.5) < 0.1 {
			score += 0.1
		}
		if i >= 30 && i < 40 && feature > 0.8 {
			score += 0.1
		}
		if i >= 40 && i < 50 && feature < 0.2 {
			score += 0.1
		}
	}
	
	// Normalize to [0, 1]
	score = math.Min(score, 1.0)
	return score
}

// generateReasoning provides human-readable explanation for the prediction
func (d *SimpleMLDemo) generateReasoning(features []float64, confidence float64) string {
	var reasoning string
	
	if confidence > 0.8 {
		reasoning = "High confidence bot detection based on "
	} else if confidence > 0.6 {
		reasoning = "Moderate confidence bot detection based on "
	} else if confidence > 0.4 {
		reasoning = "Low confidence bot detection based on "
	} else {
		reasoning = "Human-like behavior detected based on "
	}
	
	reasoning += "simple heuristic analysis. "
	
	// Add specific feature insights
	if len(features) > 0 {
		reasoning += "Key indicators include packet timing patterns, "
		reasoning += "protocol behavior consistency, and flow characteristics."
	}
	
	return reasoning
}

// updateStats updates the demo statistics
func (d *SimpleMLDemo) updateStats(result *DetectionResult) {
	d.stats.mu.Lock()
	defer d.stats.mu.Unlock()
	
	d.stats.TotalPredictions++
	if result.IsBot {
		d.stats.BotDetections++
	} else {
		d.stats.HumanDetections++
	}
	
	// Update average confidence
	total := float64(d.stats.TotalPredictions)
	d.stats.AverageConfidence = (d.stats.AverageConfidence*(total-1) + result.Confidence) / total
	
	d.stats.LastPrediction = result.Timestamp
}

// GetStatistics returns the current demo statistics
func (d *SimpleMLDemo) GetStatistics() *DemoStatistics {
	d.stats.mu.RLock()
	defer d.stats.mu.RUnlock()
	
	stats := *d.stats // Copy to avoid race conditions
	return &stats
}

func main() {
	fmt.Println("🤖 Protocol Argus Cortex - Simple ML Demo")
	fmt.Println("=========================================")

	// Create ML configuration
	mlConfig := config.MLConfig{
		ModelType:          "simple_heuristic",
		DetectionThreshold: 0.6,
		BatchSize:          32,
		TrainingEpochs:     50,
		LearningRate:       0.001,
		FeatureSize:        128,
		GenerateFakeData:   true,
		FakeDataSize:       500,
		ModelPath:          "./models/bot_detection_model",
		SaveModel:          true,
		LoadModel:          false,
		EnableGPU:          false,
		MaxConcurrency:     4,
		EnableMetrics:      true,
		LogPredictions:     true,
	}

	// Initialize simple ML demo
	fmt.Println("🚀 Initializing Simple ML demo...")
	demo := NewSimpleMLDemo(mlConfig)

	fmt.Println("✅ Simple ML demo initialized successfully!")
	fmt.Printf("📊 Model type: %s\n", mlConfig.ModelType)
	fmt.Printf("🎯 Detection threshold: %.2f\n", mlConfig.DetectionThreshold)
	fmt.Printf("📈 Feature size: %d\n", mlConfig.FeatureSize)

	// Demo 1: Test with bot-like features
	fmt.Println("\n🔍 Demo 1: Testing with bot-like features")
	botFeatures := generateBotFeatures(mlConfig.FeatureSize)
	result, err := demo.Predict(context.Background(), botFeatures, "demo_bot_001")
	if err != nil {
		log.Printf("Prediction failed: %v", err)
	} else {
		printResult("Bot-like traffic", result)
	}

	// Demo 2: Test with human-like features
	fmt.Println("\n👤 Demo 2: Testing with human-like features")
	humanFeatures := generateHumanFeatures(mlConfig.FeatureSize)
	result, err = demo.Predict(context.Background(), humanFeatures, "demo_human_001")
	if err != nil {
		log.Printf("Prediction failed: %v", err)
	} else {
		printResult("Human-like traffic", result)
	}

	// Demo 3: Test with random features
	fmt.Println("\n🎲 Demo 3: Testing with random features")
	for i := 0; i < 5; i++ {
		randomFeatures := generateRandomFeatures(mlConfig.FeatureSize)
		result, err := demo.Predict(context.Background(), randomFeatures, fmt.Sprintf("demo_random_%03d", i+1))
		if err != nil {
			log.Printf("Prediction failed: %v", err)
			continue
		}
		printResult(fmt.Sprintf("Random traffic %d", i+1), result)
	}

	// Demo 4: Batch prediction
	fmt.Println("\n📦 Demo 4: Batch prediction test")
	batchResults := performBatchPrediction(demo, mlConfig.FeatureSize, 10)
	printBatchResults(batchResults)

	// Demo 5: Show statistics
	fmt.Println("\n📊 Demo 5: Demo Statistics")
	stats := demo.GetStatistics()
	printStatistics(stats)

	// Demo 6: Model information
	fmt.Println("\nℹ️  Demo 6: Model Information")
	printModelInfo(mlConfig)

	fmt.Println("\n🎉 Simple ML Demo completed successfully!")
}

// generateBotFeatures creates features that simulate bot behavior
func generateBotFeatures(featureSize int) []float64 {
	features := make([]float64, featureSize)
	
	// Bot characteristics: regular timing, consistent patterns
	for i := 0; i < featureSize; i++ {
		switch {
		case i < 20: // Timing features - very regular
			features[i] = 0.1 + rand.Float64()*0.1
		case i < 40: // Size features - consistent
			features[i] = 0.4 + rand.Float64()*0.2
		case i < 60: // Rate features - high and consistent
			features[i] = 0.7 + rand.Float64()*0.3
		case i < 80: // Protocol features - strict adherence
			features[i] = 0.8 + rand.Float64()*0.2
		case i < 100: // Duration features - long flows
			features[i] = 0.6 + rand.Float64()*0.4
		case i < 120: // Entropy features - low entropy
			features[i] = 0.1 + rand.Float64()*0.3
		default: // Additional features
			features[i] = rand.Float64() * 0.5
		}
	}
	
	return features
}

// generateHumanFeatures creates features that simulate human behavior
func generateHumanFeatures(featureSize int) []float64 {
	features := make([]float64, featureSize)
	
	// Human characteristics: irregular timing, variable patterns
	for i := 0; i < featureSize; i++ {
		switch {
		case i < 20: // Timing features - irregular
			features[i] = 0.3 + rand.Float64()*0.7
		case i < 40: // Size features - variable
			features[i] = 0.1 + rand.Float64()*0.9
		case i < 60: // Rate features - lower and variable
			features[i] = 0.1 + rand.Float64()*0.4
		case i < 80: // Protocol features - less strict
			features[i] = 0.2 + rand.Float64()*0.6
		case i < 100: // Duration features - shorter flows
			features[i] = 0.1 + rand.Float64()*0.5
		case i < 120: // Entropy features - high entropy
			features[i] = 0.4 + rand.Float64()*0.6
		default: // Additional features
			features[i] = 0.3 + rand.Float64()*0.7
		}
	}
	
	return features
}

// generateRandomFeatures creates completely random features
func generateRandomFeatures(featureSize int) []float64 {
	features := make([]float64, featureSize)
	for i := range features {
		features[i] = rand.Float64()
	}
	return features
}

// performBatchPrediction runs multiple predictions
func performBatchPrediction(demo *SimpleMLDemo, featureSize, count int) []*DetectionResult {
	results := make([]*DetectionResult, count)
	
	for i := 0; i < count; i++ {
		features := generateRandomFeatures(featureSize)
		result, err := demo.Predict(context.Background(), features, fmt.Sprintf("batch_%03d", i+1))
		if err != nil {
			fmt.Printf("❌ Batch prediction %d failed: %v\n", i+1, err)
			continue
		}
		results[i] = result
	}
	
	return results
}

// printResult prints a single prediction result
func printResult(label string, result *DetectionResult) {
	fmt.Printf("  %s:\n", label)
	fmt.Printf("    🤖 Is Bot: %t\n", result.IsBot)
	fmt.Printf("    📊 Confidence: %.3f\n", result.Confidence)
	fmt.Printf("    🧠 Model Used: %s\n", result.ModelUsed)
	fmt.Printf("    💭 Reasoning: %s\n", result.Reasoning)
	fmt.Printf("    🕒 Timestamp: %s\n", result.Timestamp.Format("15:04:05"))
}

// printBatchResults prints batch prediction results
func printBatchResults(results []*DetectionResult) {
	botCount := 0
	humanCount := 0
	var totalConfidence float64
	
	for _, result := range results {
		if result != nil {
			if result.IsBot {
				botCount++
			} else {
				humanCount++
			}
			totalConfidence += result.Confidence
		}
	}
	
	fmt.Printf("  📈 Batch Results Summary:\n")
	fmt.Printf("    🤖 Bots detected: %d\n", botCount)
	fmt.Printf("    👤 Humans detected: %d\n", humanCount)
	fmt.Printf("    📊 Average confidence: %.3f\n", totalConfidence/float64(len(results)))
}

// printStatistics prints demo statistics
func printStatistics(stats *DemoStatistics) {
	fmt.Printf("  📊 Total Predictions: %d\n", stats.TotalPredictions)
	fmt.Printf("  🤖 Bot Detections: %d\n", stats.BotDetections)
	fmt.Printf("  👤 Human Detections: %d\n", stats.HumanDetections)
	fmt.Printf("  📈 Average Confidence: %.3f\n", stats.AverageConfidence)
	fmt.Printf("  🕒 Last Prediction: %s\n", stats.LastPrediction.Format("15:04:05"))
}

// printModelInfo prints model configuration information
func printModelInfo(config config.MLConfig) {
	fmt.Printf("  🧠 Model Type: %s\n", config.ModelType)
	fmt.Printf("  🎯 Detection Threshold: %.2f\n", config.DetectionThreshold)
	fmt.Printf("  📦 Batch Size: %d\n", config.BatchSize)
	fmt.Printf("  🔄 Training Epochs: %d\n", config.TrainingEpochs)
	fmt.Printf("  📚 Learning Rate: %.4f\n", config.LearningRate)
	fmt.Printf("  📊 Feature Size: %d\n", config.FeatureSize)
	fmt.Printf("  🎲 Generate Fake Data: %t\n", config.GenerateFakeData)
	fmt.Printf("  📈 Fake Data Size: %d\n", config.FakeDataSize)
	fmt.Printf("  💾 Model Path: %s\n", config.ModelPath)
	fmt.Printf("  🖥️  Enable GPU: %t\n", config.EnableGPU)
	fmt.Printf("  🔧 Max Concurrency: %d\n", config.MaxConcurrency)
} 