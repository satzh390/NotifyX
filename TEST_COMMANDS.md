# Test Commands for NotifyX

## Quick Reference

### Prerequisites
Ensure MongoDB is running for integration tests:
```powershell
docker compose -f docker-compose.local.yaml up -d
```

### Run All Tests (Unit + Integration)

**Windows (PowerShell) - From project root:**
```powershell
# Start MongoDB if not running
docker compose -f docker-compose.local.yaml up -d

# Run all tests
cd app\api; go test -tags=integration ./... -v;
cd ..\core; go test -tags=integration ./... -v;
cd ..\httpx; go test ./... -v;
cd ..\processor; go test ./... -v
```

**Linux/Mac - From project root:**
```bash
# Start MongoDB if not running
docker compose -f docker-compose.local.yaml up -d

# Run all tests
cd app/api && go test -tags=integration ./... -v \
  && cd ../core && go test -tags=integration ./... -v \
  && cd ../httpx && go test ./... -v \
  && cd ../processor && go test ./... -v
```

### Run Unit Tests Only

**Windows (PowerShell):**
```powershell
# API unit tests
cd app\api
go test ./... -v

# Core unit tests
cd ..\core
go test ./... -v

# HTTPX unit tests
cd ..\httpx
go test ./... -v

# Processor unit tests
cd ..\processor
go test ./... -v
```

**Linux/Mac:**
```bash
# API unit tests
cd app/api
go test ./... -v

# Core unit tests
cd ../core
go test ./... -v

# HTTPX unit tests
cd ../httpx
go test ./... -v

# Processor unit tests
cd ../processor
go test ./... -v
```

### Run Integration Tests Only

**Windows (PowerShell):**
```powershell
# Ensure MongoDB is running first!
docker compose -f docker-compose.local.yaml up -d

# API integration tests
cd app\api
go test -tags=integration ./... -v

# Core integration tests
cd ..\core
go test -tags=integration ./... -v

# Processor currently has unit tests only; run `go test ./... -v` if needed.
```

**Linux/Mac:**
```bash
# Ensure MongoDB is running first!
docker compose -f docker-compose.local.yaml up -d

# API integration tests
cd app/api
go test -tags=integration ./... -v

# Core integration tests
cd ../core
go test -tags=integration ./... -v

# Processor currently has unit tests only; run `go test ./... -v` if needed.
```

### Run Specific Test Package

**Windows (PowerShell):**
```powershell
# Example: Test subscriber handler only
cd app\api
go test ./internal/handlers/subscriber -v

# Example: Test with integration tag
go test -tags=integration ./internal/handlers/subscriber/integration -v
```

**Linux/Mac:**
```bash
# Example: Test subscriber handler only
cd app/api
go test ./internal/handlers/subscriber -v

# Example: Test with integration tag
go test -tags=integration ./internal/handlers/subscriber/integration -v
```

### Test Coverage

**Windows (PowerShell):**
```powershell
cd app\api
go test -tags=integration -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Linux/Mac:**
```bash
cd app/api
go test -tags=integration -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Notes

- **Unit tests** don't require MongoDB and run quickly
- **Integration tests** require MongoDB to be running
- Use `-v` flag for verbose output
- Use `-tags=integration` to include integration tests
- Pagination is now **0-based** (page 0 = first page, page 1 = second page, etc.)

