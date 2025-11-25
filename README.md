# NotifyX

NotifyX — Event-Driven Real-Time Notification System

NotifyX is a high-performance, event-driven notification system designed to deliver notifications with low latency, high throughput, and a pluggable architecture. It focuses on decoupling the notification engine from authentication, user management, and other platform-specific concerns so consumers can integrate custom components as needed.

**Key goals**

- Event-driven: Trigger notifications solely based on events.
- Decoupled: No hard dependency on internal auth/user systems; integrations are pluggable.
- High performance: Designed for low-latency, high-throughput workflows and horizontal scalability.
- Extensible: Allow custom notifiers, template providers, and user-storage backends.
- Reliable: Support retry strategies and configurable delivery semantics.

**Contributing**

- Open an issue to discuss major changes before implementing.
- Follow the repository's coding style and add tests for new behavior.
- Provide clear documentation and a small example for new integrations.

**License**

This repository does not include a license file. Add a license (e.g., `MIT` or `Apache-2.0`) as appropriate for your project.

---

## Local Infrastructure

Launch the supporting services (MongoDB, Kafka, LocalStack, mock OAuth server) with:

```
docker compose -f docker-compose.local.yaml up -d
```

The mock OAuth server issues tokens for issuer `http://localhost:8081/default` with scopes `notify:read` and `notify:write`. Update `app/api/config/config.yaml` if you need different claims or audiences. When you're done testing, tear everything down with:

```
docker compose -f docker-compose.local.yaml down -v
```

By default the API expects MongoDB at `mongodb://localhost:27017/notifyx`; adjust `storage.mongo` inside `app/api/config/config.yaml` if you point it somewhere else.

## Getting Started

This guide will help you set up and run NotifyX locally using Docker Compose for infrastructure and Go for the services.

### Prerequisites

- [Docker](https://www.docker.com/get-started) and Docker Compose
- [Go](https://go.dev/dl/) 1.23 or later
- `curl` (for testing API endpoints)

### Step 1: Start Local Infrastructure

Start all required services (MongoDB, Kafka, Redis, LocalStack, and mock OAuth server) using Docker Compose:

**Windows (Command Prompt):**
```cmd
docker compose -f docker-compose.local.yaml up -d
```

**Linux/Mac:**
```bash
docker compose -f docker-compose.local.yaml up -d
```

This will start the following services:
- **MongoDB** on port `27017`
- **Kafka** on port `9092`
- **Redis** on port `6379`
- **LocalStack** (AWS services) on port `4566`
- **Mock OAuth Server** on port `8081`

Verify services are running:
```cmd
docker compose -f docker-compose.local.yaml ps
```

### Step 2: Configure the Services

The default configuration files are already set up to work with the local Docker Compose stack:

- **API Config**: `app/api/config/config.yaml`
  - MongoDB: `mongodb://localhost:27017`
  - OAuth: `http://localhost:8081/default`
  
- **Processor Config**: `app/processor/config/config.yaml`
  - MongoDB: `mongodb://localhost:27017`
  - Kafka: `localhost:9092`

No changes needed unless you want to customize the configuration. See [Configuration](#configuration) section for details.

### Step 3: Run the API Service

**Windows (Command Prompt):**
```cmd
cd app\api
go run ./cmd
```

**Linux/Mac:**
```bash
cd app/api
go run ./cmd
```

The API will start on `http://localhost:8080` by default.

**Verify API is running:**
```cmd
curl http://localhost:8080/health
```

You should see:
```json
{"status":"ok","service":"notifyx-api"}
```

### Step 4: Run the Processor Service

Open a new terminal window and start the processor:

**Windows (Command Prompt):**
```cmd
cd app\processor
go run ./cmd
```

**Linux/Mac:**
```bash
cd app/processor
go run ./cmd
```

The processor will:
- Connect to Kafka and listen for events on the configured input topic
- Process events, resolve recipients, and publish delivery tasks to worker topics
- Use Redis cache if enabled (enabled by default in local config)

**Note:** Both services can run simultaneously. The processor will automatically consume events from Kafka as they are published.

### Step 5: Test the Setup

#### 5.1 Get an OAuth Token

**Windows (Command Prompt):**
```cmd
curl -X POST "http://localhost:8081/default/token" -H "Content-Type: application/x-www-form-urlencoded" -d "grant_type=client_credentials" -d "client_id=my-client" -d "client_secret=secret" -d "scope=notify:read notify:write"
```

**Linux/Mac:**
```bash
TOKEN=$(curl -X POST "http://localhost:8081/default/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=my-client" \
  -d "client_secret=secret" \
  -d "scope=notify:read notify:write" | jq -r '.access_token')
echo "Token: $TOKEN"
```

Save the `access_token` from the response.

#### 5.2 Make an API Call

**Windows (Command Prompt):**
```cmd
set TOKEN=YOUR_ACCESS_TOKEN_HERE
curl -X GET "http://localhost:8080/api/v1/groups?page=0&pageSize=10" -H "accept: application/json" -H "Authorization: Bearer %TOKEN%"
```

**Linux/Mac:**
```bash
TOKEN="YOUR_ACCESS_TOKEN_HERE"
curl -X GET "http://localhost:8080/api/v1/groups?page=0&pageSize=10" \
  -H "accept: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

#### 5.3 Access API Documentation

Open your browser and navigate to:
```
http://localhost:8080/swagger/index.html
```

This provides an interactive Swagger UI to explore and test all API endpoints.

### Stopping the Services

**Stop the API and Processor:**
Press `Ctrl+C` in each terminal window.

**Stop Docker Compose services:**
```cmd
docker compose -f docker-compose.local.yaml down
```

**Stop and remove all data (clean slate):**
```cmd
docker compose -f docker-compose.local.yaml down -v
```

### Troubleshooting

**Port already in use:**
- Check if services are already running: `docker compose -f docker-compose.local.yaml ps`
- Stop conflicting services or change ports in `docker-compose.local.yaml`

**Cannot connect to MongoDB/Kafka:**
- Ensure Docker Compose services are running: `docker compose -f docker-compose.local.yaml ps`
- Check service logs: `docker compose -f docker-compose.local.yaml logs mongo` (or `kafka`, `redis`, etc.)

**API/Processor fails to start:**
- Verify config files exist: `app/api/config/config.yaml` and `app/processor/config/config.yaml`
- Check error messages for missing required configuration values
- Ensure Go modules are downloaded: `go mod download` in each service directory

## Configuration

Both the API and Processor services support configuration via YAML files and environment variable overrides.

### Config File Location

- **API**: `app/api/config/config.yaml` (override with `NOTIFYX_API_CONFIG`)
- **Processor**: `app/processor/config/config.yaml` (override with `NOTIFYX_PROCESSOR_CONFIG`)

### Environment Variable Overrides

Any config value can be overridden using environment variables following this pattern:
- **API**: `NOTIFYX_API_<SECTION>__<KEY>` (dots become double underscores)
- **Processor**: `NOTIFYX_PROCESSOR_<SECTION>__<KEY>` (dots become double underscores)

**Example - Override MongoDB URI:**

**Windows (Command Prompt):**
```cmd
set NOTIFYX_API_STORAGE__MONGO__URI=mongodb://prod:27017
set NOTIFYX_PROCESSOR_STORAGE__MONGO__URI=mongodb://prod:27017
```

**Linux/Mac:**
```bash
export NOTIFYX_API_STORAGE__MONGO__URI="mongodb://prod:27017"
export NOTIFYX_PROCESSOR_STORAGE__MONGO__URI="mongodb://prod:27017"
```

For complete documentation and examples, see:
- [API Configuration Guide](app/api/README.md#configuration)
- [Processor Configuration Guide](app/processor/README.md#configuration)

### 1. Get an OAuth Token

To authenticate with the API, first obtain an access token from the OAuth server:

**Windows (Command Prompt):**
```cmd
curl -X POST "http://localhost:8081/default/token" -H "Content-Type: application/x-www-form-urlencoded" -d "grant_type=client_credentials" -d "client_id=my-client" -d "client_secret=secret" -d "scope=notify:read notify:write"
```

**Linux/Mac:**
```bash
TOKEN=$(curl -X POST "http://localhost:8081/default/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=my-client" \
  -d "client_secret=secret" \
  -d "scope=notify:read notify:write" | jq -r '.access_token')
echo "Token: $TOKEN"
```

### 2. Call the API

Use the token in the `Authorization` header for all API requests:

**Windows (Command Prompt):**
```cmd
set TOKEN=YOUR_ACCESS_TOKEN_HERE
curl -X GET "http://localhost:8080/api/v1/groups?page=0&pageSize=10&sortOrder=asc" -H "accept: application/json" -H "Authorization: Bearer %TOKEN%"
```

**Linux/Mac:**
```bash
TOKEN="YOUR_ACCESS_TOKEN_HERE"
curl -X GET "http://localhost:8080/api/v1/groups?page=0&pageSize=10&sortOrder=asc" \
  -H "accept: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Pagination

All list endpoints use **0-based pagination**:
- `page=0` - First page
- `page=1` - Second page
- `page=2` - Third page
- etc.

Query parameters:
- `page` (default: 0) - Page number (0-based)
- `pageSize` (default: 20, max: 100) - Items per page
- `sortBy` - Sort field(s) in format `field:asc` or `field:desc` (comma-separated)
- `sortOrder` - Global sort order when `sortBy` is not specified

### 4. API Documentation

Once the API is running, access the interactive Swagger documentation at:
```
http://localhost:8080/swagger/index.html
```

### Updating Swagger Docs

If you change the API annotations, regenerate the Swagger bundle so `docs/` stays in sync:

**1. Install the CLI (one time)**
```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
```

**2. Rebuild the docs from the API module**

**Windows (Command Prompt):**
```cmd
cd app\api
%USERPROFILE%\go\bin\swag.exe init -g cmd\main.go -o docs
```

**Linux/Mac:**
```bash
cd app/api
$(go env GOPATH)/bin/swag init -g cmd/main.go -o docs
```

Commit the updated `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml` files with your change.