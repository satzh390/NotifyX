# Testing Guide

## Running Tests

### Unit Tests

Run all unit tests:
```bash
cd app/core
go test ./...

cd app/api
go test ./...
```

Run tests for a specific package:
```bash
go test ./core/utils
go test ./api/internal/handlers
```

Run tests with verbose output:
```bash
go test -v ./...
```

### Integration Tests

Integration tests require MongoDB to be running. Start the local infrastructure:

```bash
docker compose -f docker-compose.local.yaml up -d
```

Run integration tests:
```bash
cd app/api
go test -tags=integration ./internal/handlers

cd app/core
go test -tags=integration ./adapters/mongo
```

Run all tests (unit + integration):
```bash
go test -tags=integration ./...
```

### Test Coverage

Generate coverage report:
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Structure

### Unit Tests

- `core/utils/*_test.go` - Tests for utility functions (query parsing, merge patch)
- `api/internal/handlers/*_handler_test.go` - Handler tests with mocked stores
- `api/internal/middlewares/authorization_test.go` - Authorization middleware tests
- `api/internal/server/fiber_test.go` - Server tests

### Integration Tests

- `api/internal/handlers/*_integration_test.go` - End-to-end API tests
- `core/adapters/mongo/mongo_store_test.go` - MongoDB adapter tests

## Writing Tests

### Unit Test Example

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "normal case",
            input:    "test",
            expected: "TEST",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FunctionName(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Integration Test Example

```go
//go:build integration
// +build integration

func TestIntegration_Feature(t *testing.T) {
    // Setup test environment
    app, cleanup := setupIntegrationTest(t)
    defer cleanup()

    // Test implementation
    // ...
}
```

## Continuous Integration

Tests should be run in CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run unit tests
  run: go test ./...

- name: Run integration tests
  run: |
    docker compose -f docker-compose.local.yaml up -d
    go test -tags=integration ./...
```

