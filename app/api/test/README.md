# Testing Guide

## Running Tests

### Prerequisites

For integration tests, MongoDB must be running. Start the local infrastructure:

```bash
# From project root
docker compose -f docker-compose.yaml up -d
```

Or if using a different compose file:
```bash
docker compose -f docker-compose.local.yaml up -d
```

### Unit Tests (UT)

**Run all unit tests from project root:**

**Windows (PowerShell):**
```powershell
# Run API unit tests
cd app\api
go test ./... -v

# Run core unit tests
cd ..\core
go test ./... -v

# Run httpx unit tests
cd ..\httpx
go test ./... -v
```

**Linux/Mac:**
```bash
# Run API unit tests
cd app/api
go test ./... -v

# Run core unit tests
cd ../core
go test ./... -v

# Run httpx unit tests
cd ../httpx
go test ./... -v
```

**Run unit tests for a specific package:**
```bash
# Example: Test only subscriber handler
cd app/api
go test ./internal/handlers/subscriber -v

# Example: Test only query parsing
cd app/httpx
go test ./... -v
```

### Integration Tests (IT)

**Run all integration tests:**

**Windows (PowerShell):**
```powershell
# From project root
cd app\api
go test -tags=integration ./... -v

cd ..\core
go test -tags=integration ./... -v
```

**Linux/Mac:**
```bash
# From project root
cd app/api
go test -tags=integration ./... -v

cd ../core
go test -tags=integration ./... -v
```

### Run Both Unit and Integration Tests

**Windows (PowerShell):**
```powershell
# Run all tests (UT + IT) from project root
# Make sure MongoDB is running first!
docker compose -f docker-compose.yaml up -d

# Run all tests
cd app\api
go test -tags=integration ./... -v

cd ..\core
go test -tags=integration ./... -v

cd ..\httpx
go test ./... -v
```

**Linux/Mac:**
```bash
# Run all tests (UT + IT) from project root
# Make sure MongoDB is running first!
docker compose -f docker-compose.yaml up -d

# Run all tests
cd app/api
go test -tags=integration ./... -v

cd ../core
go test -tags=integration ./... -v

cd ../httpx
go test ./... -v
```

**One-liner to run all tests (from project root):**

**Windows (PowerShell):**
```powershell
# Ensure MongoDB is running, then:
cd app\api; go test -tags=integration ./... -v; cd ..\core; go test -tags=integration ./... -v; cd ..\httpx; go test ./... -v
```

**Linux/Mac:**
```bash
# Ensure MongoDB is running, then:
cd app/api && go test -tags=integration ./... -v && cd ../core && go test -tags=integration ./... -v && cd ../httpx && go test ./... -v
```

### Test Coverage

Generate coverage report:

**Windows (PowerShell):**
```powershell
# Generate coverage for unit tests
cd app\api
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Generate coverage for integration tests
go test -tags=integration -coverprofile=coverage-it.out ./...
go tool cover -html=coverage-it.out
```

**Linux/Mac:**
```bash
# Generate coverage for unit tests
cd app/api
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Generate coverage for integration tests
go test -tags=integration -coverprofile=coverage-it.out ./...
go tool cover -html=coverage-it.out
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

