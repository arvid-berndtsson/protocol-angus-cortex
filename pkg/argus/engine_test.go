package argus

import (
	"context"
	"testing"
	"time"

	"github.com/arvid-berndtsson/protocol-argus-cortex/internal/cortex"
	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	cfg := config.CaptureConfig{
		Interface:  "eth0",
		BPFFilter:  "tcp or udp",
		BufferSize: 1024 * 1024,
	}

	cortexCfg := config.CortexConfig{
		ModelPath:          "./test_model.onnx",
		DetectionThreshold: 0.85,
		BatchSize:          32,
		InferenceTimeout:   1000,
	}

	cortexEngine, err := cortex.NewEngine(cortexCfg)
	require.NoError(t, err)
	defer cortexEngine.Close()

	engine, err := NewEngine(cfg, cortexEngine)
	require.NoError(t, err)
	defer engine.Close()

	assert.NotNil(t, engine)
	assert.Equal(t, cfg, engine.config)
	assert.Equal(t, cortexEngine, engine.cortex)
	assert.NotNil(t, engine.flows)
	assert.NotNil(t, engine.stats)
}

func TestGenerateFlowID(t *testing.T) {
	engine := &Engine{}

	flowID := engine.generateFlowID("192.168.1.100", "8.8.8.8", 54321, 443)
	expected := "192.168.1.100:54321-8.8.8.8:443"
	assert.Equal(t, expected, flowID)

	flowID = engine.generateFlowID("10.0.0.1", "1.1.1.1", 12345, 80)
	expected = "10.0.0.1:12345-1.1.1.1:80"
	assert.Equal(t, expected, flowID)
}

func TestAddPacketToFlow(t *testing.T) {
	engine := &Engine{
		flows: make(map[string]*Flow),
		stats: &CaptureStats{},
	}

	packet := &Packet{
		Timestamp: time.Now(),
		Size:      1200,
		Direction: "outbound",
		Protocol:  "TCP",
		Headers:   make(map[string]interface{}),
	}

	flowID := "test-flow-1"
	engine.addPacketToFlow(flowID, packet)

	// Check that flow was created
	flow, exists := engine.flows[flowID]
	assert.True(t, exists)
	assert.Equal(t, flowID, flow.ID)
	assert.Len(t, flow.Packets, 1)
	assert.Equal(t, packet, flow.Packets[0])

	// Add another packet to the same flow
	packet2 := &Packet{
		Timestamp: time.Now(),
		Size:      800,
		Direction: "inbound",
		Protocol:  "TCP",
		Headers:   make(map[string]interface{}),
	}

	engine.addPacketToFlow(flowID, packet2)
	assert.Len(t, flow.Packets, 2)
}

func TestExtractFeatures(t *testing.T) {
	engine := &Engine{}

	flow := &Flow{
		ID:        "test-flow",
		StartTime: time.Now().Add(-5 * time.Minute),
		LastSeen:  time.Now(),
		Packets: []*Packet{
			{
				Timestamp: time.Now().Add(-4 * time.Minute),
				Size:      1200,
			},
			{
				Timestamp: time.Now().Add(-3 * time.Minute),
				Size:      800,
			},
			{
				Timestamp: time.Now().Add(-2 * time.Minute),
				Size:      1400,
			},
		},
	}

	features := engine.extractFeatures(flow)

	// Check that features array has correct size
	assert.Len(t, features, 128)

	// Check that average packet size is calculated correctly
	expectedAvgSize := float64(1200+800+1400) / 3.0
	assert.Equal(t, expectedAvgSize, features[0])

	// Check that packet count is set
	assert.Equal(t, float64(3), features[20])

	// Check that flow duration is set
	duration := flow.LastSeen.Sub(flow.StartTime).Seconds()
	assert.Equal(t, duration, features[21])
}

func TestSimulatePacketCapture(t *testing.T) {
	engine := &Engine{
		stats: &CaptureStats{},
		flows: make(map[string]*Flow),
	}

	initialPackets := engine.stats.TotalPackets
	initialFlows := engine.stats.ActiveFlows

	engine.simulatePacketCapture()

	// Check that packets were added
	assert.Greater(t, engine.stats.TotalPackets, initialPackets)
	assert.Greater(t, engine.stats.ActiveFlows, initialFlows)
}

func TestRemoveOldFlows(t *testing.T) {
	engine := &Engine{
		flows: make(map[string]*Flow),
		stats: &CaptureStats{},
	}

	// Add a recent flow
	recentFlow := &Flow{
		ID:        "recent-flow",
		LastSeen:  time.Now(),
		StartTime: time.Now().Add(-1 * time.Minute),
	}
	engine.flows["recent-flow"] = recentFlow

	// Add an old flow
	oldFlow := &Flow{
		ID:        "old-flow",
		LastSeen:  time.Now().Add(-10 * time.Minute),
		StartTime: time.Now().Add(-15 * time.Minute),
	}
	engine.flows["old-flow"] = oldFlow

	engine.removeOldFlows()

	// Check that only the recent flow remains
	assert.Contains(t, engine.flows, "recent-flow")
	assert.NotContains(t, engine.flows, "old-flow")
}

func TestGetStatistics(t *testing.T) {
	engine := &Engine{
		stats: &CaptureStats{
			TotalPackets:  100,
			ActiveFlows:   5,
			AnalyzedFlows: 3,
			LastPacket:    time.Now(),
		},
	}

	stats := engine.GetStatistics()

	assert.Equal(t, int64(100), stats.TotalPackets)
	assert.Equal(t, int64(5), stats.ActiveFlows)
	assert.Equal(t, int64(3), stats.AnalyzedFlows)
}

func TestEngineStartStop(t *testing.T) {
	cfg := config.CaptureConfig{
		Interface:  "eth0",
		BPFFilter:  "tcp or udp",
		BufferSize: 1024 * 1024,
	}

	cortexCfg := config.CortexConfig{
		ModelPath:          "./test_model.onnx",
		DetectionThreshold: 0.85,
		BatchSize:          32,
		InferenceTimeout:   1000,
	}

	cortexEngine, err := cortex.NewEngine(cortexCfg)
	require.NoError(t, err)
	defer cortexEngine.Close()

	engine, err := NewEngine(cfg, cortexEngine)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start the engine
	err = engine.Start(ctx)
	assert.NoError(t, err)

	// Wait for context to be cancelled
	<-ctx.Done()

	// Close the engine
	err = engine.Close()
	assert.NoError(t, err)
}
