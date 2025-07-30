package ml

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sync"
	"time"

	"gonum.org/v1/gonum/mat"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

// MLEngine represents a ML engine using Gorgonia and Gonum
type MLEngine struct {
	// Neural Network (Gorgonia)
	nnModel *NeuralNetwork

	// SVM Classifier (Gonum-based)
	svmModel *SVMClassifier

	// Data generation
	dataGen *DataGenerator

	// Configuration
	config MLConfig
	mu     sync.RWMutex
	stats  *MLStatistics
	ctx    context.Context
	cancel context.CancelFunc
}

// MLConfig holds configuration for the ML engine
type MLConfig struct {
	ModelType          string  `yaml:"model_type"` // "neural_network", "svm", "ensemble"
	DetectionThreshold float64 `yaml:"detection_threshold"`
	BatchSize          int     `yaml:"batch_size"`
	TrainingEpochs     int     `yaml:"training_epochs"`
	LearningRate       float64 `yaml:"learning_rate"`
	FeatureSize        int     `yaml:"feature_size"`
	GenerateFakeData   bool    `yaml:"generate_fake_data"`
	FakeDataSize       int     `yaml:"fake_data_size"`
}

// MLStatistics holds ML engine statistics
type MLStatistics struct {
	TotalPredictions  int64         `json:"total_predictions"`
	BotDetections     int64         `json:"bot_detections"`
	HumanDetections   int64         `json:"human_detections"`
	AverageConfidence float64       `json:"average_confidence"`
	ModelAccuracy     float64       `json:"model_accuracy"`
	TrainingTime      time.Duration `json:"training_time"`
	LastPrediction    time.Time     `json:"last_prediction"`
	mu                sync.RWMutex
}

// NeuralNetwork represents a Gorgonia-based neural network
type NeuralNetwork struct {
	graph   *gorgonia.ExprGraph
	input   *gorgonia.Node
	output  *gorgonia.Node
	vm      gorgonia.VM
	trained bool
}

// SVMClassifier represents a Support Vector Machine classifier using Gonum
type SVMClassifier struct {
	weights *mat.VecDense
	bias    float64
	trained bool
}

// DetectionResult represents the result of ML-based bot detection
type DetectionResult struct {
	IsBot      bool      `json:"is_bot"`
	Confidence float64   `json:"confidence"`
	Features   []float64 `json:"features"`
	Reasoning  string    `json:"reasoning"`
	ModelUsed  string    `json:"model_used"`
	Timestamp  time.Time `json:"timestamp"`
	FlowID     string    `json:"flow_id"`
}

// DataGenerator generates fake training data for bot detection
type DataGenerator struct {
	rand *rand.Rand
	mu   sync.Mutex
}

// NewMLEngine creates a new ML engine instance
func NewMLEngine(config MLConfig) (*MLEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &MLEngine{
		config: config,
		stats:  &MLStatistics{},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize data generator
	engine.dataGen = &DataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Initialize models based on configuration
	if err := engine.initializeModels(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize models: %w", err)
	}

	// Generate and train on fake data if enabled
	if config.GenerateFakeData {
		if err := engine.TrainOnFakeData(); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to train on fake data: %w", err)
		}
	}

	slog.Info("ML engine initialized",
		"model_type", config.ModelType,
		"threshold", config.DetectionThreshold,
		"feature_size", config.FeatureSize)

	return engine, nil
}

// initializeModels initializes the selected ML models
func (e *MLEngine) initializeModels() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch e.config.ModelType {
	case "neural_network":
		return e.initializeNeuralNetwork()
	case "svm":
		return e.initializeSVM()
	case "ensemble":
		return e.initializeEnsemble()
	default:
		return fmt.Errorf("unsupported model type: %s", e.config.ModelType)
	}
}

// initializeNeuralNetwork sets up a Gorgonia-based neural network
func (e *MLEngine) initializeNeuralNetwork() error {
	// Create computation graph
	g := gorgonia.NewGraph()

	// Input layer
	input := gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(1, e.config.FeatureSize), gorgonia.WithName("input"))

	// Hidden layer weights and bias
	hiddenWeights := gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(e.config.FeatureSize, 64), gorgonia.WithName("hidden_weights"))
	hiddenBias := gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(1, 64), gorgonia.WithName("hidden_bias"))

	// Output layer weights and bias
	outputWeights := gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(64, 1), gorgonia.WithName("output_weights"))
	outputBias := gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(1, 1), gorgonia.WithName("output_bias"))

	// Forward pass - simplified to avoid complex Gorgonia API
	// For now, we'll use a simple approach that doesn't require complex matrix operations
	hidden := gorgonia.Must(gorgonia.Add(gorgonia.Must(gorgonia.Mul(input, hiddenWeights)), hiddenBias))
	hidden = gorgonia.Must(gorgonia.Rectify(hidden))

	output := gorgonia.Must(gorgonia.Add(gorgonia.Must(gorgonia.Mul(hidden, outputWeights)), outputBias))
	output = gorgonia.Must(gorgonia.Sigmoid(output))

	// Create VM
	vm := gorgonia.NewTapeMachine(g)

	e.nnModel = &NeuralNetwork{
		graph:   g,
		input:   input,
		output:  output,
		vm:      vm,
		trained: false,
	}

	return nil
}

// initializeSVM sets up a simple SVM classifier using Gonum
func (e *MLEngine) initializeSVM() error {
	weights := mat.NewVecDense(e.config.FeatureSize, nil)
	e.svmModel = &SVMClassifier{
		weights: weights,
		bias:    0.0,
		trained: false,
	}
	return nil
}

// initializeEnsemble sets up all models for ensemble prediction
func (e *MLEngine) initializeEnsemble() error {
	if err := e.initializeNeuralNetwork(); err != nil {
		return err
	}
	if err := e.initializeSVM(); err != nil {
		return err
	}
	return nil
}

// TrainOnFakeData generates fake data and trains the models
func (e *MLEngine) TrainOnFakeData() error {
	slog.Info("Generating fake training data", "size", e.config.FakeDataSize)

	startTime := time.Now()

	// Generate fake data
	features, labels := e.dataGen.GenerateFakeData(e.config.FakeDataSize, e.config.FeatureSize)

	// Train models based on type
	switch e.config.ModelType {
	case "neural_network":
		return e.trainNeuralNetwork(features, labels)
	case "svm":
		return e.trainSVM(features, labels)
	case "ensemble":
		return e.trainEnsemble(features, labels)
	default:
		return fmt.Errorf("unsupported model type for training: %s", e.config.ModelType)
	}

	e.stats.mu.Lock()
	e.stats.TrainingTime = time.Since(startTime)
	e.stats.mu.Unlock()

	slog.Info("Training completed", "duration", time.Since(startTime))
	return nil
}

// Predict performs bot detection using the trained model
func (e *MLEngine) Predict(ctx context.Context, features []float64, flowID string) (*DetectionResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var confidence float64
	var modelUsed string

	switch e.config.ModelType {
	case "neural_network":
		conf, err := e.predictNeuralNetwork(features)
		if err != nil {
			return nil, err
		}
		confidence = conf
		modelUsed = "neural_network"

	case "svm":
		conf, err := e.predictSVM(features)
		if err != nil {
			return nil, err
		}
		confidence = conf
		modelUsed = "svm"

	case "ensemble":
		conf, err := e.predictEnsemble(features)
		if err != nil {
			return nil, err
		}
		confidence = conf
		modelUsed = "ensemble"

	default:
		return nil, fmt.Errorf("unsupported model type: %s", e.config.ModelType)
	}

	isBot := confidence > e.config.DetectionThreshold
	reasoning := e.generateReasoning(features, confidence, modelUsed)

	result := &DetectionResult{
		IsBot:      isBot,
		Confidence: confidence,
		Features:   features,
		Reasoning:  reasoning,
		ModelUsed:  modelUsed,
		Timestamp:  time.Now(),
		FlowID:     flowID,
	}

	e.updateStats(result)

	return result, nil
}

// predictNeuralNetwork performs prediction using the neural network
func (e *MLEngine) predictNeuralNetwork(features []float64) (float64, error) {
	if e.nnModel == nil || !e.nnModel.trained {
		return e.simulatePrediction(features), nil
	}

	// Convert features to tensor
	inputTensor := tensor.New(tensor.WithShape(1, len(features)), tensor.WithBacking(features))

	// Set input value
	gorgonia.Let(e.nnModel.input, inputTensor)

	// Run forward pass
	if err := e.nnModel.vm.RunAll(); err != nil {
		return 0, fmt.Errorf("neural network inference failed: %w", err)
	}

	// Get output
	outputValue := e.nnModel.output.Value()
	if outputTensor, ok := outputValue.(tensor.Tensor); ok {
		if outputData, ok := outputTensor.Data().([]float64); ok && len(outputData) > 0 {
			return outputData[0], nil
		}
	}

	return 0, fmt.Errorf("failed to extract neural network output")
}

// predictSVM performs prediction using SVM with Gonum
func (e *MLEngine) predictSVM(features []float64) (float64, error) {
	if e.svmModel == nil || !e.svmModel.trained {
		return e.simulatePrediction(features), nil
	}

	// Create feature vector
	featureVec := mat.NewVecDense(len(features), features)

	// Linear SVM prediction: w^T * x + b
	var prediction float64
	prediction = mat.Dot(e.svmModel.weights, featureVec) + e.svmModel.bias

	// Convert to probability using sigmoid
	return 1.0 / (1.0 + math.Exp(-prediction)), nil
}

// predictEnsemble performs prediction using all models and averages results
func (e *MLEngine) predictEnsemble(features []float64) (float64, error) {
	var predictions []float64

	// Neural network prediction
	if nnPred, err := e.predictNeuralNetwork(features); err == nil {
		predictions = append(predictions, nnPred)
	}

	// SVM prediction
	if svmPred, err := e.predictSVM(features); err == nil {
		predictions = append(predictions, svmPred)
	}

	if len(predictions) == 0 {
		return e.simulatePrediction(features), nil
	}

	// Average predictions
	var sum float64
	for _, pred := range predictions {
		sum += pred
	}
	return sum / float64(len(predictions)), nil
}

// simulatePrediction provides a fallback prediction when models aren't trained
func (e *MLEngine) simulatePrediction(features []float64) float64 {
	// Simple heuristic-based prediction
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
	}

	// Normalize to [0, 1]
	score = math.Min(score, 1.0)
	return score
}

// generateReasoning provides human-readable explanation for the prediction
func (e *MLEngine) generateReasoning(features []float64, confidence float64, modelUsed string) string {
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

	reasoning += modelUsed + " model analysis. "

	// Add specific feature insights
	if len(features) > 0 {
		reasoning += "Key indicators include packet timing patterns, "
		reasoning += "protocol behavior consistency, and flow characteristics."
	}

	return reasoning
}

// updateStats updates the ML engine statistics
func (e *MLEngine) updateStats(result *DetectionResult) {
	e.stats.mu.Lock()
	defer e.stats.mu.Unlock()

	e.stats.TotalPredictions++
	if result.IsBot {
		e.stats.BotDetections++
	} else {
		e.stats.HumanDetections++
	}

	// Update average confidence
	total := float64(e.stats.TotalPredictions)
	e.stats.AverageConfidence = (e.stats.AverageConfidence*(total-1) + result.Confidence) / total

	e.stats.LastPrediction = result.Timestamp
}

// GetStatistics returns the current ML engine statistics
func (e *MLEngine) GetStatistics() *MLStatistics {
	e.stats.mu.RLock()
	defer e.stats.mu.RUnlock()

	stats := *e.stats // Copy to avoid race conditions
	return &stats
}

// Close cleans up resources
func (e *MLEngine) Close() error {
	e.cancel()

	if e.nnModel != nil && e.nnModel.vm != nil {
		e.nnModel.vm.Close()
	}

	return nil
}

// Training methods
func (e *MLEngine) trainNeuralNetwork(features [][]float64, labels []int) error {
	// Simplified training - in real implementation, this would use backpropagation
	e.nnModel.trained = true
	return nil
}

func (e *MLEngine) trainSVM(features [][]float64, labels []int) error {
	// Simplified SVM training using gradient descent
	if len(features) == 0 || len(labels) == 0 {
		return fmt.Errorf("no training data provided")
	}

	// Simple linear SVM training
	for i := 0; i < 100; i++ { // 100 iterations
		for j, feature := range features {
			if j < len(labels) {
				// Update weights based on gradient
				label := float64(labels[j])
				if label == 0 {
					label = -1 // Convert to -1/1 labels
				}

				// Simple gradient update
				for k, f := range feature {
					if k < e.svmModel.weights.Len() {
						currentWeight := e.svmModel.weights.AtVec(k)
						newWeight := currentWeight + 0.01*label*f
						e.svmModel.weights.SetVec(k, newWeight)
					}
				}
			}
		}
	}

	e.svmModel.trained = true
	return nil
}

func (e *MLEngine) trainEnsemble(features [][]float64, labels []int) error {
	// Train all models
	if err := e.trainNeuralNetwork(features, labels); err != nil {
		return err
	}
	if err := e.trainSVM(features, labels); err != nil {
		return err
	}
	return nil
}
