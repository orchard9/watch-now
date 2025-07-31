# watch-now

A universal development monitor that provides real-time health visibility for any software project.

## Overview

watch-now is a configuration-driven monitoring tool that provides comprehensive health monitoring for:
- REST and gRPC service health and availability
- External dependencies and infrastructure status
- Code quality metrics (lint, format, tests)
- Real-time state exposure via REST API and SSE
- Web UI support through API endpoints

## Quick Start

```bash
# Install watch-now
go install github.com/orchard9/watch-now@latest

# Create configuration
cat > .watch-now.yaml << EOF
services:
  - name: my-api
    type: rest
    url: http://localhost:8080
    health: /health

checks:
  - name: lint
    command: make
    args: ["lint"]
  - name: test
    command: make
    args: ["test"]

interval: 30s
api:
  enabled: true
  port: 9090
EOF

# Start monitoring
watch-now

# Access API at http://localhost:9090/api/state
# Stream updates at http://localhost:9090/sse
```

## Features

- **Easy Configuration**: Simple YAML-based configuration
- **Service Monitoring**: Monitor REST and gRPC services
- **Code Quality Checks**: Run linters, formatters, and tests
- **Real-time State**: In-memory state with API access
- **REST API**: Full state access via REST endpoints
- **SSE Support**: Real-time updates via Server-Sent Events
- **Language Agnostic**: Works with any toolchain
- **CI Integration**: Exit codes for automated pipelines

## Supported Stacks

- **Languages**: Go, JavaScript/TypeScript, Python, Java, Rust
- **Frameworks**: Express, FastAPI, Spring Boot, Chi, Gin
- **Infrastructure**: Docker, Kubernetes, AWS, GCP
- **Databases**: PostgreSQL, MySQL, MongoDB, Redis
- **CI/CD**: GitHub Actions, GitLab CI, Jenkins

## Example Output

```
[14:23:45] System Status
--------------------------------------------------------------------------------
INFRA Infrastructure:
  [OK] PostgreSQL
  [OK] Redis
  [WARN] Elasticsearch

SERVICES Services:
  [OK] API Service (HTTP:8080, gRPC:8081) - Last seen: 14:23:45
  [OK] Auth Service (HTTP:8082) - Last seen: 14:23:45
  [FAIL] Payment Service (HTTP:8084)

CI Code Quality:
  [OK] Formatting
  [OK] Linting
  [OK] Complexity (<10)
  [FAIL] Dead Code Check
  [OK] Build

STATUS: [WARN] Services need attention
================================================================================
Next check in 60 seconds... (Ctrl+C to stop)
```

## Configuration

watch-now uses a `.watch-now.yaml` configuration file:

```yaml
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

checks:
  - name: format
    command: make
    args: ["fmt"]
  - name: lint
    command: make
    args: ["lint"]
  - name: test
    command: make
    args: ["test"]

interval: 30s

api:
  enabled: true
  port: 9090
```

## API Endpoints

- `GET /api/state` - Get all monitor states
- `GET /api/monitor/:name` - Get specific monitor state  
- `GET /api/health` - API health check
- `GET /sse` - Server-sent events stream for real-time updates

## Documentation

- [Usage Guide](usage.md) - Detailed usage instructions
- [Architecture](architecture.md) - How watch-now works
- [Why watch-now?](why.md) - Motivation and use cases
- [Plugin Development](plugins.md) - Creating custom plugins

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.