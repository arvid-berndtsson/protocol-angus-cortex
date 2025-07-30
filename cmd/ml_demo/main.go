package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/config"
	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/ml"
)

func main() {
	fmt.Println("🤖 Protocol Argus Cortex - ML Demo")
	fmt.Println("==================================")

	// Create ML configuration
	mlConfig := config.MLConfig{
		ModelType:          "ensemble", // Use ensemble of all models
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

	// Initialize ML engine
	fmt.Println("🚀 Initializing ML engine...")
	engine, err := ml.NewMLEngine(ml.MLConfig{
		ModelType:          mlConfig.ModelType,
		DetectionThreshold: mlConfig.DetectionThreshold,
		BatchSize:          mlConfig.BatchSize,
		TrainingEpochs:     mlConfig.TrainingEpochs,
		LearningRate:       mlConfig.LearningRate,
		FeatureSize:        mlConfig.FeatureSize,
		GenerateFakeData:   mlConfig.GenerateFakeData,
		FakeDataSize:       mlConfig.FakeDataSize,
	})
	if err != nil {
		log.Fatalf("Failed to initialize ML engine: %v", err)
	}
	defer engine.Close()

	fmt.Println("✅ ML engine initialized successfully!")
	fmt.Printf("📊 Model type: %s\n", mlConfig.ModelType)
	fmt.Printf("🎯 Detection threshold: %.2f\n", mlConfig.DetectionThreshold)
	fmt.Printf("📈 Feature size: %d\n", mlConfig.FeatureSize)

	// Demo 1: Test with bot-like features
	fmt.Println("\n🔍 Demo 1: Testing with bot-like features")
	botFeatures := generateBotFeatures(mlConfig.FeatureSize)
	result, err := engine.Predict(context.Background(), botFeatures, "demo_bot_001")
	if err != nil {
		log.Printf("Prediction failed: %v", err)
	} else {
		printResult("Bot-like traffic", result)
	}

	// Demo 2: Test with human-like features
	fmt.Println("\n👤 Demo 2: Testing with human-like features")
	humanFeatures := generateHumanFeatures(mlConfig.FeatureSize)
	result, err = engine.Predict(context.Background(), humanFeatures, "demo_human_001")
	if err != nil {
		log.Printf("Prediction failed: %v", err)
	} else {
		printResult("Human-like traffic", result)
	}

	// Demo 3: Test with random features
	fmt.Println("\n🎲 Demo 3: Testing with random features")
	for i := 0; i < 5; i++ {
		randomFeatures := generateRandomFeatures(mlConfig.FeatureSize)
		result, err := engine.Predict(context.Background(), randomFeatures, fmt.Sprintf("demo_random_%03d", i+1))
		if err != nil {
			log.Printf("Prediction failed: %v", err)
			continue
		}
		printResult(fmt.Sprintf("Random traffic %d", i+1), result)
	}

	// Demo 4: Batch prediction
	fmt.Println("\n📦 Demo 4: Batch prediction test")
	batchResults := performBatchPrediction(engine, mlConfig.FeatureSize, 10)
	printBatchResults(batchResults)

	// Demo 5: Show statistics
	fmt.Println("\n📊 Demo 5: ML Engine Statistics")
	stats := engine.GetStatistics()
	printStatistics(stats)

	// Demo 6: Model information
	fmt.Println("\nℹ️  Demo 6: Model Information")
	printModelInfo(mlConfig)

	fmt.Println("\n🎉 ML Demo completed successfully!")
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
func performBatchPrediction(engine *ml.MLEngine, featureSize, count int) []*ml.DetectionResult {
	results := make([]*ml.DetectionResult, count)
	
	for i := 0; i < count; i++ {
		features := generateRandomFeatures(featureSize)
		result, err := engine.Predict(context.Background(), features, fmt.Sprintf("batch_%03d", i+1))
		if err != nil {
			fmt.Printf("❌ Batch prediction %d failed: %v\n", i+1, err)
			continue
		}
		results[i] = result
	}
	
	return results
}

// printResult prints a single prediction result
func printResult(label string, result *ml.DetectionResult) {
	fmt.Printf("  %s:\n", label)
	fmt.Printf("    🤖 Is Bot: %t\n", result.IsBot)
	fmt.Printf("    📊 Confidence: %.3f\n", result.Confidence)
	fmt.Printf("    🧠 Model Used: %s\n", result.ModelUsed)
	fmt.Printf("    💭 Reasoning: %s\n", result.Reasoning)
	fmt.Printf("    🕒 Timestamp: %s\n", result.Timestamp.Format("15:04:05"))
}

// printBatchResults prints batch prediction results
func printBatchResults(results []*ml.DetectionResult) {
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

// printStatistics prints ML engine statistics
func printStatistics(stats *ml.MLStatistics) {
	fmt.Printf("  📊 Total Predictions: %d\n", stats.TotalPredictions)
	fmt.Printf("  🤖 Bot Detections: %d\n", stats.BotDetections)
	fmt.Printf("  👤 Human Detections: %d\n", stats.HumanDetections)
	fmt.Printf("  📈 Average Confidence: %.3f\n", stats.AverageConfidence)
	fmt.Printf("  🎯 Model Accuracy: %.3f\n", stats.ModelAccuracy)
	fmt.Printf("  ⏱️  Training Time: %v\n", stats.TrainingTime)
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

// prettyPrintJSON prints data in formatted JSON
func prettyPrintJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
} 