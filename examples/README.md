# watch-now Examples

This directory contains example configurations for different types of projects.

## acecam Examples

Three configurations for monitoring the acecam monorepo at different levels:

### acecam-simple.yaml ✅ **TESTED**

Basic backend quality checks only (no services). Works immediately:
- Go formatting, linting, complexity, deadcode checks
- API documentation generation

**Usage:**
```bash
cp examples/acecam-simple.yaml /path/to/acecam/monorepo/.watch-now.yaml
cd /path/to/acecam/monorepo
watch-now --once
```

### acecam-full.yaml

Complete monitoring when services are running:
- All backend quality checks
- 6 microservice health monitoring (IAM, Social, Analytics, Gaming, Notification, Logging)

**Usage:**
```bash
cp examples/acecam-full.yaml /path/to/acecam/monorepo/.watch-now.yaml
cd /path/to/acecam/monorepo
make docker-up  # Start infrastructure
make dev        # Start all services
watch-now       # Monitor everything
```

### acecam-monorepo.yaml

Ultimate configuration including frontend checks:
- Everything from acecam-full.yaml
- Frontend TypeScript, linting, deadcode, tests
- Infrastructure dependency checks

**Usage:**
```bash
cp examples/acecam-monorepo.yaml /path/to/acecam/monorepo/.watch-now.yaml
cd /path/to/acecam/monorepo
make docker-up && make dev
cd frontend && pnpm install  # Ensure frontend deps
watch-now
```

### Expected Output

```
watch-now - Universal Development Monitor
================================================================================

[14:23:45] System Status
--------------------------------------------------------------------------------

SERVICES Services:
  [OK] iam - Check passed in 45ms
  [OK] social - Check passed in 32ms
  [OK] analytics - Check passed in 28ms
  [OK] gaming - Check passed in 41ms
  [OK] notification - Check passed in 38ms
  [OK] logging - Check passed in 35ms

CHECKS Code Quality:
  [OK] go-format - Check passed in 120ms
  [OK] go-lint - Check passed in 3.2s
  [OK] go-complexity - Check passed in 180ms
  [OK] go-deadcode - Check passed in 240ms
  [OK] go-test - Check passed in 8.5s
  [OK] go-build - Check passed in 12.3s
  [OK] api-docs - Check passed in 2.1s
  [OK] frontend-typecheck - Check passed in 5.8s
  [OK] frontend-lint - Check passed in 2.3s
  [OK] frontend-deadcode - Check passed in 1.2s
  [OK] frontend-test - Check passed in 7.9s

[OK] STATUS: All systems operational
================================================================================
```

### Customization

Adjust the configuration for your environment:

- **Service ports**: Update the port numbers in the `services` section
- **Health endpoints**: Some services might use `/healthz` instead of `/health`
- **Timeouts**: Increase timeouts for slower operations
- **Additional checks**: Add custom quality checks or remove unused ones
- **Dependencies**: Add or remove infrastructure dependencies as needed

## Creating Your Own Configuration

1. Start with the basic template:
   ```yaml
   services: []
   checks: []
   interval: 30s
   api:
     enabled: true
     port: 9090
   ```

2. Add your services and their health endpoints
3. Add your quality checks (make targets, npm scripts, etc.)
4. Test with `watch-now --once`
5. Run continuously with `watch-now`