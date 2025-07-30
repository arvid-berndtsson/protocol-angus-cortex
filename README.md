# Protocol Argus Cortex

[![Go Report Card](https://goreportcard.com/badge/github.com/your-username/protocol-argus-cortex)](https://goreportcard.com/report/github.com/your-username/protocol-argus-cortex)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/your-username/protocol-argus-cortex/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

An advanced, real-time network traffic analysis engine using a neural network to detect and classify bot activity over modern internet protocols like HTTP/2, HTTP/3, and QUIC.

## Overview

The digital landscape is rife with sophisticated bots that can evade traditional detection methods based on simple signatures or IP blacklists. **Protocol Argus Cortex** is designed to address this challenge by operating at a deeper level.

*   **Argus Engine**: The "all-seeing" eye that captures and performs deep inspection of network packets in real-time. It extracts a rich set of behavioral features and metadata, rather than just inspecting payloads.
*   **Cortex Engine**: The "brain" of the operation. It feeds the features extracted by Argus into a pre-trained neural network to classify traffic as human or bot, providing a confidence score and reasoning for its verdict.

This project focuses on the *behavioral fingerprint* of a connection—how it communicates, not just what it says.

## Core Features

*   **Real-time Packet Capture**: High-performance packet capture using `gopacket`.
*   **Advanced Protocol Support**: Parsers for identifying behavioral patterns in TCP, UDP, QUIC, HTTP/2, and HTTP/3.
*   **Neural Network Inference**: Go bindings for a TensorFlow/ONNX model to perform fast, in-process traffic classification.
*   **Behavioral Feature Extraction**: Generates feature vectors from traffic flow, including packet size, timing, TLS handshake parameters, and protocol-specific state transitions.
*   **Extensible Architecture**: Easily add new protocol parsers and analysis modules.
*   **Metrics & API**: Exposes Prometheus metrics and a simple REST API for querying detection statistics.

## Architecture
```
                                 +-----------------------+
                              |   Cortex Engine       |
                              | (NN Model Inference)  |
                              +-----------+-----------+
                                          ^
                                          | (Feature Vector)
 +----------------+      +-------------------+ | +----------------------+
| Live Network   |----->|   Argus Engine    |-->|  Detection Results   |
| Traffic (NIC)  |      | (Packet Capture & |   | (API / Prometheus /  |
+----------------+      | Feature Extractor)|   |  Logging)            |
+-------------------+   +----------------------+
```
## Getting Started

### Prerequisites

*   Go 1.21+
*   `libpcap` library installed (`sudo apt-get install libpcap-dev` on Debian/Ubuntu)
*   A pre-trained model file (e.g., `model.onnx`)

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/your-username/protocol-argus-cortex.git
    cd protocol-argus-cortex
    ```

2.  **Install dependencies:**
    ```sh
    go mod tidy
    ```

3.  **Build the application:**
    ```sh
    go build -o protocol-argus-cortex ./cmd/protocol-argus-cortex/
    ```

### Configuration

Copy the example configuration and modify it for your environment.

```sh
cp config.yml.example config.yml
 ⁠config.yml:
 server:
  api_port: 8080
  metrics_port: 9090

capture:
  # The network interface to monitor (e.g., eth0, en0)
  interface: "eth0"
  # BPF filter to capture specific traffic
  bpf_filter: "tcp or udp port 443"

cortex:
  # Path to the trained neural network model
  model_path: "./models/bot_detection_v1.onnx"
  # The minimum confidence score to trigger an alert
  detection_threshold: 0.85
 Running the Engine
```
You need elevated privileges to capture network packets.
```sh
sudo ./protocol-argus-cortex --config config.yml
```

### Project Structure
```
├── cmd/protocol-argus-cortex/
│   └── main.go               # Main application entry point
├── internal/
│   ├── api/                  # REST API and metrics server
│   └── cortex/               # Neural network loading and inference logic
├── pkg/
│   ├── argus/                # Packet capture and feature extraction
│   ├── config/               # Configuration loading
│   └── protocol/             # Protocol-specific parsers (HTTP2, QUIC)
├── models/
│   └── .gitkeep              # Placeholder for model files
├── config.yml.example        # Example configuration
├── go.mod
├── go.sum
└── README.md
```
### Contribution
Contributions are welcome! Please open an issue to discuss your ideas or submit a pull request. Ensure your code is formatted with ⁠go fmt before submitting.
