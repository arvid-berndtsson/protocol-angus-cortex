package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/arvid-berndtsson/protocol-argus-cortex/internal/cortex"
	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/argus"
	"github.com/arvid-berndtsson/protocol-argus-cortex/pkg/config"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the API server
type Server struct {
	config       config.ServerConfig
	cortexEngine *cortex.Engine
	argusEngine  *argus.Engine
	router       *mux.Router
	server       *http.Server
	metrics      *Metrics
}

// Metrics holds Prometheus metrics
type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	botDetections   prometheus.Counter
	humanDetections prometheus.Counter
	activeFlows     prometheus.Gauge
	totalPackets    prometheus.Counter
}

// NewServer creates a new API server
func NewServer(cfg config.ServerConfig, cortexEngine *cortex.Engine, argusEngine *argus.Engine) *Server {
	router := mux.NewRouter()

	server := &Server{
		config:       cfg,
		cortexEngine: cortexEngine,
		argusEngine:  argusEngine,
		router:       router,
		metrics:      newMetrics(),
	}

	server.setupRoutes()
	server.setupMiddleware()

	return server
}

// newMetrics creates and registers Prometheus metrics
func newMetrics() *Metrics {
	metrics := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "argus_cortex_requests_total",
				Help: "Total number of API requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "argus_cortex_request_duration_seconds",
				Help:    "Request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		botDetections: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "argus_cortex_bot_detections_total",
				Help: "Total number of bot detections",
			},
		),
		humanDetections: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "argus_cortex_human_detections_total",
				Help: "Total number of human detections",
			},
		),
		activeFlows: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "argus_cortex_active_flows",
				Help: "Number of active network flows",
			},
		),
		totalPackets: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "argus_cortex_packets_total",
				Help: "Total number of packets captured",
			},
		),
	}

	// Register metrics
	prometheus.MustRegister(
		metrics.requestsTotal,
		metrics.requestDuration,
		metrics.botDetections,
		metrics.humanDetections,
		metrics.activeFlows,
		metrics.totalPackets,
	)

	return metrics
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// API endpoints
	s.router.HandleFunc("/api/v1/status", s.handleStatus).Methods("GET")
	s.router.HandleFunc("/api/v1/statistics", s.handleStatistics).Methods("GET")
	s.router.HandleFunc("/api/v1/flows", s.handleFlows).Methods("GET")
	s.router.HandleFunc("/api/v1/analyze", s.handleAnalyze).Methods("POST")

	// Prometheus metrics
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Root endpoint
	s.router.HandleFunc("/", s.handleRoot).Methods("GET")
}

// setupMiddleware configures request middleware
func (s *Server) setupMiddleware() {
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.metricsMiddleware)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.APIPort),
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("Starting API server", "port", s.config.APIPort)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handleRoot handles the root endpoint
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"name":        "Protocol Argus Cortex",
		"version":     "1.0.0",
		"description": "Advanced network traffic analysis engine for bot detection",
		"endpoints": map[string]string{
			"health":     "/health",
			"status":     "/api/v1/status",
			"statistics": "/api/v1/statistics",
			"flows":      "/api/v1/flows",
			"analyze":    "/api/v1/analyze",
			"metrics":    "/metrics",
		},
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(time.Now()).String(), // Simplified
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleStatus handles status requests
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	cortexStats := s.cortexEngine.GetStatistics()
	argusStats := s.argusEngine.GetStatistics()

	response := map[string]interface{}{
		"status": "operational",
		"cortex": map[string]interface{}{
			"total_inferences":   cortexStats.TotalInferences,
			"bot_detections":     cortexStats.BotDetections,
			"human_detections":   cortexStats.HumanDetections,
			"average_confidence": cortexStats.AverageConfidence,
			"last_inference":     cortexStats.LastInference,
		},
		"argus": map[string]interface{}{
			"total_packets":  argusStats.TotalPackets,
			"active_flows":   argusStats.ActiveFlows,
			"analyzed_flows": argusStats.AnalyzedFlows,
			"last_packet":    argusStats.LastPacket,
		},
		"timestamp": time.Now().UTC(),
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleStatistics handles statistics requests
func (s *Server) handleStatistics(w http.ResponseWriter, r *http.Request) {
	cortexStats := s.cortexEngine.GetStatistics()
	argusStats := s.argusEngine.GetStatistics()

	// Update Prometheus metrics
	s.metrics.botDetections.Add(float64(cortexStats.BotDetections))
	s.metrics.humanDetections.Add(float64(cortexStats.HumanDetections))
	s.metrics.activeFlows.Set(float64(argusStats.ActiveFlows))
	s.metrics.totalPackets.Add(float64(argusStats.TotalPackets))

	response := map[string]interface{}{
		"cortex": cortexStats,
		"argus":  argusStats,
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleFlows handles flow listing requests
func (s *Server) handleFlows(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would return actual flow data
	response := map[string]interface{}{
		"flows": []map[string]interface{}{
			{
				"id":         "192.168.1.100:54321-8.8.8.8:443",
				"src_ip":     "192.168.1.100",
				"dst_ip":     "8.8.8.8",
				"protocol":   "TCP",
				"packets":    15,
				"start_time": time.Now().Add(-5 * time.Minute),
				"last_seen":  time.Now(),
			},
			{
				"id":         "10.0.0.50:12345-1.1.1.1:80",
				"src_ip":     "10.0.0.50",
				"dst_ip":     "1.1.1.1",
				"protocol":   "TCP",
				"packets":    8,
				"start_time": time.Now().Add(-2 * time.Minute),
				"last_seen":  time.Now(),
			},
		},
		"total": 2,
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleAnalyze handles manual analysis requests
func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Features []float64 `json:"features"`
		FlowID   string    `json:"flow_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(request.Features) == 0 {
		s.writeError(w, http.StatusBadRequest, "Features array is required")
		return
	}

	if request.FlowID == "" {
		request.FlowID = fmt.Sprintf("manual_%d", time.Now().Unix())
	}

	// Perform analysis
	result, err := s.cortexEngine.Analyze(r.Context(), request.Features, request.FlowID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Analysis failed: %v", err))
		return
	}

	// Update metrics based on result
	if result.IsBot {
		s.metrics.botDetections.Inc()
	} else {
		s.metrics.humanDetections.Inc()
	}

	s.writeJSON(w, http.StatusOK, result)
}

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	response := map[string]interface{}{
		"error":     message,
		"status":    status,
		"timestamp": time.Now().UTC(),
	}

	s.writeJSON(w, status, response)
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		slog.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", duration,
			"user_agent", r.UserAgent(),
		)
	})
}

// metricsMiddleware updates Prometheus metrics
func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		s.metrics.requestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			fmt.Sprintf("%d", wrapped.statusCode),
		).Inc()

		s.metrics.requestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration.Seconds())
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
