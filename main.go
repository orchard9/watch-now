package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/orchard9/watch-now/internal/api"
	"github.com/orchard9/watch-now/internal/config"
	"github.com/orchard9/watch-now/internal/core"
	"github.com/orchard9/watch-now/internal/detector"
	"github.com/orchard9/watch-now/internal/monitors"
)

// Version information
var (
	version = "0.1.0"
	commit  = "dev"
	date    = "unknown"
)

// Color helpers
var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	blue   = color.New(color.FgBlue)
	bold   = color.New(color.Bold)
)

func main() {
	// Command line flags
	showVersion := flag.Bool("version", false, "Show version information")
	runOnce := flag.Bool("once", false, "Run once and exit")
	configPath := flag.String("config", ".watch-now.yaml", "Path to configuration file")
	initConfig := flag.Bool("init", false, "Generate a configuration file for the current project")
	port := flag.Int("port", 0, "Port for REST API (0 for ephemeral port)")
	showExamples := flag.Bool("show-examples", false, "Show example configurations")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "watch-now is a universal development monitor for code quality and service health.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --init                    Generate configuration for current project\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --once                    Run monitoring once and exit\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --config custom.yaml      Use custom configuration file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --port 8080               Set API port (enables API)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s                           Start continuous monitoring\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nConfiguration File Format (.watch-now.yaml):\n")
		fmt.Fprintf(os.Stderr, "  services:                      # Service health monitoring\n")
		fmt.Fprintf(os.Stderr, "    - name: api-server           # Service name\n")
		fmt.Fprintf(os.Stderr, "      type: rest                 # Service type (rest/grpc)\n")
		fmt.Fprintf(os.Stderr, "      url: http://localhost:8080 # Service URL\n")
		fmt.Fprintf(os.Stderr, "      health: /health            # Health endpoint path\n")
		fmt.Fprintf(os.Stderr, "      timeout: 5s                # Request timeout\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  checks:                        # Code quality checks\n")
		fmt.Fprintf(os.Stderr, "    - name: lint                 # Check name\n")
		fmt.Fprintf(os.Stderr, "      command: make              # Command to run\n")
		fmt.Fprintf(os.Stderr, "      args: [\"lint\"]             # Command arguments\n")
		fmt.Fprintf(os.Stderr, "      timeout: 60s               # Check timeout\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  interval: 30s                  # Monitoring interval\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  api:                           # REST API configuration\n")
		fmt.Fprintf(os.Stderr, "    enabled: true                # Enable/disable API\n")
		fmt.Fprintf(os.Stderr, "    port: 0                      # API port (0 = ephemeral)\n")
		fmt.Fprintf(os.Stderr, "\nExample Configuration:\n")
		fmt.Fprintf(os.Stderr, "  # Monitoring a Go microservice project\n")
		fmt.Fprintf(os.Stderr, "  services:\n")
		fmt.Fprintf(os.Stderr, "    - name: api\n")
		fmt.Fprintf(os.Stderr, "      type: rest\n")
		fmt.Fprintf(os.Stderr, "      url: http://localhost:8080\n")
		fmt.Fprintf(os.Stderr, "      health: /health\n")
		fmt.Fprintf(os.Stderr, "      timeout: 5s\n")
		fmt.Fprintf(os.Stderr, "    - name: worker\n")
		fmt.Fprintf(os.Stderr, "      type: rest\n")
		fmt.Fprintf(os.Stderr, "      url: http://localhost:8081\n")
		fmt.Fprintf(os.Stderr, "      health: /healthz\n")
		fmt.Fprintf(os.Stderr, "      timeout: 5s\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  checks:\n")
		fmt.Fprintf(os.Stderr, "    - name: test\n")
		fmt.Fprintf(os.Stderr, "      command: go\n")
		fmt.Fprintf(os.Stderr, "      args: [\"test\", \"./...\"]\n")
		fmt.Fprintf(os.Stderr, "      timeout: 120s\n")
		fmt.Fprintf(os.Stderr, "    - name: lint\n")
		fmt.Fprintf(os.Stderr, "      command: golangci-lint\n")
		fmt.Fprintf(os.Stderr, "      args: [\"run\"]\n")
		fmt.Fprintf(os.Stderr, "      timeout: 60s\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  interval: 30s\n")
		fmt.Fprintf(os.Stderr, "  api:\n")
		fmt.Fprintf(os.Stderr, "    enabled: true\n")
		fmt.Fprintf(os.Stderr, "    port: 9090\n")
		fmt.Fprintf(os.Stderr, "\nUse --show-examples to see more configuration examples.\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("watch-now %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	if *showExamples {
		showExampleConfigurations()
		os.Exit(0)
	}

	if *initConfig {
		generateConfig(*configPath)
		return
	}

	// Load configuration and initialize engine
	engine, cfg := initializeEngine(*configPath)

	// Override API port if specified via flag
	if *port != 0 {
		cfg.API.Port = *port
		cfg.API.Enabled = true
	}

	// Print header
	printHeader()

	// Set up context for graceful shutdown
	ctx := setupGracefulShutdown()

	if *runOnce {
		runOnceMode(ctx, engine)
	} else {
		runContinuousMode(ctx, engine, cfg)
	}
}

func initializeEngine(configPath string) (*core.Engine, *config.Config) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	engine := core.NewEngine(cfg)
	if err := engine.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing engine: %v\n", err)
		os.Exit(1)
	}

	return engine, cfg
}

func printHeader() {
	fmt.Println(bold.Sprint("watch-now - Universal Development Monitor"))
	fmt.Println("================================================================================")
}

func setupGracefulShutdown() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	return ctx
}

func runOnceMode(ctx context.Context, engine *core.Engine) {
	// Start engine
	go func() {
		if err := engine.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Engine error: %v\n", err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // Give monitors time to start

	// Wait for initial results
	// Increased timeout to 60s to accommodate sequential golangci-lint execution
	// (5 services × ~10s per lint check = ~50s + 10s buffer)
	waitForResults(engine, 60*time.Second)
	runMonitor(engine)

	// Exit with appropriate code
	status := getOverallStatus(engine.State().GetAll())
	if status == monitors.StatusFail {
		os.Exit(1)
	}
}

func runContinuousMode(ctx context.Context, engine *core.Engine, cfg *config.Config) {
	fmt.Printf("Monitoring every %v. Press Ctrl+C to stop.\n", cfg.Interval)

	// Start API server if needed
	var apiServer *api.Server
	if cfg.API.Enabled {
		apiServer = api.NewServer(engine, cfg.API.Port)
		go func() {
			if err := apiServer.Start(); err != nil {
				log.Printf("API server error: %v", err)
			}
		}()
		fmt.Printf("API enabled at http://localhost:%d\n", apiServer.Port())
		fmt.Printf("  Status: http://localhost:%d/api/status\n", apiServer.Port())
		fmt.Printf("  Events: http://localhost:%d/api/events\n", apiServer.Port())
	}
	fmt.Println("================================================================================")

	// Start monitoring in background
	go func() {
		if err := engine.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Engine error: %v\n", err)
		}
	}()

	// Wait for initial results before first display
	waitForResults(engine, 10*time.Second)

	// Initial display with results
	runMonitor(engine)

	// Display results periodically
	ticker := time.NewTicker(5 * time.Second) // Update display every 5 seconds
	defer ticker.Stop()

	// Clean up API server on exit
	defer func() {
		if apiServer != nil {
			_ = apiServer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			clearScreen()
			runMonitor(engine)
		}
	}
}

func runMonitor(engine *core.Engine) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("\n%s System Status\n", bold.Sprintf("[%s]", timestamp))
	fmt.Println("--------------------------------------------------------------------------------")

	// Get all results from state
	results := engine.State().GetAll()

	// Group results by type
	var qualityResults []*monitors.Result
	var serviceResults []*monitors.Result

	for _, result := range results {
		switch result.Type {
		case monitors.TypeQuality:
			qualityResults = append(qualityResults, result)
		case monitors.TypeREST, monitors.TypeGRPC:
			serviceResults = append(serviceResults, result)
		}
	}

	sort.Slice(serviceResults, func(i, j int) bool {
		return strings.ToLower(serviceResults[i].Name) < strings.ToLower(serviceResults[j].Name)
	})
	sort.Slice(qualityResults, func(i, j int) bool {
		return strings.ToLower(qualityResults[i].Name) < strings.ToLower(qualityResults[j].Name)
	})

	// Display services
	if len(serviceResults) > 0 {
		fmt.Printf("\n%s Services:\n", blue.Sprint("SERVICES"))
		for _, result := range serviceResults {
			displayResult(result)
		}
	} else {
		fmt.Printf("\n%s Services:\n", blue.Sprint("SERVICES"))
		fmt.Printf("  %s No services configured\n", yellow.Sprint("[INFO]"))
	}

	// Display code quality
	if len(qualityResults) > 0 {
		fmt.Printf("\n%s Code Quality:\n", blue.Sprint("CHECKS"))
		for _, result := range qualityResults {
			displayResult(result)
		}
	} else {
		fmt.Printf("\n%s Code Quality:\n", blue.Sprint("CHECKS"))
		fmt.Printf("  %s No checks configured\n", yellow.Sprint("[INFO]"))
	}

	// Overall status
	status := getOverallStatus(results)
	statusColor := green
	statusText := "All systems operational"

	switch status {
	case monitors.StatusWarn:
		statusColor = yellow
		statusText = "Some checks need attention"
	case monitors.StatusFail:
		statusColor = red
		statusText = "Some checks are failing"
	}

	fmt.Printf("\n%s %s\n", statusColor.Sprintf("[%s]", strings.ToUpper(string(status))), bold.Sprint("STATUS: "+statusText))
	fmt.Println("================================================================================")
}

func displayResult(result *monitors.Result) {
	var statusColor *color.Color
	var statusText string

	switch result.Status {
	case monitors.StatusOK:
		statusColor = green
		statusText = "OK"
	case monitors.StatusWarn:
		statusColor = yellow
		statusText = "WARN"
	case monitors.StatusFail:
		statusColor = red
		statusText = "FAIL"
	case monitors.StatusInfo:
		statusColor = blue
		statusText = "INFO"
	}

	message := result.Message
	if (result.Type == monitors.TypeREST || result.Type == monitors.TypeGRPC) && result.Metadata != nil {
		if urlValue, ok := result.Metadata["url"].(string); ok && urlValue != "" {
			message = fmt.Sprintf("%s @ %s", message, urlValue)
		}
	}

	fmt.Printf("  %s %s - %s\n",
		statusColor.Sprintf("[%s]", statusText),
		result.Name,
		message)
}

func getOverallStatus(results map[string]*monitors.Result) monitors.Status {
	if len(results) == 0 {
		return monitors.StatusInfo
	}

	hasWarn := false
	for _, result := range results {
		if result.Status == monitors.StatusFail {
			return monitors.StatusFail
		}
		if result.Status == monitors.StatusWarn {
			hasWarn = true
		}
	}

	if hasWarn {
		return monitors.StatusWarn
	}
	return monitors.StatusOK
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func waitForResults(engine *core.Engine, timeout time.Duration) {
	expectedCount := engine.MonitorCount()
	if expectedCount == 0 {
		return // No monitors to wait for
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if time.Now().After(deadline) {
			return
		}

		// Check if we have results from all monitors
		results := engine.State().GetAll()
		if len(results) >= expectedCount {
			// Wait a bit more to ensure all monitors complete
			time.Sleep(200 * time.Millisecond)
			return
		}
	}
}

func generateConfig(configPath string) {
	fmt.Println(bold.Sprint("watch-now Configuration Generator"))
	fmt.Println("================================================================================")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("%s Configuration file %s already exists.\n", yellow.Sprint("WARNING:"), configPath)
		fmt.Print("Overwrite? (y/N): ")

		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Configuration generation cancelled.")
			return
		}
	}

	// Analyze current project
	fmt.Printf("Analyzing project in %s...\n", getCurrentDir())

	d := detector.NewProjectDetector(".")
	projectInfo, err := d.DetectProject()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing project: %v\n", err)
		os.Exit(1)
	}

	// Generate configuration
	cfg := d.GenerateConfig()

	// Create YAML content with comments
	yamlContent := createYAMLWithComments(projectInfo, cfg)

	// Write to file
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing configuration file: %v\n", err)
		os.Exit(1)
	}

	// Show summary
	fmt.Printf("\n%s Configuration generated: %s\n", green.Sprint("✓"), configPath)
	fmt.Printf("Project type: %s\n", projectInfo.Type)
	fmt.Printf("Services detected: %d\n", len(projectInfo.Services))
	fmt.Printf("Quality checks: %d\n", len(projectInfo.QualityChecks))

	if len(projectInfo.Services) > 0 {
		fmt.Printf("\nDetected services:\n")
		for _, service := range projectInfo.Services {
			fmt.Printf("  - %s (%s%s)\n", service.Name, service.URL, service.Health)
		}
	}

	if len(projectInfo.QualityChecks) > 0 {
		fmt.Printf("\nQuality checks:\n")
		for _, check := range projectInfo.QualityChecks {
			fmt.Printf("  - %s: %s %s\n", check.Name, check.Command, strings.Join(check.Args, " "))
		}
	}

	fmt.Printf("\n%s Run 'watch-now --once' to test your configuration\n", blue.Sprint("TIP:"))
}

func createYAMLWithComments(projectInfo *detector.ProjectInfo, cfg *config.Config) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# watch-now configuration for %s project\n", projectInfo.Type))
	sb.WriteString("# Generated automatically - customize as needed\n\n")

	if len(cfg.Services) > 0 {
		sb.WriteString("# Service health monitoring\n")
		sb.WriteString("services:\n")
		for _, service := range cfg.Services {
			sb.WriteString(fmt.Sprintf("  - name: %s\n", service.Name))
			sb.WriteString("    type: rest\n")
			sb.WriteString(fmt.Sprintf("    url: %s\n", service.URL))
			sb.WriteString(fmt.Sprintf("    health: %s\n", service.Health))
			sb.WriteString(fmt.Sprintf("    timeout: %ds\n", int(service.Timeout.Seconds())))
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("# No services detected - add them manually if needed\n")
		sb.WriteString("services: []\n\n")
	}

	if len(cfg.Checks) > 0 {
		sb.WriteString("# Code quality checks\n")
		sb.WriteString("checks:\n")
		for _, check := range cfg.Checks {
			sb.WriteString(fmt.Sprintf("  - name: %s\n", check.Name))
			sb.WriteString(fmt.Sprintf("    command: %s\n", check.Command))
			if len(check.Args) > 0 {
				sb.WriteString("    args: [")
				for i, arg := range check.Args {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf(`"%s"`, arg))
				}
				sb.WriteString("]\n")
			}
			sb.WriteString(fmt.Sprintf("    timeout: %ds\n", int(check.Timeout.Seconds())))
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("# No quality checks detected - add them manually\n")
		sb.WriteString("checks: []\n\n")
	}

	sb.WriteString("# Monitoring interval\n")
	sb.WriteString(fmt.Sprintf("interval: %ds\n\n", int(cfg.Interval.Seconds())))

	sb.WriteString("# REST API and SSE for web UI integration\n")
	sb.WriteString("api:\n")
	sb.WriteString(fmt.Sprintf("  enabled: %t\n", cfg.API.Enabled))
	sb.WriteString(fmt.Sprintf("  port: %d  # 0 = ephemeral port\n", cfg.API.Port))

	return sb.String()
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func showExampleConfigurations() {
	fmt.Println(bold.Sprint("watch-now Example Configurations"))
	fmt.Println("================================================================================")

	fmt.Println("\n" + blue.Sprint("1. Simple Go Project"))
	fmt.Println("------------------------------")
	fmt.Println(`# Basic Go project with standard make targets
checks:
  - name: test
    command: go
    args: ["test", "./..."]
    timeout: 120s
  - name: build
    command: go
    args: ["build", "./..."]
    timeout: 180s
  - name: fmt
    command: gofmt
    args: ["-l", "."]
    timeout: 30s

interval: 60s`)

	fmt.Println("\n" + blue.Sprint("2. Node.js Full Stack Application"))
	fmt.Println("------------------------------")
	fmt.Println(`# Node.js app with frontend and backend services
services:
  - name: backend
    type: rest
    url: http://localhost:3000
    health: /api/health
    timeout: 5s
  - name: frontend
    type: rest
    url: http://localhost:3001
    health: /
    timeout: 5s

checks:
  - name: lint
    command: npm
    args: ["run", "lint"]
    timeout: 60s
  - name: test
    command: npm
    args: ["test"]
    timeout: 120s
  - name: typecheck
    command: npm
    args: ["run", "typecheck"]
    timeout: 60s

interval: 30s
api:
  enabled: true
  port: 9090`)

	fmt.Println("\n" + blue.Sprint("3. Microservices with Docker"))
	fmt.Println("------------------------------")
	fmt.Println(`# Multiple services with infrastructure checks
services:
  - name: auth-service
    type: rest
    url: http://localhost:8080
    health: /health
    timeout: 5s
    headers:
      Authorization: "Bearer dev-token"
  - name: user-service
    type: rest
    url: http://localhost:8081
    health: /health
    timeout: 5s
  - name: payment-service
    type: rest
    url: http://localhost:8082
    health: /healthz
    timeout: 10s

checks:
  - name: docker-compose
    command: docker-compose
    args: ["ps"]
    timeout: 10s
  - name: integration-tests
    command: make
    args: ["test-integration"]
    timeout: 300s

interval: 60s
api:
  enabled: true
  port: 0  # Use ephemeral port`)

	fmt.Println("\n" + blue.Sprint("4. Python Data Science Project"))
	fmt.Println("------------------------------")
	fmt.Println(`# Python project with notebook and API monitoring
services:
  - name: jupyter
    type: rest
    url: http://localhost:8888
    health: /api
    timeout: 5s
  - name: model-api
    type: rest
    url: http://localhost:5000
    health: /health
    timeout: 5s

checks:
  - name: pytest
    command: pytest
    args: ["-v"]
    timeout: 180s
  - name: mypy
    command: mypy
    args: ["."]
    timeout: 60s
  - name: black
    command: black
    args: ["--check", "."]
    timeout: 30s
  - name: notebook-test
    command: pytest
    args: ["--nbval", "notebooks/"]
    timeout: 300s

interval: 60s`)

	fmt.Println("\n" + blue.Sprint("5. Monorepo with Multiple Services"))
	fmt.Println("------------------------------")
	fmt.Println(`# Large monorepo with service discovery
services:
  - name: api-gateway
    type: rest
    url: http://localhost:8000
    health: /health
    timeout: 5s
  - name: service-a
    type: rest
    url: http://localhost:8001
    health: /health
    timeout: 5s
  - name: service-b
    type: rest
    url: http://localhost:8002
    health: /health
    timeout: 5s
  - name: service-c
    type: grpc
    url: localhost:9001
    health: /grpc.health.v1.Health/Check
    timeout: 5s

checks:
  - name: lint-all
    command: make
    args: ["lint"]
    timeout: 180s
  - name: test-unit
    command: make
    args: ["test-unit"]
    timeout: 300s
  - name: test-integration
    command: make
    args: ["test-integration"]
    timeout: 600s
  - name: build-all
    command: make
    args: ["build"]
    timeout: 300s
  - name: security-scan
    command: make
    args: ["security-scan"]
    timeout: 120s

interval: 120s
api:
  enabled: true
  port: 9090`)

	fmt.Println("\n================================================================================")
	fmt.Println("To use any example:")
	fmt.Println("1. Copy the configuration to .watch-now.yaml")
	fmt.Println("2. Adjust ports, commands, and timeouts for your project")
	fmt.Println("3. Run 'watch-now' to start monitoring")
	fmt.Println("\nOr use 'watch-now --init' to auto-generate a configuration for your project.")
}
