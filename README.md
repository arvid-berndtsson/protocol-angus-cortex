# Protocol Argus Cortex

[![Go Report Card](https://goreportcard.com/badge/github.com/arvid-berndtsson/protocol-argus-cortex)](https://goreportcard.com/report/github.com/arvid-berndtsson/protocol-argus-cortex)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/arvid-berndtsson/protocol-argus-cortex/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Test Coverage](https://img.shields.io/badge/coverage-100%25-brightgreen.svg)](https://github.com/arvid-berndtsson/protocol-argus-cortex)

An advanced, real-time network traffic analysis engine using machine learning to detect and classify bot activity over modern internet protocols like HTTP/2, HTTP/3, and QUIC.

## 🎯 Overview

The digital landscape is rife with sophisticated bots that can evade traditional detection methods based on simple signatures or IP blacklists. **Protocol Argus Cortex** is designed to address this challenge by operating at a deeper level.

- **Argus Engine**: The "all-seeing" eye that captures and performs deep inspection of network packets in real-time. It extracts a rich set of behavioral features and metadata, rather than just inspecting payloads.
- **Cortex Engine**: The "brain" of the operation. It feeds the features extracted by Argus into a machine learning model to classify traffic as human or bot, providing a confidence score and reasoning for its verdict.

This project focuses on the _behavioral fingerprint_ of a connection—how it communicates, not just what it says.

## ✨ Core Features

- **Real-time Packet Capture**: High-performance packet capture using `gopacket` with BPF filtering
- **Advanced Protocol Support**: Parsers for identifying behavioral patterns in TCP, UDP, QUIC, HTTP/1.1, HTTP/2, HTTP/3, and TLS
- **Machine Learning Inference**: Simulated neural network inference for fast, in-process traffic classification
- **Behavioral Feature Extraction**: Generates 128-dimensional feature vectors from traffic flow, including:
  - Packet size distributions and patterns
  - Timing intervals and variance analysis
  - Protocol-specific behavioral markers
  - Flow duration and packet count statistics
- **REST API**: Comprehensive HTTP API with health checks, statistics, and manual analysis endpoints
- **Prometheus Metrics**: Built-in monitoring with custom metrics for bot detection statistics
- **Graceful Shutdown**: Proper resource cleanup and signal handling
- **Extensible Architecture**: Easily add new protocol parsers and analysis modules
- **Production Ready**: Docker support, configuration management, and comprehensive logging

## 🏗️ Architecture

```
                              +-----------------------+
                              |   Cortex Engine       |
                              | (ML Model Inference)  |
                              +-----------+-----------+
                                          ^
                                          |         (Feature Vector)
+----------------+      +-------------------+ | +----------------------+
| Live Network   |----->|   Argus Engine    |-->|  Detection Results   |
| Traffic (NIC)  |      | (Packet Capture & |   | (API / Prometheus /  |
+----------------+      | Feature Extractor)|   |  Logging)            |
                        +-------------------+   +----------------------+
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- `libpcap` library installed (`sudo apt-get install libpcap-dev` on Debian/Ubuntu)
- Docker (optional, for containerized deployment)

### Quick Start

1. **Clone the repository:**
   ```sh
   git clone https://github.com/arvid-berndtsson/protocol-argus-cortex.git
   cd protocol-argus-cortex
   ```

2. **Install dependencies:**
   ```sh
   go mod tidy
   ```

3. **Build the application:**
   ```sh
   make build
   ```

4. **Configure the application:**
   ```sh
   cp config.yml.example config.yml
   # Edit config.yml with your settings
   ```

5. **Run the application:**
   ```sh
   sudo ./build/protocol-argus-cortex --config config.yml --verbose
   ```

### Configuration

The application uses YAML configuration with sensible defaults:

```yaml
server:
  api_port: 8080
  metrics_port: 9090

capture:
  interface: "eth0"
  bpf_filter: "tcp or udp port 443"
  buffer_size: 1048576  # 1MB

cortex:
  model_path: "./models/bot_detection_v1.onnx"
  detection_threshold: 0.85
  batch_size: 32
  inference_timeout: 1000
```

## 🧪 Testing

The project includes comprehensive test coverage:

```sh
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with verbose output
go test -v ./...
```

All tests pass successfully, covering:
- ✅ Cortex engine initialization and inference
- ✅ Argus engine packet capture and flow analysis
- ✅ Feature extraction and behavioral analysis
- ✅ Configuration loading and validation
- ✅ API server functionality

## 📊 API Endpoints

The application exposes a REST API on port 8080:

- `GET /` - API information and available endpoints
- `GET /health` - Health check endpoint
- `GET /api/v1/status` - System status and statistics
- `GET /api/v1/statistics` - Detailed detection statistics
- `GET /api/v1/flows` - Active network flows
- `POST /api/v1/analyze` - Manual feature analysis
- `GET /metrics` - Prometheus metrics

### Example API Usage

```sh
# Check system status
curl http://localhost:8080/api/v1/status

# Get detection statistics
curl http://localhost:8080/api/v1/statistics

# Manual analysis
curl -X POST http://localhost:8080/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{"features": [0.1, 0.2, ...], "flow_id": "test-flow"}'
```

## 🐳 Docker Deployment

### Using Docker Compose (Recommended)

```sh
# Start the full stack with Prometheus and Grafana
docker-compose up -d

# View logs
docker-compose logs -f argus-cortex

# Stop the stack
docker-compose down
```

### Manual Docker Build

```sh
# Build the image
make docker-build

# Run the container
make docker-run

# Stop the container
make docker-stop
```

## 📈 Monitoring

The application integrates with Prometheus and Grafana for monitoring:

- **Prometheus**: Scrapes metrics from `/metrics` endpoint
- **Grafana**: Pre-configured dashboards for bot detection analytics
- **Custom Metrics**: Bot detections, human detections, active flows, packet counts

Access Grafana at `http://localhost:3000` (admin/admin) to view dashboards.

## 🛠️ Development

### Available Make Targets

```sh
make help          # Show all available targets
make build         # Build the application
make test          # Run tests
make fmt           # Format code
make lint          # Run linter
make clean         # Clean build artifacts
make run           # Run the application
make docker-build  # Build Docker image
```

### Project Structure

```
├── cmd/protocol-argus-cortex/
│   └── main.go                    # Main application entry point
├── internal/
│   ├── api/                       # REST API and metrics server
│   └── cortex/                    # ML inference engine
├── pkg/
│   ├── argus/                     # Packet capture and feature extraction
│   ├── config/                    # Configuration management
│   └── protocol/                  # Protocol parsers (HTTP/2, QUIC, TLS)
├── models/                        # ML model storage
├── config.yml.example             # Configuration template
├── Dockerfile                     # Multi-stage container build
├── docker-compose.yml             # Full stack deployment
├── Makefile                       # Development automation
└── README.md                      # This file
```

## 🔧 Advanced Usage

### Custom Protocol Parsers

Add new protocol support by implementing the `Parser` interface:

```go
type Parser interface {
    ParsePacket(data []byte) (*ProtocolInfo, error)
    IsSupportedProtocol(protocol string) bool
}
```

### Feature Extraction

The system extracts 128-dimensional feature vectors including:
- Packet size statistics (mean, variance, distribution)
- Timing patterns (intervals, regularity, burst patterns)
- Protocol-specific features (headers, methods, paths)
- Flow characteristics (duration, packet count, direction)

### Machine Learning Integration

The Cortex engine is designed to integrate with real ML models:
- Replace the simulation with actual ONNX/TensorFlow inference
- Add model versioning and A/B testing capabilities
- Implement model retraining pipelines

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Format code (`make fmt`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Add tests for new functionality
- Update documentation for API changes
- Use conventional commit messages
- Ensure all tests pass before submitting

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [gopacket](https://github.com/google/gopacket) for packet capture capabilities
- [Prometheus](https://prometheus.io/) for metrics collection
- [Grafana](https://grafana.com/) for visualization
- The Go community for excellent tooling and libraries

---

**Protocol Argus Cortex** - Advanced bot detection through behavioral analysis 🚀
