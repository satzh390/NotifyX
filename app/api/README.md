## notifyx-api

This service exposes the public REST interface used to manage organizations, customers, subscribers, groups, templates, rules, and ingest events.

### Quick Start

1. **Start local infrastructure:**
   ```cmd
   docker compose -f ../../docker-compose.local.yaml up -d
   ```

2. **Run the API:**
   ```cmd
   go run ./cmd
   ```

3. **Verify it's running:**
   ```cmd
   curl http://localhost:8080/health
   ```

The API will be available at `http://localhost:8080` with Swagger docs at `http://localhost:8080/swagger/index.html`.

### Running locally

Edit `config/config.yaml` with your issuer/JWKS details, then run:

```
go run ./cmd
```

To load a different config path, set `NOTIFYX_API_CONFIG=/path/to/custom.yaml`.

### Configuration

Configuration is loaded from `config/config.yaml` by default. To use a different config file, set the `NOTIFYX_API_CONFIG` environment variable:

**Windows (Command Prompt):**
```cmd
set NOTIFYX_API_CONFIG=C:\path\to\custom.yaml
```

**Linux/Mac:**
```bash
export NOTIFYX_API_CONFIG=/path/to/custom.yaml
```

#### Overriding Config Values with Environment Variables

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

**Windows (Command Prompt):**
```cmd
REM Override MongoDB URI
set NOTIFYX_API_STORAGE__MONGO__URI=mongodb://production:27017

REM Override OAuth issuer
set NOTIFYX_API_OAUTH__ISSUER=https://auth.example.com

REM Run the API
go run ./cmd
```

**Linux/Mac:**
```bash
# Override MongoDB URI
export NOTIFYX_API_STORAGE__MONGO__URI="mongodb://production:27017"

# Override OAuth issuer
export NOTIFYX_API_OAUTH__ISSUER="https://auth.example.com"

# Run the API
go run ./cmd
```

**Using a .env file:**

**Windows (Command Prompt):**
```cmd
REM Create .env file with:
REM NOTIFYX_API_STORAGE__MONGO__URI=mongodb://production:27017
REM NOTIFYX_API_STORAGE__MONGO__DATABASE=notifyx_prod
REM NOTIFYX_API_OAUTH__ISSUER=https://auth.example.com
REM NOTIFYX_API_OAUTH__JWKS=https://auth.example.com/jwks

REM Load and run (requires a .env loader or manual set commands)
```

**Linux/Mac:**
```bash
# .env file
NOTIFYX_API_STORAGE__MONGO__URI=mongodb://production:27017
NOTIFYX_API_STORAGE__MONGO__DATABASE=notifyx_prod
NOTIFYX_API_OAUTH__ISSUER=https://auth.example.com
NOTIFYX_API_OAUTH__JWKS=https://auth.example.com/jwks

# Load and run
source .env
go run ./cmd
```

**Important Notes:**
- Environment variables take precedence over config file values
- Use double underscores (`__`) to represent dots (`.`) in nested config paths
- Array values (like `audiences`) should be comma-separated strings
- The config file is still required - environment variables only override specific values

### Authentication & Authorization

1. Clients must send `Authorization: Bearer <access_token>` on every call.
2. The API validates the JWT locally using the configured issuer + JWKS, ensuring signature, expiry, and (optionally) audience are correct. Required claims:
   - `customerId` (used for tenant partitioning, no header overrides)
   - `scope` or `scp` (space-delimited string or string array)
3. Scopes drive route-level authorization:
   - `notify:write` for mutating resources (e.g., `POST /subscribers`)
   - `notify:read` for read-only endpoints (e.g., `GET /subscribers/:id`)

#### Getting an OAuth Token

To get an access token from the local OAuth server:

**Windows (Command Prompt):**
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

#### Example API Call

Once you have the token, use it in the `Authorization` header:

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

**Note:** Pagination is 0-based. Use `page=0` for the first page, `page=1` for the second page, etc.

The service currently persists data via MongoDB through `core/adapters/mongo`. Configure `storage.mongo` in `config/config.yaml` (or override via env) to point at your local cluster or the Docker Compose Mongo service.

### Pagination

All list endpoints support pagination with the following query parameters:
- `page` (0-based): Page number (default: 0)
- `pageSize`: Number of items per page (default: 20, max: 100)
- `sortBy`: Sort field(s) in format `field:asc` or `field:desc` (comma-separated for multiple fields)
- `sortOrder`: Global sort order (asc/desc) - used when `sortBy` is not specified

Example:
```
GET /api/v1/groups?page=0&pageSize=10&sortBy=name:asc,createdAt:desc
```

### Using the local docker stack

`docker-compose.local.yaml` spins up MongoDB, Kafka, LocalStack, and a mock OAuth server. After running:

```
docker compose -f ../docker-compose.local.yaml up -d
```

configure `config/config.yaml` to use issuer `http://localhost:8081/default` and JWKS `http://localhost:8081/default/jwks`, then start the API as usual.
The bundled config already points `storage.mongo.uri` at the Compose Mongo instance (`mongodb://localhost:27017`); update it if you run Mongo elsewhere.

