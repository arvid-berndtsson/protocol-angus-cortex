package ml

import (
	"fmt"
	"math"
)

// GenerateFakeData creates synthetic training data for bot detection
func (dg *DataGenerator) GenerateFakeData(size, featureSize int) ([][]float64, []int) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	features := make([][]float64, size)
	labels := make([]int, size)

	// Generate 50% bot data, 50% human data
	botCount := size / 2

	// Generate bot-like traffic patterns
	for i := 0; i < botCount; i++ {
		features[i] = dg.generateBotFeatures(featureSize)
		labels[i] = 1 // Bot label
	}

	// Generate human-like traffic patterns
	for i := botCount; i < size; i++ {
		features[i] = dg.generateHumanFeatures(featureSize)
		labels[i] = 0 // Human label
	}

	// Shuffle the data
	dg.shuffleData(features, labels)

	return features, labels
}

// generateBotFeatures creates features that simulate bot behavior
func (dg *DataGenerator) generateBotFeatures(featureSize int) []float64 {
	features := make([]float64, featureSize)

	// Bot characteristics:
	// - Regular timing patterns (low variance)
	// - Consistent packet sizes
	// - High request rates
	// - Predictable behavior patterns

	// Packet timing features (0-19)
	for i := 0; i < 20; i++ {
		if i < featureSize {
			// Bots have more regular timing
			baseTime := 0.1 + dg.rand.Float64()*0.2
			variance := 0.01 + dg.rand.Float64()*0.05 // Low variance
			features[i] = baseTime + (dg.rand.Float64()-0.5)*variance
		}
	}

	// Packet size features (20-39)
	for i := 20; i < 40; i++ {
		if i < featureSize {
			// Bots often have consistent packet sizes
			baseSize := 0.3 + dg.rand.Float64()*0.4
			variance := 0.05 + dg.rand.Float64()*0.1 // Low variance
			features[i] = baseSize + (dg.rand.Float64()-0.5)*variance
		}
	}

	// Request rate features (40-59)
	for i := 40; i < 60; i++ {
		if i < featureSize {
			// Bots typically have high, consistent request rates
			features[i] = 0.7 + dg.rand.Float64()*0.3
		}
	}

	// Protocol behavior features (60-79)
	for i := 60; i < 80; i++ {
		if i < featureSize {
			// Bots often follow strict protocol patterns
			features[i] = 0.8 + dg.rand.Float64()*0.2
		}
	}

	// Flow duration features (80-99)
	for i := 80; i < 100; i++ {
		if i < featureSize {
			// Bots often have longer, more persistent flows
			features[i] = 0.6 + dg.rand.Float64()*0.4
		}
	}

	// Entropy features (100-119)
	for i := 100; i < 120; i++ {
		if i < featureSize {
			// Bots often have lower entropy in their behavior
			features[i] = 0.2 + dg.rand.Float64()*0.3
		}
	}

	// Additional behavioral features (120+)
	for i := 120; i < featureSize; i++ {
		// Random bot-specific patterns
		features[i] = dg.rand.Float64() * 0.8
	}

	return features
}

// generateHumanFeatures creates features that simulate human behavior
func (dg *DataGenerator) generateHumanFeatures(featureSize int) []float64 {
	features := make([]float64, featureSize)

	// Human characteristics:
	// - Irregular timing patterns (high variance)
	// - Variable packet sizes
	// - Lower request rates
	// - Unpredictable behavior patterns

	// Packet timing features (0-19)
	for i := 0; i < 20; i++ {
		if i < featureSize {
			// Humans have irregular timing
			baseTime := 0.5 + dg.rand.Float64()*0.5
			variance := 0.2 + dg.rand.Float64()*0.3 // High variance
			features[i] = baseTime + (dg.rand.Float64()-0.5)*variance
		}
	}

	// Packet size features (20-39)
	for i := 20; i < 40; i++ {
		if i < featureSize {
			// Humans have variable packet sizes
			baseSize := 0.1 + dg.rand.Float64()*0.9
			variance := 0.3 + dg.rand.Float64()*0.4 // High variance
			features[i] = baseSize + (dg.rand.Float64()-0.5)*variance
		}
	}

	// Request rate features (40-59)
	for i := 40; i < 60; i++ {
		if i < featureSize {
			// Humans typically have lower, variable request rates
			features[i] = 0.1 + dg.rand.Float64()*0.4
		}
	}

	// Protocol behavior features (60-79)
	for i := 60; i < 80; i++ {
		if i < featureSize {
			// Humans often have less strict protocol adherence
			features[i] = 0.3 + dg.rand.Float64()*0.5
		}
	}

	// Flow duration features (80-99)
	for i := 80; i < 100; i++ {
		if i < featureSize {
			// Humans often have shorter, more variable flows
			features[i] = 0.1 + dg.rand.Float64()*0.6
		}
	}

	// Entropy features (100-119)
	for i := 100; i < 120; i++ {
		if i < featureSize {
			// Humans often have higher entropy in their behavior
			features[i] = 0.5 + dg.rand.Float64()*0.5
		}
	}

	// Additional behavioral features (120+)
	for i := 120; i < featureSize; i++ {
		// Random human-specific patterns
		features[i] = 0.2 + dg.rand.Float64()*0.8
	}

	return features
}

// shuffleData randomly shuffles the features and labels while maintaining correspondence
func (dg *DataGenerator) shuffleData(features [][]float64, labels []int) {
	for i := len(features) - 1; i > 0; i-- {
		j := dg.rand.Intn(i + 1)
		features[i], features[j] = features[j], features[i]
		labels[i], labels[j] = labels[j], labels[i]
	}
}

// GenerateRealisticFeatures creates features that simulate real network traffic
func (dg *DataGenerator) GenerateRealisticFeatures(featureSize int) []float64 {
	features := make([]float64, featureSize)

	// Simulate realistic network traffic patterns
	for i := 0; i < featureSize; i++ {
		switch {
		case i < 20: // Timing features
			// Exponential distribution for inter-packet times
			lambda := 0.1 + dg.rand.Float64()*0.2
			features[i] = -math.Log(1-dg.rand.Float64()) / lambda
			features[i] = math.Min(features[i], 1.0) // Normalize to [0,1]

		case i < 40: // Size features
			// Normal distribution for packet sizes
			mean := 0.3 + dg.rand.Float64()*0.4
			stddev := 0.1 + dg.rand.Float64()*0.2
			features[i] = dg.rand.NormFloat64()*stddev + mean
			features[i] = math.Max(0, math.Min(features[i], 1.0)) // Clamp to [0,1]

		case i < 60: // Rate features
			// Poisson-like distribution for request rates
			rate := 0.1 + dg.rand.Float64()*0.8
			features[i] = rate

		case i < 80: // Protocol features
			// Categorical-like features
			features[i] = float64(dg.rand.Intn(10)) / 10.0

		case i < 100: // Duration features
			// Weibull distribution for flow durations
			scale := 0.2 + dg.rand.Float64()*0.6
			shape := 1.0 + dg.rand.Float64()*2.0
			u := dg.rand.Float64()
			features[i] = scale * math.Pow(-math.Log(1-u), 1/shape)
			features[i] = math.Min(features[i], 1.0)

		case i < 120: // Entropy features
			// Shannon entropy-like features
			features[i] = dg.rand.Float64()

		default: // Additional features
			features[i] = dg.rand.Float64()
		}
	}

	return features
}

// GenerateAnomalousFeatures creates features that represent anomalous behavior
func (dg *DataGenerator) GenerateAnomalousFeatures(featureSize int) []float64 {
	features := make([]float64, featureSize)

	// Generate extreme values that might indicate anomalies
	for i := 0; i < featureSize; i++ {
		// 20% chance of extreme value
		if dg.rand.Float64() < 0.2 {
			// Generate extreme value (very high or very low)
			if dg.rand.Float64() < 0.5 {
				features[i] = 0.9 + dg.rand.Float64()*0.1 // Very high
			} else {
				features[i] = dg.rand.Float64() * 0.1 // Very low
			}
		} else {
			// Normal value
			features[i] = dg.rand.Float64()
		}
	}

	return features
}

// CalculateFeatureStatistics calculates basic statistics for feature analysis
func (dg *DataGenerator) CalculateFeatureStatistics(features [][]float64) map[string]float64 {
	if len(features) == 0 {
		return nil
	}

	featureSize := len(features[0])
	stats := make(map[string]float64)

	// Calculate mean for each feature
	for i := 0; i < featureSize; i++ {
		var sum float64
		for _, feature := range features {
			if i < len(feature) {
				sum += feature[i]
			}
		}
		stats[fmt.Sprintf("mean_%d", i)] = sum / float64(len(features))
	}

	// Calculate variance for each feature
	for i := 0; i < featureSize; i++ {
		mean := stats[fmt.Sprintf("mean_%d", i)]
		var variance float64
		for _, feature := range features {
			if i < len(feature) {
				diff := feature[i] - mean
				variance += diff * diff
			}
		}
		stats[fmt.Sprintf("variance_%d", i)] = variance / float64(len(features))
	}

	return stats
}
