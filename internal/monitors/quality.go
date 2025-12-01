package monitors

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/orchard9/watch-now/internal/config"
)

// Global mutex to prevent concurrent golangci-lint execution
// golangci-lint uses file-based locking and fails with exit code 2
// when multiple instances try to run simultaneously
var golangciLintMutex sync.Mutex

type QualityMonitor struct {
	name    string
	command string
	args    []string
	timeout time.Duration
}

func NewQualityMonitor(cfg config.CheckConfig) *QualityMonitor {
	return &QualityMonitor{
		name:    cfg.Name,
		command: cfg.Command,
		args:    cfg.Args,
		timeout: cfg.Timeout,
	}
}

func (m *QualityMonitor) Name() string {
	return m.name
}

func (m *QualityMonitor) Type() MonitorType {
	return TypeQuality
}

// isGolangciLint checks if this monitor is running golangci-lint
func (m *QualityMonitor) isGolangciLint() bool {
	// Check if command is golangci-lint
	if strings.Contains(m.command, "golangci-lint") {
		return true
	}
	// Check if any args contain golangci-lint (e.g., via make)
	for _, arg := range m.args {
		if strings.Contains(arg, "lint") {
			// This is a lint command via make - assume it runs golangci-lint
			return true
		}
	}
	return false
}

func (m *QualityMonitor) Check(ctx context.Context) (*Result, error) {
	start := time.Now()

	// Serialize golangci-lint execution to prevent file lock contention
	// golangci-lint uses file-based locking and fails when run concurrently
	if m.isGolangciLint() {
		golangciLintMutex.Lock()
		defer golangciLintMutex.Unlock()
	}

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(checkCtx, m.command, m.args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()
	duration := time.Since(start)

	result := &Result{
		Name:      m.name,
		Type:      TypeQuality,
		Timestamp: time.Now(),
		Duration:  duration,
		Metadata:  make(map[string]interface{}),
	}

	// Add command info to metadata
	result.Metadata["command"] = fmt.Sprintf("%s %s", m.command, strings.Join(m.args, " "))

	if err != nil {
		// Check if it was a timeout
		if checkCtx.Err() == context.DeadlineExceeded {
			result.Status = StatusFail
			result.Message = fmt.Sprintf("Command timed out after %v", m.timeout)
			return result, nil
		}

		// Command failed
		result.Status = StatusFail
		result.Message = fmt.Sprintf("Command failed: %v", err)

		// Include stderr in metadata if available
		if stderr.Len() > 0 {
			result.Metadata["stderr"] = stderr.String()
		}

		// Check exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.Metadata["exit_code"] = exitErr.ExitCode()
		}

		return result, nil
	}

	// Command succeeded
	result.Status = StatusOK
	result.Message = fmt.Sprintf("Check passed in %v", duration.Round(time.Millisecond))

	// Include stdout if it's not too large
	if stdout.Len() > 0 && stdout.Len() < 1024 {
		result.Metadata["output"] = stdout.String()
	}

	return result, nil
}
