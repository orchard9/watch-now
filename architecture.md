# Architecture

This document describes the internal architecture of watch-now and how it achieves universal project monitoring.

## Design Principles

1. **Easy Configuration**: Simple YAML-based configuration for services and checks
2. **Real-time State**: In-memory state management with API/SSE exposure
3. **External Monitoring**: Focus on monitoring REST/gRPC services and dependencies
4. **Extensibility**: Plugin system for custom checks and monitors
5. **Performance**: Concurrent checks with configurable parallelism
6. **Portability**: Single binary, no runtime dependencies
7. **Clarity**: Clear separation of concerns and modular design

## High-Level Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   CLI Layer     │     │   Web UI        │     │   REST API      │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                         │
         │                       ├─────────────────────────┤
         │                       │    SSE Stream (/sse)    │
         └───────────────────────┴─────────────────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │    Core Engine          │
                    │  ┌──────────────────┐  │
                    │  │ Config Manager   │  │
                    │  ├──────────────────┤  │
                    │  │ State Store      │  │
                    │  ├──────────────────┤  │
                    │  │ Monitor Registry │  │
                    │  ├──────────────────┤  │
                    │  │ Scheduler        │  │
                    │  ├──────────────────┤  │
                    │  │ Result Collector │  │
                    │  ├──────────────────┤  │
                    │  │ API Server       │  │
                    │  └──────────────────┘  │
                    └────────────┬────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │       Monitors          │
                    ├─────────────────────────┤
                    │ REST Service Monitor    │
                    │ gRPC Service Monitor    │
                    │ Code Quality Monitor    │
                    │ Dependency Monitor      │
                    │ Infrastructure Monitor  │
                    └─────────────────────────┘
```

## Core Components

### 1. CLI Layer

Handles command-line interface and user interactions:

```go
// cmd/root.go
type CLI struct {
    engine     *core.Engine
    config     *config.Config
    outputter  output.Outputter
}
```

### 2. Core Engine

Central orchestrator that coordinates all monitoring activities:

```go
// core/engine.go
type Engine struct {
    config      *config.Config
    state       *StateStore
    monitors    *MonitorRegistry
    scheduler   *Scheduler
    collector   *ResultCollector
    apiServer   *APIServer
    plugins     *PluginManager
}
```

### 3. Config Manager

Handles configuration loading and validation:

```go
// config/manager.go
type Manager struct {
    loader      Loader        // Loads from .watch-now.yaml
    validator   Validator     // Validates configuration
}

// config/config.go
type Config struct {
    Services []ServiceConfig `yaml:"services"`
    Checks   []CheckConfig   `yaml:"checks"`
    Interval time.Duration   `yaml:"interval"`
    API      APIConfig       `yaml:"api"`
}

type ServiceConfig struct {
    Name     string            `yaml:"name"`
    Type     string            `yaml:"type"` // rest, grpc
    URL      string            `yaml:"url"`
    Health   string            `yaml:"health"`
    Headers  map[string]string `yaml:"headers"`
    Timeout  time.Duration     `yaml:"timeout"`
}
```

### 4. State Store

In-memory state management for monitor results:

```go
// core/state.go
type StateStore struct {
    mu      sync.RWMutex
    results map[string]*MonitorResult
    history map[string][]HistoryEntry
}

type MonitorResult struct {
    Name      string                 `json:"name"`
    Type      string                 `json:"type"`
    Status    Status                 `json:"status"`
    Message   string                 `json:"message"`
    Metadata  map[string]interface{} `json:"metadata"`
    Timestamp time.Time              `json:"timestamp"`
    Duration  time.Duration          `json:"duration"`
}

func (s *StateStore) Update(result *MonitorResult)
func (s *StateStore) Get(name string) *MonitorResult
func (s *StateStore) GetAll() map[string]*MonitorResult
func (s *StateStore) Subscribe() <-chan StateUpdate
```

### 5. Monitors

Perform actual health checks:

```go
// monitors/interface.go
type Monitor interface {
    Name() string
    Type() MonitorType
    Check(ctx context.Context) (*Result, error)
    Configure(config interface{}) error
}

// monitors/rest.go
type RESTMonitor struct {
    name     string
    url      string
    health   string
    timeout  time.Duration
    headers  map[string]string
}

// monitors/grpc.go
type GRPCMonitor struct {
    name    string
    address string
    service string
    timeout time.Duration
}

// monitors/quality.go
type QualityMonitor struct {
    name    string
    command string
    args    []string
    timeout time.Duration
}
```

### 6. Monitor Registry

Manages available monitors and their lifecycle:

```go
// core/registry.go
type MonitorRegistry struct {
    monitors map[string]Monitor
    mu       sync.RWMutex
}

func (r *MonitorRegistry) Register(monitor Monitor) error
func (r *MonitorRegistry) Get(name string) (Monitor, error)
func (r *MonitorRegistry) List() []Monitor
```

### 7. Scheduler

Executes monitors based on configuration:

```go
// core/scheduler.go
type Scheduler struct {
    interval   time.Duration
    timeout    time.Duration
    concurrent int
    monitors   []Monitor
}

func (s *Scheduler) Run(ctx context.Context) <-chan Result
```

### 8. API Server

Provides REST API and SSE endpoints:

```go
// api/server.go
type APIServer struct {
    state  *StateStore
    port   int
    router *mux.Router
}

func (s *APIServer) Start() error
func (s *APIServer) handleGetState(w http.ResponseWriter, r *http.Request)
func (s *APIServer) handleGetMonitor(w http.ResponseWriter, r *http.Request)
func (s *APIServer) handleSSE(w http.ResponseWriter, r *http.Request)

// API Endpoints:
// GET /api/state - Get all monitor states
// GET /api/monitor/:name - Get specific monitor state
// GET /api/health - API health check
// GET /sse - Server-sent events stream
```

### 9. Plugin System

Enables extensibility:

```go
// plugins/interface.go
type Plugin interface {
    Name() string
    Version() string
    Init(config interface{}) error
    Execute(ctx context.Context) (interface{}, error)
}

// plugins/manager.go
type PluginManager struct {
    plugins   map[string]Plugin
    loader    PluginLoader
    registry  *Registry
}
```

## Data Flow

```
1. Configuration Phase
   ├── Load .watch-now.yaml
   ├── Validate configuration
   └── Initialize state store

2. Initialization Phase
   ├── Create monitors from config
   ├── Register monitors
   ├── Start API server
   └── Initialize SSE broadcaster

3. Execution Phase (Loop)
   ├── Scheduler triggers checks
   ├── Monitors execute concurrently
   ├── Results collected
   ├── State store updated
   ├── SSE events broadcast
   └── Terminal output rendered

4. API Service
   ├── Serve REST endpoints
   ├── Handle SSE connections
   ├── Provide real-time updates
   └── Expose state for UI
```

## Key Design Patterns

### 1. Plugin Architecture

```go
// Plugins are loaded dynamically
// plugins/loader.go
func LoadPlugin(path string) (Plugin, error) {
    p, err := plugin.Open(path)
    if err != nil {
        return nil, err
    }
    
    symbol, err := p.Lookup("Plugin")
    if err != nil {
        return nil, err
    }
    
    return symbol.(Plugin), nil
}
```

### 2. Strategy Pattern for Monitors

Different monitor types implement the same interface:

```go
type Result struct {
    Name      string
    Type      MonitorType
    Status    Status
    Message   string
    Metadata  map[string]interface{}
    Timestamp time.Time
}
```

### 3. Observer Pattern for Results

```go
type ResultObserver interface {
    OnResult(result *Result)
}

type ResultCollector struct {
    observers []ResultObserver
}
```

### 4. Factory Pattern for Detectors

```go
func CreateDetector(projectPath string) Detector {
    if exists(filepath.Join(projectPath, "go.mod")) {
        return &GoDetector{}
    }
    if exists(filepath.Join(projectPath, "package.json")) {
        return &NodeDetector{}
    }
    // ... more detectors
}
```

## Configuration Schema

```yaml
# .watch-now.yaml
services:
  - name: api-service
    type: rest
    url: http://localhost:8080
    health: /health
    timeout: 5s
    headers:
      Authorization: Bearer ${API_TOKEN}
  
  - name: grpc-service
    type: grpc
    address: localhost:50051
    service: health.v1.HealthService
    timeout: 5s

checks:
  - name: format
    command: make
    args: ["fmt"]
    
  - name: lint
    command: make
    args: ["lint"]
    
  - name: tests
    command: make
    args: ["test"]

interval: 30s

api:
  enabled: true
  port: 9090
```

## Extensibility Points

### 1. Custom Monitors

```go
// example_plugin.go
type CustomMonitor struct{}

func (m *CustomMonitor) Name() string { return "custom" }
func (m *CustomMonitor) Check(ctx context.Context) (*Result, error) {
    // Custom logic here
}
```

### 2. Output Formatters

```go
type Formatter interface {
    Format(results []Result) ([]byte, error)
}
```

### 3. Notification Handlers

```go
type Notifier interface {
    Notify(event Event) error
}
```

## Performance Considerations

### 1. Concurrent Execution

All monitors run concurrently with configurable parallelism:

```go
sem := make(chan struct{}, s.concurrent)
for _, monitor := range s.monitors {
    sem <- struct{}{}
    go func(m Monitor) {
        defer func() { <-sem }()
        result, err := m.Check(ctx)
        results <- result
    }(monitor)
}
```

### 2. Caching

Results are cached to avoid redundant checks:

```go
type Cache struct {
    store sync.Map
    ttl   time.Duration
}
```

### 3. Resource Limits

Prevent resource exhaustion:

```go
type Limiter struct {
    cpu    int
    memory int64
    fds    int
}
```

## Error Handling

### 1. Graceful Degradation

If a monitor fails, others continue:

```go
func (e *Engine) Run() {
    for result := range e.scheduler.Run(ctx) {
        if result.Error != nil {
            e.handleError(result.Error)
            continue
        }
        e.collector.Collect(result)
    }
}
```

### 2. Retry Logic

Built-in retry for transient failures:

```go
type RetryableMonitor struct {
    Monitor
    maxRetries int
    backoff    time.Duration
}
```

## Security Considerations

### 1. No Credentials in Config

Sensitive data uses environment variables:

```yaml
notifications:
  - type: slack
    webhook: ${SLACK_WEBHOOK_URL}
```

### 2. Plugin Sandboxing

Plugins run with limited permissions:

```go
type SandboxedPlugin struct {
    plugin Plugin
    limits ResourceLimits
}
```

### 3. Secure Communication

HTTPS by default for web UI:

```go
server := &http.Server{
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
}
```

## Future Architecture Considerations

1. **Distributed Monitoring**: Multiple watch-now instances reporting to central dashboard
2. **Historical Data**: Time-series storage for trend analysis
3. **AI-Powered Insights**: Anomaly detection and predictive alerts
4. **Cloud Integration**: Monitoring cloud resources (AWS, GCP, Azure)
5. **Mobile Apps**: Native mobile apps for monitoring on the go