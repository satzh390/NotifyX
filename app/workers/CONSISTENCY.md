# Worker Consistency Checklist

This document ensures all workers follow the same patterns and structure.

## Structure

Each worker should have:
- `cmd/main.go` - Entry point
- `config/config.go` - Configuration loader
- `config/config.yaml` - Default configuration
- `internal/provider/` - Provider implementations
- `internal/worker/` - Worker implementation
- `README.md` - Documentation
- `Dockerfile` - Container build
- `go.mod` - Go module definition
- `*_test.go` - Test files

## Code Patterns

### Configuration
- Use `NOTIFYX_WORKER_{CHANNEL}_CONFIG` env var for config path
- Use `NOTIFYX_WORKER_{CHANNEL}_` prefix for env var overrides
- Double underscores (`__`) for nested config paths

### Imports
- Import workerx packages: `github.com/notifyx/workerx/...`
- Import core packages: `github.com/notifyx/core/...`
- Use local config: `github.com/notifyx/worker-{channel}/config`

### Error Handling
- Use structured logging with slog
- Send failed tasks to DLQ
- Log delivery results (success/failure)

### Testing
- Unit tests for providers
- Unit tests for workers (with mocks)
- Integration tests (optional, in `integration/` folder)

## Documentation

Each README should include:
1. Quick Start section
2. How It Works section
3. Configuration section (with env var examples)
4. Provider-specific configuration
5. Delivery Result Handling section
6. Using Local Docker Stack section

## Docker

- Use multi-stage builds
- Use distroless base image
- Copy config file to `/app/config.yaml`
- Set config env var

## Dependencies

- Add `github.com/stretchr/testify` for tests
- Keep dependencies minimal
- Use replace directives for local modules

