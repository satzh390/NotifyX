# NotifyX API Service

This service exposes the public REST interface used to manage organizations, customers, subscribers, groups, templates, rules, and ingest events.

## Quick Start

1. **Start local infrastructure:**
   ```bash
   docker compose -f ../../docker-compose.local.yaml up -d
   ```

2. **Run the API:**
   ```bash
   go run ./cmd
   ```

3. **Verify it's running:**
   ```bash
   curl http://localhost:8080/health
   ```

The API will be available at `http://localhost:8080` with Swagger docs at `http://localhost:8080/swagger/index.html`.

## Configuration

Configuration is loaded from `config/config.yaml` by default. To use a different config file, set the `NOTIFYX_API_CONFIG` environment variable:

**Windows:**
```cmd
set NOTIFYX_API_CONFIG=C:\path\to\custom.yaml
```

**Linux/Mac:**
```bash
export NOTIFYX_API_CONFIG=/path/to/custom.yaml
```

### Overriding Config Values with Environment Variables

Any configuration value can be overridden using environment variables. The pattern is:
- **Prefix**: `NOTIFYX_API_`
- **Key mapping**: Dots (`.`) in config keys become double underscores (`__`)
- **Case**: Environment variable names are case-sensitive (use uppercase)

**Examples:**

| Config File Path | Environment Variable | Example Value |
|-----------------|---------------------|---------------|
| `http.addr` | `NOTIFYX_API_HTTP__ADDR` | `:8080` |
| `oauth.issuer` | `NOTIFYX_API_OAUTH__ISSUER` | `https://auth.example.com` |
| `oauth.jwks` | `NOTIFYX_API_OAUTH__JWKS` | `https://auth.example.com/jwks` |
| `oauth.audiences` | `NOTIFYX_API_OAUTH__AUDIENCES` | `app1,app2,app3` (comma-separated) |
| `storage.mongo.uri` | `NOTIFYX_API_STORAGE__MONGO__URI` | `mongodb://prod:27017` |
| `storage.mongo.database` | `NOTIFYX_API_STORAGE__MONGO__DATABASE` | `notifyx_prod` |

**How to use:**

**Windows:**
```cmd
set NOTIFYX_API_STORAGE__MONGO__URI=mongodb://production:27017
set NOTIFYX_API_OAUTH__ISSUER=https://auth.example.com
go run ./cmd
```

**Linux/Mac:**
```bash
export NOTIFYX_API_STORAGE__MONGO__URI="mongodb://production:27017"
export NOTIFYX_API_OAUTH__ISSUER="https://auth.example.com"
go run ./cmd
```

**Important Notes:**
- Environment variables take precedence over config file values
- Use double underscores (`__`) to represent dots (`.`) in nested config paths
- Array values (like `audiences`) should be comma-separated strings
- The config file is still required - environment variables only override specific values

## Authentication & Authorization

1. Clients must send `Authorization: Bearer <access_token>` on every call.
2. The API validates the JWT locally using the configured issuer + JWKS, ensuring signature, expiry, and (optionally) audience are correct. Required claims:
   - `customerId` (used for tenant partitioning, no header overrides)
   - `scope` or `scp` (space-delimited string or string array)
3. Scopes drive route-level authorization:
   - `notify:write` for mutating resources (e.g., `POST /subscribers`)
   - `notify:read` for read-only endpoints (e.g., `GET /subscribers/:id`)

### Getting an OAuth Token

To get an access token from the local OAuth server:

**Windows:**
```cmd
curl -X POST "http://localhost:8081/default/token" -H "Content-Type: application/x-www-form-urlencoded" -d "grant_type=client_credentials" -d "client_id=my-client" -d "client_secret=secret" -d "scope=notify:read notify:write"
```

**Linux/Mac:**
```bash
curl -X POST "http://localhost:8081/default/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=my-client" \
  -d "client_secret=secret" \
  -d "scope=notify:read notify:write"
```

The response will contain an `access_token` field. Extract this token and use it in subsequent API calls.

### Example API Call

Once you have the token, use it in the `Authorization` header:

**Windows:**
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

## Pagination

All list endpoints support pagination with the following query parameters:
- `page` (0-based): Page number (default: 0)
- `pageSize`: Number of items per page (default: 20, max: 100)
- `sortBy`: Sort field(s) in format `field:asc` or `field:desc` (comma-separated for multiple fields)
- `sortOrder`: Global sort order (asc/desc) - used when `sortBy` is not specified

**Example:**
```
GET /api/v1/groups?page=0&pageSize=10&sortBy=name:asc,createdAt:desc
```

## API Endpoints

### Core Resources

- **Organizations**: `/api/v1/organizations` - Manage organizations
- **Customers**: `/api/v1/customers` - Manage customers (bound to organizations)
- **Subscribers**: `/api/v1/subscribers` - Manage subscribers
- **Groups**: `/api/v1/groups` - Manage subscriber groups
- **Templates**: `/api/v1/templates` - Manage notification templates
- **Rules**: `/api/v1/rules` - Manage event routing rules
- **App Configs**: `/api/v1/app-configs` - Manage push notification app configurations

### App Configuration (Push Notifications)

The AppConfig API allows you to manage push notification configurations for multiple apps within an organization.

**Create AppConfig:**
```bash
curl -X POST "http://localhost:8080/api/v1/app-configs" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "id": "my-ios-app",
    "name": "My iOS App",
    "provider": "apns",
    "apns": {
      "keyId": "YOUR_KEY_ID",
      "teamId": "YOUR_TEAM_ID",
      "bundleId": "com.yourcompany.app",
      "keyPath": "/etc/secrets/apns/apns-key.p8",
      "production": false
    }
  }'
```

**Get AppConfig:**
```bash
curl -X GET "http://localhost:8080/api/v1/app-configs/my-ios-app" \
  -H "Authorization: Bearer $TOKEN"
```

**List AppConfigs:**
```bash
curl -X GET "http://localhost:8080/api/v1/app-configs?page=0&pageSize=20" \
  -H "Authorization: Bearer $TOKEN"
```

**Update AppConfig:**
```bash
curl -X PUT "http://localhost:8080/api/v1/app-configs/my-ios-app" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "id": "my-ios-app",
    "name": "My iOS App Updated",
    "provider": "apns",
    "apns": {
      "keyId": "NEW_KEY_ID",
      "teamId": "YOUR_TEAM_ID",
      "bundleId": "com.yourcompany.app",
      "keyPath": "/etc/secrets/apns/apns-key.p8",
      "production": true
    }
  }'
```

**Delete AppConfig:**
```bash
curl -X DELETE "http://localhost:8080/api/v1/app-configs/my-ios-app" \
  -H "Authorization: Bearer $TOKEN"
```

### Subscriber Push Tokens

Subscribers now support multiple push tokens via the `pushTokens` map:

**Create Subscriber with Push Tokens:**
```bash
curl -X POST "http://localhost:8080/api/v1/subscribers" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "subscriberId": "user-123",
    "email": "user@example.com",
    "pushTokens": {
      "my-ios-app": "apns-token-123",
      "my-android-app": "fcm-token-456"
    },
    "preferences": {
      "channels": {
        "push": true
      },
      "language": "en"
    }
  }'
```

**Update Push Tokens:**
```bash
curl -X PATCH "http://localhost:8080/api/v1/subscribers/user-123" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "pushTokens": {
      "my-ios-app": "new-apns-token-789",
      "my-android-app": "fcm-token-456"
    }
  }'
```

### Rules with App ID

For push notifications, include `appId` in Rule metadata:

**Create Rule with App ID:**
```bash
curl -X POST "http://localhost:8080/api/v1/rules" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "eventType": "order.created",
    "channels": ["push"],
    "templateRefs": {
      "push": "order-notification-template"
    },
    "metadata": {
      "appId": "my-ios-app"
    }
  }'
```

## API Documentation

Once the API is running, access the interactive Swagger documentation at:
```
http://localhost:8080/swagger/index.html
```

### Updating Swagger Docs

If you change the API annotations, regenerate the Swagger bundle:

**1. Install the CLI (one time):**
```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
```

**2. Rebuild the docs:**

**Windows:**
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

## Using the Local Docker Stack

`docker-compose.local.yaml` spins up MongoDB, Kafka, LocalStack, and a mock OAuth server. After running:

```bash
docker compose -f ../../docker-compose.local.yaml up -d
```

Configure `config/config.yaml` to use:
- OAuth issuer: `http://localhost:8081/default`
- OAuth JWKS: `http://localhost:8081/default/jwks`
- MongoDB: `mongodb://localhost:27017`

The bundled config already points to the Compose services; update it if you run services elsewhere.
