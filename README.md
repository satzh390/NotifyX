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

### 1. Get an OAuth Token

To authenticate with the API, first obtain an access token from the OAuth server:

**Windows (PowerShell):**
```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8081/default/token" `
  -Method Post `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    grant_type = "client_credentials"
    client_id = "my-client"
    client_secret = "secret"
    scope = "notify:read notify:write"
  }
$token = $response.access_token
Write-Host "Token: $token"
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

**Windows (PowerShell):**
```powershell
$token = "YOUR_ACCESS_TOKEN_HERE"
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/groups?page=0&pageSize=10&sortOrder=asc" `
  -Method Get `
  -Headers @{
    "Authorization" = "Bearer $token"
    "Accept" = "application/json"
  }
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