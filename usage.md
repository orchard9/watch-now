# Usage Guide

This guide covers all aspects of using watch-now to monitor your development projects.

## Installation

### From Source
```bash
go install github.com/orchard9/watch-now@latest
```

### Using Homebrew (macOS/Linux)
```bash
brew tap orchard9/tools
brew install watch-now
```

### Pre-built Binaries
Download from [releases page](https://github.com/orchard9/watch-now/releases)

## Getting Started

### Initialize a New Project

```bash
cd /path/to/your/project
watch-now init
```

This will:
1. Auto-detect your project structure
2. Find services and their ports
3. Identify infrastructure components
4. Create a `.watch-now.yaml` configuration file

### Basic Monitoring

```bash
# Start continuous monitoring (updates every 60 seconds)
watch-now

# Run once and exit
watch-now --once

# Custom interval
watch-now --interval 30s

# JSON output
watch-now --format json

# Quiet mode (only errors)
watch-now --quiet
```

## Configuration

### Configuration File

watch-now looks for `.watch-now.yaml` in the current directory:

```yaml
# Project metadata
project:
  name: my-awesome-app
  type: microservices  # microservices, monolith, library

# Service definitions
services:
  - name: api
    type: http
    port: 8080
    health_endpoint: /healthz
    ready_endpoint: /readyz
    
  - name: worker
    type: process
    command: ps aux | grep worker
    
  - name: grpc-service
    type: grpc
    port: 50051
    health_service: grpc.health.v1.Health

# Infrastructure monitoring
infrastructure:
  - name: postgres
    type: docker
    container: myapp-postgres-1
    
  - name: redis
    type: tcp
    host: localhost
    port: 6379
    
  - name: elasticsearch
    type: http
    url: http://localhost:9200/_cluster/health

# Code quality checks
checks:
  formatting:
    command: make fmt
    enabled: true
    
  linting:
    command: make lint
    timeout: 30s
    
  complexity:
    command: make complexity
    max_value: 10
    
  dead_code:
    command: make deadcode
    
  tests:
    command: make test
    show_coverage: true
    
  build:
    command: make build

# Monitoring settings
monitor:
  interval: 60s
  timeout: 5s
  concurrent_checks: true
  
# Output settings
output:
  format: terminal  # terminal, json, web
  colors: true
  timestamps: true
  
# Notifications (optional)
notifications:
  - type: slack
    webhook: ${SLACK_WEBHOOK_URL}
    events: [service_down, build_failed]
    
  - type: email
    smtp_server: smtp.gmail.com:587
    from: monitor@example.com
    to: [dev-team@example.com]
    events: [all_clear, degraded, failure]
```

### Environment Variables

```bash
# Override config file
WATCH_NOW_CONFIG=/custom/path/config.yaml watch-now

# Set monitoring interval
WATCH_NOW_INTERVAL=30s watch-now

# Disable colors
NO_COLOR=1 watch-now

# Debug mode
WATCH_NOW_DEBUG=1 watch-now
```

## Command Line Options

```bash
watch-now [flags]

Flags:
  --once              Run once and exit
  --interval duration Monitoring interval (default 60s)
  --config string     Config file path (default ".watch-now.yaml")
  --format string     Output format: terminal, json, web (default "terminal")
  --port int          Web UI port (default 8888)
  --quiet             Only show errors
  --verbose           Show detailed output
  --no-color          Disable colored output
  --services string   Monitor specific services (comma-separated)
  --checks string     Run specific checks (comma-separated)
  --help              Show help
  --version           Show version
```

## Advanced Usage

### Service Filtering

```bash
# Monitor only specific services
watch-now --services api,worker

# Exclude services
watch-now --exclude-services legacy-service
```

### Check Filtering

```bash
# Run only specific checks
watch-now --checks linting,tests

# Skip specific checks
watch-now --skip-checks dead_code
```

### CI/CD Integration

```bash
# Exit with error code if any check fails
watch-now --once --exit-on-failure

# GitHub Actions example
- name: Monitor Project Health
  run: watch-now --once --format json > health-report.json
```

### Web Dashboard

```bash
# Start with web UI
watch-now --format web --port 8888

# Access at http://localhost:8888
```

### Custom Health Checks

Create custom checks using plugins:

```yaml
checks:
  custom_security_scan:
    type: plugin
    plugin: security-scanner
    config:
      scan_type: dependencies
      
  api_contract_test:
    type: plugin
    plugin: contract-tester
    config:
      spec_file: openapi.yaml
```

### Monitoring Profiles

Use different configs for different environments:

```bash
# Development
watch-now --config .watch-now.dev.yaml

# Production monitoring
watch-now --config .watch-now.prod.yaml
```

## Output Formats

### Terminal (Default)

Color-coded, real-time updating display with system status overview.

### JSON

```bash
watch-now --format json --once
```

Output structure:
```json
{
  "timestamp": "2024-07-30T14:23:45Z",
  "status": "degraded",
  "infrastructure": {
    "postgres": { "status": "healthy" },
    "redis": { "status": "healthy" }
  },
  "services": [
    {
      "name": "api",
      "status": "healthy",
      "response_time_ms": 45
    }
  ],
  "checks": {
    "formatting": { "passed": true },
    "linting": { "passed": false, "errors": ["..."] }
  }
}
```

### Web Dashboard

Interactive web interface with:
- Real-time updates via WebSocket
- Historical graphs
- Service dependency visualization
- Log aggregation
- Alert configuration

## Troubleshooting

### Common Issues

1. **Service not detected**
   ```bash
   watch-now --verbose  # Shows detection process
   ```

2. **Health check timeouts**
   ```yaml
   monitor:
     timeout: 10s  # Increase timeout
   ```

3. **Permission errors**
   ```bash
   # Ensure watch-now can access Docker
   sudo usermod -aG docker $USER
   ```

### Debug Mode

```bash
WATCH_NOW_DEBUG=1 watch-now --verbose
```

Shows:
- Configuration loading process
- Service detection logic
- Health check execution details
- Error stack traces

## Best Practices

1. **Start Simple**: Let auto-detection work first, then customize
2. **Version Control**: Commit `.watch-now.yaml` to your repository
3. **CI Integration**: Use `--once --exit-on-failure` in CI pipelines
4. **Custom Checks**: Add project-specific health checks
5. **Monitoring Profiles**: Use different configs for dev/staging/prod

## Examples

### Microservices Project

```yaml
project:
  name: e-commerce-platform
  type: microservices

services:
  - name: api-gateway
    port: 8080
    health_endpoint: /health
    
  - name: auth-service
    port: 8081
    health_endpoint: /health
    
  - name: product-service
    port: 8082
    health_endpoint: /health
    
  - name: order-service
    port: 8083
    health_endpoint: /health

infrastructure:
  - name: postgres
    type: docker
    container: platform-postgres
    
  - name: rabbitmq
    type: http
    url: http://localhost:15672/api/health/checks/virtual-hosts
```

### Monolithic Application

```yaml
project:
  name: django-app
  type: monolith

services:
  - name: web
    port: 8000
    health_endpoint: /health/
    
  - name: celery-worker
    type: process
    command: pgrep -f "celery worker"
    
  - name: celery-beat
    type: process
    command: pgrep -f "celery beat"

checks:
  migrations:
    command: python manage.py showmigrations --plan | grep -q "\[ \]"
    
  static_files:
    command: python manage.py collectstatic --check
```

### Library Project

```yaml
project:
  name: awesome-utils
  type: library

checks:
  tests:
    command: go test ./...
    show_coverage: true
    
  examples:
    command: go build ./examples/...
    
  documentation:
    command: go doc -all ./... > /dev/null
```