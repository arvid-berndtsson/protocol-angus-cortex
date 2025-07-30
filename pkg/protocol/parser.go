package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// Parser represents a protocol parser
type Parser struct {
	supportedProtocols map[string]bool
}

// NewParser creates a new protocol parser
func NewParser() *Parser {
	return &Parser{
		supportedProtocols: map[string]bool{
			"HTTP/1.1": true,
			"HTTP/2":   true,
			"HTTP/3":   true,
			"QUIC":     true,
			"TLS":      true,
		},
	}
}

// ParsePacket attempts to parse a packet and extract protocol information
func (p *Parser) ParsePacket(data []byte) (*ProtocolInfo, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("packet too small to parse")
	}

	info := &ProtocolInfo{
		RawData: data,
	}

	// Try to identify the protocol
	protocol, err := p.identifyProtocol(data)
	if err != nil {
		return nil, err
	}

	info.Protocol = protocol

	// Parse based on protocol type
	switch protocol {
	case "HTTP/1.1":
		return p.parseHTTP11(data, info)
	case "HTTP/2":
		return p.parseHTTP2(data, info)
	case "HTTP/3":
		return p.parseHTTP3(data, info)
	case "QUIC":
		return p.parseQUIC(data, info)
	case "TLS":
		return p.parseTLS(data, info)
	default:
		return info, nil
	}
}

// ProtocolInfo contains parsed protocol information
type ProtocolInfo struct {
	Protocol   string                 `json:"protocol"`
	Version    string                 `json:"version"`
	Headers    map[string]string      `json:"headers"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	RawData    []byte                 `json:"-"`
	Features   map[string]interface{} `json:"features"`
}

// identifyProtocol attempts to identify the protocol from packet data
func (p *Parser) identifyProtocol(data []byte) (string, error) {
	// Check for TLS handshake
	if len(data) >= 5 && data[0] == 0x16 {
		return "TLS", nil
	}

	// Check for HTTP/1.1
	if bytes.HasPrefix(data, []byte("GET ")) ||
		bytes.HasPrefix(data, []byte("POST ")) ||
		bytes.HasPrefix(data, []byte("HTTP/1.1")) {
		return "HTTP/1.1", nil
	}

	// Check for HTTP/2 preface
	if bytes.HasPrefix(data, []byte("PRI * HTTP/2.0")) {
		return "HTTP/2", nil
	}

	// Check for QUIC (simplified)
	if len(data) >= 4 && (data[0]&0xC0) == 0x40 {
		return "QUIC", nil
	}

	// Check for HTTP/3 (over QUIC)
	if len(data) >= 8 && (data[0]&0xC0) == 0x40 {
		// This is a simplified check - real HTTP/3 detection is more complex
		return "HTTP/3", nil
	}

	return "Unknown", nil
}

// parseHTTP11 parses HTTP/1.1 packets
func (p *Parser) parseHTTP11(data []byte, info *ProtocolInfo) (*ProtocolInfo, error) {
	lines := strings.Split(string(data), "\r\n")
	if len(lines) < 2 {
		return info, fmt.Errorf("invalid HTTP/1.1 format")
	}

	// Parse first line
	firstLine := lines[0]
	parts := strings.Fields(firstLine)
	if len(parts) < 2 {
		return info, fmt.Errorf("invalid HTTP/1.1 first line")
	}

	if strings.HasPrefix(firstLine, "HTTP/") {
		// Response
		if len(parts) >= 2 {
			info.StatusCode = 200 // Simplified
			info.Version = parts[0]
		}
	} else {
		// Request
		info.Method = parts[0]
		if len(parts) >= 2 {
			info.Path = parts[1]
		}
		if len(parts) >= 3 {
			info.Version = parts[2]
		}
	}

	// Parse headers
	info.Headers = make(map[string]string)
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			break
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			info.Headers[key] = value

			if strings.EqualFold(key, "User-Agent") {
				info.UserAgent = value
			}
		}
	}

	// Extract features
	info.Features = p.extractHTTP11Features(info)

	return info, nil
}

// parseHTTP2 parses HTTP/2 packets
func (p *Parser) parseHTTP2(data []byte, info *ProtocolInfo) (*ProtocolInfo, error) {
	info.Version = "HTTP/2"

	// HTTP/2 parsing is complex - this is a simplified version
	if len(data) >= 9 {
		frameType := data[3]
		flags := data[4]

		info.Features = map[string]interface{}{
			"frame_type": frameType,
			"flags":      flags,
			"stream_id":  binary.BigEndian.Uint32(data[5:9]),
		}
	}

	return info, nil
}

// parseHTTP3 parses HTTP/3 packets
func (p *Parser) parseHTTP3(data []byte, info *ProtocolInfo) (*ProtocolInfo, error) {
	info.Version = "HTTP/3"

	// HTTP/3 parsing is very complex - this is a simplified version
	info.Features = map[string]interface{}{
		"quic_version": "unknown",
		"stream_type":  "unknown",
	}

	return info, nil
}

// parseQUIC parses QUIC packets
func (p *Parser) parseQUIC(data []byte, info *ProtocolInfo) (*ProtocolInfo, error) {
	info.Version = "QUIC"

	if len(data) >= 1 {
		headerForm := (data[0] & 0x80) >> 7
		info.Features = map[string]interface{}{
			"header_form": headerForm,
			"packet_type": data[0] & 0x7F,
		}
	}

	return info, nil
}

// parseTLS parses TLS packets
func (p *Parser) parseTLS(data []byte, info *ProtocolInfo) (*ProtocolInfo, error) {
	info.Version = "TLS"

	if len(data) >= 5 {
		contentType := data[0]
		version := binary.BigEndian.Uint16(data[1:3])

		info.Features = map[string]interface{}{
			"content_type": contentType,
			"version":      version,
			"length":       binary.BigEndian.Uint16(data[3:5]),
		}
	}

	return info, nil
}

// extractHTTP11Features extracts behavioral features from HTTP/1.1 traffic
func (p *Parser) extractHTTP11Features(info *ProtocolInfo) map[string]interface{} {
	features := make(map[string]interface{})

	// Header count
	features["header_count"] = len(info.Headers)

	// User agent analysis
	if info.UserAgent != "" {
		features["user_agent_length"] = len(info.UserAgent)
		features["has_bot_keywords"] = p.hasBotKeywords(info.UserAgent)
	}

	// Method analysis
	if info.Method != "" {
		features["method"] = info.Method
		features["is_get"] = info.Method == "GET"
		features["is_post"] = info.Method == "POST"
	}

	// Path analysis
	if info.Path != "" {
		features["path_length"] = len(info.Path)
		features["has_query_params"] = strings.Contains(info.Path, "?")
	}

	return features
}

// hasBotKeywords checks if user agent contains bot-related keywords
func (p *Parser) hasBotKeywords(userAgent string) bool {
	botKeywords := []string{
		"bot", "crawler", "spider", "scraper", "automation",
		"headless", "selenium", "phantom", "puppet",
	}

	lowerUA := strings.ToLower(userAgent)
	for _, keyword := range botKeywords {
		if strings.Contains(lowerUA, keyword) {
			return true
		}
	}

	return false
}

// IsSupportedProtocol checks if a protocol is supported
func (p *Parser) IsSupportedProtocol(protocol string) bool {
	return p.supportedProtocols[protocol]
}
