package argus

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/arvid-berndtsson/protocol-argus-cortex/internal/cortex"
	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/config"
	"github.com/google/gopacket/pcap"
)

// Engine represents the packet capture and feature extraction engine
type Engine struct {
	config  config.CaptureConfig
	cortex  *cortex.Engine
	handle  *pcap.Handle
	flows   map[string]*Flow
	flowsMu sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	stats   *CaptureStats
}

// Flow represents a network flow being tracked
type Flow struct {
	ID              string
	SrcIP           net.IP
	DstIP           net.IP
	SrcPort         uint16
	DstPort         uint16
	Protocol        string
	Packets         []*Packet
	StartTime       time.Time
	LastSeen        time.Time
	Features        []float64
	AnalysisPending bool
	mu              sync.RWMutex
}

// Packet represents a captured network packet
type Packet struct {
	Timestamp time.Time
	Size      int
	Direction string // "inbound" or "outbound"
	Protocol  string
	Headers   map[string]interface{}
}

// CaptureStats holds packet capture statistics
type CaptureStats struct {
	TotalPackets  int64     `json:"total_packets"`
	ActiveFlows   int64     `json:"active_flows"`
	AnalyzedFlows int64     `json:"analyzed_flows"`
	LastPacket    time.Time `json:"last_packet"`
	mu            sync.RWMutex
}

// NewEngine creates a new Argus engine instance
func NewEngine(cfg config.CaptureConfig, cortexEngine *cortex.Engine) (*Engine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &Engine{
		config: cfg,
		cortex: cortexEngine,
		flows:  make(map[string]*Flow),
		ctx:    ctx,
		cancel: cancel,
		stats:  &CaptureStats{},
	}

	// Initialize packet capture handle
	if err := engine.initializeCapture(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize packet capture: %w", err)
	}

	slog.Info("Argus engine initialized",
		"interface", cfg.Interface,
		"bpf_filter", cfg.BPFFilter,
		"buffer_size", cfg.BufferSize)

	return engine, nil
}

// initializeCapture sets up the packet capture interface
func (e *Engine) initializeCapture() error {
	// In a real implementation, this would open the actual network interface
	// For now, we'll simulate the handle creation
	slog.Info("Initializing packet capture", "interface", e.config.Interface)

	// Simulate handle creation
	e.handle = &pcap.Handle{} // This would be the actual handle in real implementation

	return nil
}

// Start begins packet capture and analysis
func (e *Engine) Start(ctx context.Context) error {
	slog.Info("Starting packet capture")

	// Start packet processing goroutine
	go e.processPackets(ctx)

	// Start flow analysis goroutine
	go e.analyzeFlows(ctx)

	// Start flow cleanup goroutine
	go e.cleanupFlows(ctx)

	return nil
}

// processPackets handles incoming packets
func (e *Engine) processPackets(ctx context.Context) {
	// In a real implementation, this would read from the pcap handle
	// For simulation, we'll generate some fake packets
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simulate packet capture
			e.simulatePacketCapture()
		}
	}
}

// simulatePacketCapture generates simulated network packets
func (e *Engine) simulatePacketCapture() {
	// Generate some realistic-looking packet data
	packets := []struct {
		srcIP   string
		dstIP   string
		srcPort uint16
		dstPort uint16
		size    int
	}{
		{"192.168.1.100", "8.8.8.8", 54321, 443, 1200},
		{"10.0.0.50", "1.1.1.1", 12345, 80, 800},
		{"172.16.0.10", "208.67.222.222", 65432, 53, 512},
	}

	for _, pkt := range packets {
		flowID := e.generateFlowID(pkt.srcIP, pkt.dstIP, pkt.srcPort, pkt.dstPort)

		packet := &Packet{
			Timestamp: time.Now(),
			Size:      pkt.size,
			Direction: "outbound",
			Protocol:  "TCP",
			Headers:   make(map[string]interface{}),
		}

		e.addPacketToFlow(flowID, packet)
	}

	e.stats.mu.Lock()
	e.stats.TotalPackets += int64(len(packets))
	e.stats.LastPacket = time.Now()
	e.stats.mu.Unlock()
}

// addPacketToFlow adds a packet to the appropriate flow
func (e *Engine) addPacketToFlow(flowID string, packet *Packet) {
	e.flowsMu.Lock()
	defer e.flowsMu.Unlock()

	flow, exists := e.flows[flowID]
	if !exists {
		flow = &Flow{
			ID:        flowID,
			Packets:   make([]*Packet, 0),
			StartTime: time.Now(),
		}
		e.flows[flowID] = flow
	}

	flow.mu.Lock()
	flow.Packets = append(flow.Packets, packet)
	flow.LastSeen = packet.Timestamp
	flow.mu.Unlock()

	// Update active flows count
	e.stats.mu.Lock()
	e.stats.ActiveFlows = int64(len(e.flows))
	e.stats.mu.Unlock()
}

// generateFlowID creates a unique identifier for a network flow
func (e *Engine) generateFlowID(srcIP, dstIP string, srcPort, dstPort uint16) string {
	return fmt.Sprintf("%s:%d-%s:%d", srcIP, srcPort, dstIP, dstPort)
}

// analyzeFlows periodically analyzes flows for bot detection
func (e *Engine) analyzeFlows(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.performFlowAnalysis()
		}
	}
}

// performFlowAnalysis analyzes flows that are ready for analysis
func (e *Engine) performFlowAnalysis() {
	e.flowsMu.RLock()
	flows := make([]*Flow, 0, len(e.flows))
	for _, flow := range e.flows {
		if !flow.AnalysisPending && len(flow.Packets) >= 10 {
			flows = append(flows, flow)
		}
	}
	e.flowsMu.RUnlock()

	for _, flow := range flows {
		flow.mu.Lock()
		flow.AnalysisPending = true
		flow.mu.Unlock()

		// Extract features from the flow
		features := e.extractFeatures(flow)

		// Send to Cortex for analysis
		go func(f *Flow, feat []float64) {
			result, err := e.cortex.Analyze(e.ctx, feat, f.ID)
			if err != nil {
				slog.Error("Failed to analyze flow", "flow_id", f.ID, "error", err)
				return
			}

			slog.Info("Flow analysis completed",
				"flow_id", f.ID,
				"is_bot", result.IsBot,
				"confidence", result.Confidence)

			// Update statistics
			e.stats.mu.Lock()
			e.stats.AnalyzedFlows++
			e.stats.mu.Unlock()
		}(flow, features)
	}
}

// extractFeatures extracts behavioral features from a flow
func (e *Engine) extractFeatures(flow *Flow) []float64 {
	flow.mu.RLock()
	defer flow.mu.RUnlock()

	features := make([]float64, 128) // Match the model input size

	if len(flow.Packets) == 0 {
		return features
	}

	// Calculate packet size statistics
	var totalSize int
	var sizes []int
	for _, pkt := range flow.Packets {
		totalSize += pkt.Size
		sizes = append(sizes, pkt.Size)
	}
	avgSize := float64(totalSize) / float64(len(flow.Packets))
	features[0] = avgSize

	// Calculate timing patterns
	if len(flow.Packets) > 1 {
		var intervals []float64
		for i := 1; i < len(flow.Packets); i++ {
			interval := flow.Packets[i].Timestamp.Sub(flow.Packets[i-1].Timestamp).Seconds()
			intervals = append(intervals, interval)
		}

		// Calculate timing variance
		var sum, sumSq float64
		for _, interval := range intervals {
			sum += interval
			sumSq += interval * interval
		}
		mean := sum / float64(len(intervals))
		variance := (sumSq / float64(len(intervals))) - (mean * mean)
		features[10] = variance
	}

	// Protocol-specific features
	features[20] = float64(len(flow.Packets))                  // Packet count
	features[21] = flow.LastSeen.Sub(flow.StartTime).Seconds() // Flow duration

	// Add some realistic noise
	for i := 0; i < len(features); i++ {
		if features[i] == 0 {
			features[i] = float64(i%10) / 10.0 // Add some pattern
		}
	}

	return features
}

// cleanupFlows removes old flows
func (e *Engine) cleanupFlows(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.removeOldFlows()
		}
	}
}

// removeOldFlows removes flows that haven't been seen recently
func (e *Engine) removeOldFlows() {
	cutoff := time.Now().Add(-5 * time.Minute)

	e.flowsMu.Lock()
	defer e.flowsMu.Unlock()

	for flowID, flow := range e.flows {
		if flow.LastSeen.Before(cutoff) {
			delete(e.flows, flowID)
		}
	}

	// Update active flows count
	e.stats.mu.Lock()
	e.stats.ActiveFlows = int64(len(e.flows))
	e.stats.mu.Unlock()
}

// GetStatistics returns current capture statistics
func (e *Engine) GetStatistics() *CaptureStats {
	e.stats.mu.RLock()
	defer e.stats.mu.RUnlock()

	// Create a copy without the mutex to avoid copying lock value
	stats := CaptureStats{
		TotalPackets:  e.stats.TotalPackets,
		ActiveFlows:   e.stats.ActiveFlows,
		AnalyzedFlows: e.stats.AnalyzedFlows,
		LastPacket:    e.stats.LastPacket,
	}
	return &stats
}

// Close shuts down the Argus engine
func (e *Engine) Close() error {
	e.cancel()
	if e.handle != nil {
		// In real implementation: e.handle.Close()
	}
	slog.Info("Argus engine shutdown complete")
	return nil
}
