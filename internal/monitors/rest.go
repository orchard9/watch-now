package monitors

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/orchard9/watch-now/internal/config"
)

type RESTMonitor struct {
	name    string
	url     string
	health  string
	timeout time.Duration
	headers map[string]string
}

func NewRESTMonitor(cfg config.ServiceConfig) *RESTMonitor {
	healthPath := cfg.Health
	if healthPath == "" {
		healthPath = "/health"
	}

	return &RESTMonitor{
		name:    cfg.Name,
		url:     cfg.URL,
		health:  healthPath,
		timeout: cfg.Timeout,
		headers: cfg.Headers,
	}
}

func (m *RESTMonitor) Name() string {
	return m.name
}

func (m *RESTMonitor) Type() MonitorType {
	return TypeREST
}

func (m *RESTMonitor) Check(ctx context.Context) (*Result, error) {
	start := time.Now()

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	// Build full URL
	fullURL := m.url + m.health

	// Create request
	req, err := http.NewRequestWithContext(checkCtx, "GET", fullURL, nil)
	if err != nil {
		return &Result{
			Name:      m.name,
			Type:      TypeREST,
			Status:    StatusFail,
			Message:   fmt.Sprintf("Failed to create request: %v", err),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}

	// Add headers
	for key, value := range m.headers {
		req.Header.Set(key, value)
	}

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	duration := time.Since(start)

	result := &Result{
		Name:      m.name,
		Type:      TypeREST,
		Timestamp: time.Now(),
		Duration:  duration,
		Metadata:  make(map[string]interface{}),
	}

	// Add request info to metadata
	result.Metadata["url"] = fullURL
	result.Metadata["timeout"] = m.timeout.String()

	if err != nil {
		// Check if it was a timeout
		if checkCtx.Err() == context.DeadlineExceeded {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Request timed out after %v", m.timeout)
			return result, nil
		}

		// Request failed
		result.Status = StatusFail
		result.Message = fmt.Sprintf("Request failed: %v", err)
		return result, nil
	}

	defer resp.Body.Close()

	// Add response info to metadata
	result.Metadata["status_code"] = resp.StatusCode

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Status = StatusOK
		result.Message = fmt.Sprintf("HTTP %d in %v", resp.StatusCode, duration.Round(time.Millisecond))
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("HTTP %d (client error) in %v", resp.StatusCode, duration.Round(time.Millisecond))
	} else {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("HTTP %d (server error) in %v", resp.StatusCode, duration.Round(time.Millisecond))
	}

	return result, nil
}
