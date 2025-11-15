# Repository Guidelines

## Project Structure & Module Organization
`main.go` wires the CLI, engine, and monitors. Domain logic lives under `internal/`: `core` manages the engine, scheduler, and in-memory state; `monitors` contains REST and quality checks; `api` serves the REST/SSE endpoints; `config` loads `.watch-now.yaml`; `detector` backs `watch-now --init`. Reference docs (`README.md`, `usage.md`, `architecture.md`) and sample configs in `examples/` when extending the configuration language or documenting new monitors.

## Build, Test, and Development Commands
- `make build` – compile the binary into `build/watch-now` with version metadata.
- `make run` / `make run-once` – build then launch continuous or single-iteration monitoring.
- `make test` – run `go test -v ./...` across all packages.
- `make coverage` – write coverage data to `coverage/coverage.out` and an HTML report.
- `make fmt`, `make lint`, `make complexity`, `make deadcode` – enforce formatting, linters, and quality gates used by CI.
- `make ci` – runs fmt → lint → complexity → deadcode → test → build; gate every PR on this target.
- `make install-deps` – install `gocyclo`, `staticcheck`, and `golangci-lint` locally.

## Coding Style & Naming Conventions
Go files must be `gofmt`-clean (tabs for indentation, std lib import grouping). Keep package and file names aligned with modules (`internal/core/state.go`, `internal/api/server.go`). Exported structs and functions use PascalCase; private helpers stay camelCase. Prefer small, composable monitors implementing `monitors.Monitor`. Run `make fmt lint` before pushing; linting relies on `golangci-lint run ./...`, so add directives sparingly and justify them in PR notes.

## Testing Guidelines
Place `_test.go` files beside their sources (e.g., `internal/monitors/rest_test.go`) and cover both success and failure states. Favor table-driven `t.Run` cases to describe monitors, detectors, and API handlers. Run `make test` for fast feedback and `make coverage` when adding modules; keep coverage high for scheduler and monitor registry code. When new YAML knobs are introduced, add an example in `examples/*.yaml` and, if practical, exercise it via a config-driven test harness in `internal/config`.

## Commit & Pull Request Guidelines
History uses short, imperative summaries (“Add REST API and SSE support with --port configuration”). Match that tone, keep the subject ≤ 72 characters, and expand details in the body when behavior changes. Every PR should link issues (if any), describe the scenario being monitored, list `make ci` results, and include screenshots or sample `watch-now` output when touching the API/UI. Highlight any new configuration keys plus migration notes for downstream agents.
