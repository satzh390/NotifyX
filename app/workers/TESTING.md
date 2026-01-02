# Testing Guide for Workers

This document describes how to run unit tests and integration tests for the worker projects.

## Test Structure

### Unit Tests
- Located in `*_test.go` files alongside the source code
- Use mocks for external dependencies
- Fast execution, no external services required

### Integration Tests
- Located in `internal/integration/` directories
- Use build tag `//go:build integration`
- Require MongoDB and Kafka to be running
- Test real database interactions

## Running Tests

### Unit Tests Only

```bash
# SMS Worker
cd workers/worker-sms
go test ./...

# Email Worker
cd workers/worker-email
go test ./...

# Workerx Library
cd workerx
go test ./...
```

### Integration Tests

First, start the test infrastructure:

```bash
docker compose -f ../../../docker-compose.local.yaml up -d
```

Then run integration tests:

```bash
# SMS Worker Integration Tests
cd workers/worker-sms
go test -tags=integration ./internal/integration/...

# Email Worker Integration Tests
cd workers/worker-email
go test -tags=integration ./internal/integration/...
```

### All Tests (Unit + Integration)

```bash
# SMS Worker
cd workers/worker-sms
go test -tags=integration ./...

# Email Worker
cd workers/worker-email
go test -tags=integration ./...
```

## Test Coverage

To generate test coverage reports:

```bash
# SMS Worker
cd workers/worker-sms
go test -cover ./...

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Files

### workerx Library
- `workerx/worker/base_test.go` - Base worker tests
- `workerx/delivery/mongo_handler_test.go` - MongoDB handler tests
- `workerx/delivery/composite_handler_test.go` - Composite handler tests
- `workerx/delivery/broker_handler_test.go` - Kafka broker handler tests
- `workerx/template/render_test.go` - Template rendering tests

### worker-sms
- `internal/provider/sns_test.go` - SNS provider tests
- `internal/worker/sms_worker_test.go` - SMS worker unit tests
- `internal/worker/mocks_test.go` - Mock implementations
- `internal/integration/sms_worker_integration_test.go` - Integration tests

### worker-email
- `internal/provider/smtp_test.go` - SMTP provider tests
- `internal/worker/email_worker_test.go` - Email worker unit tests
- `internal/worker/mocks_test.go` - Mock implementations
- `internal/integration/email_worker_integration_test.go` - Integration tests

## Mock Providers

For testing, mock providers are available:
- `MockSMSProvider` - Mock SMS provider
- `MockEmailProvider` - Mock email provider
- `MockTemplateStore` - Mock template store

These can be used in unit tests to avoid external dependencies.

## Notes

- Integration tests require MongoDB running on `localhost:27017`
- Some provider tests are skipped (SNS, SMTP) as they require actual credentials or test servers
- Run `go mod tidy` in each project directory before running tests to resolve dependencies

