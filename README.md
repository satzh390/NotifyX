# NotifyX

**Event-Driven Real-Time Notification System**

NotifyX is a high-performance, event-driven notification system designed to deliver notifications with low latency, high throughput, and a pluggable architecture. It focuses on decoupling the notification engine from authentication, user management, and other platform-specific concerns so consumers can integrate custom components as needed.

## Key Features

- **Event-driven**: Trigger notifications solely based on events
- **Decoupled**: No hard dependency on internal auth/user systems; integrations are pluggable
- **High performance**: Designed for low-latency, high-throughput workflows and horizontal scalability
- **Extensible**: Allow custom notifiers, template providers, and user-storage backends
- **Reliable**: Support retry strategies and configurable delivery semantics
- **Multi-channel**: Support for Email, SMS, and Push notifications
- **Multi-app Push**: Support for multiple push notification apps per organization with per-app provider configuration
- **Custom Filters**: Business-specific filtering logic for subscribers

## Documentation

- **[Architecture & Design](design/README.md)** - High-level architecture, design principles, and data flow
- **[API Service](app/api/README.md)** - REST API documentation, authentication, and configuration
- **[Processor Service](app/processor/README.md)** - Event processing, filtering, and custom filters
- **[Custom Filters](app/processor/internal/filter/custom/README.md)** - Guide for creating custom filter implementations

## Quick Start

### Prerequisites

- [Docker](https://www.docker.com/get-started) and Docker Compose
- [Go](https://go.dev/dl/) 1.23 or later
- `curl` (for testing API endpoints)

### Step 1: Start Local Infrastructure

Start all required services (MongoDB, Kafka, Redis, LocalStack, and mock OAuth server):

```bash
docker compose -f docker-compose.local.yaml up -d
```

This starts:
- **MongoDB** on port `27017`
- **Kafka** on port `9092`
- **Redis** on port `6379`
- **LocalStack** (AWS services) on port `4566`
- **Mock OAuth Server** on port `8081`

Verify services are running:
```bash
docker compose -f docker-compose.local.yaml ps
```

### Step 2: Run the Services

**Start the API Service:**
```bash
cd app/api
go run ./cmd
```
The API will be available at `http://localhost:8080`

**Start the Processor Service (in a new terminal):**
```bash
cd app/processor
go run ./cmd
```

### Step 3: Test the Setup

**1. Get an OAuth Token:**
```bash
curl -X POST "http://localhost:8081/default/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=my-client" \
  -d "client_secret=secret" \
  -d "scope=notify:read notify:write"
```

**2. Make an API Call:**
```bash
TOKEN="YOUR_ACCESS_TOKEN_HERE"
curl -X GET "http://localhost:8080/api/v1/groups?page=0&pageSize=10" \
  -H "accept: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

**3. Access API Documentation:**
Open `http://localhost:8080/swagger/index.html` in your browser for interactive Swagger UI.

### Stopping Services

**Stop the services:** Press `Ctrl+C` in each terminal window.

**Stop Docker Compose services:**
```bash
docker compose -f docker-compose.local.yaml down
```

**Stop and remove all data:**
```bash
docker compose -f docker-compose.local.yaml down -v
```

## Configuration

Both services support configuration via YAML files and environment variable overrides.

### Config File Locations

- **API**: `app/api/config/config.yaml` (override with `NOTIFYX_API_CONFIG`)
- **Processor**: `app/processor/config/config.yaml` (override with `NOTIFYX_PROCESSOR_CONFIG`)

### Environment Variable Overrides

Any config value can be overridden using environment variables:
- **API**: `NOTIFYX_API_<SECTION>__<KEY>` (dots become double underscores)
- **Processor**: `NOTIFYX_PROCESSOR_<SECTION>__<KEY>` (dots become double underscores)

**Example:**
```bash
export NOTIFYX_API_STORAGE__MONGO__URI="mongodb://prod:27017"
export NOTIFYX_PROCESSOR_STORAGE__MONGO__URI="mongodb://prod:27017"
```

For complete configuration documentation, see:
- [API Configuration Guide](app/api/README.md#configuration)
- [Processor Configuration Guide](app/processor/README.md#configuration)

## Project Structure

```
NotifyX/
├── app/
│   ├── api/              # REST API service
│   ├── processor/        # Event processing service
│   ├── core/             # Shared domain models and storage interfaces
│   └── httpx/            # Shared HTTP utilities
├── design/                # Architecture and design documentation
├── infrastructure/        # Infrastructure setup scripts
└── docker-compose.local.yaml  # Local development stack
```

## Contributing

- Open an issue to discuss major changes before implementing
- Follow the repository's coding style and add tests for new behavior
- Provide clear documentation and examples for new integrations

## License

This repository does not include a license file. Add a license (e.g., `MIT` or `Apache-2.0`) as appropriate for your project.
