## notifyx-api

This service exposes the public REST interface used to manage subscribers and ingest events.  

### Running locally

Edit `config/config.yaml` with your issuer/JWKS details, then run:

```
go run ./cmd
```

To load a different config path, set `NOTIFYX_API_CONFIG=/path/to/custom.yaml`.

Any field can be overridden via environment variables using the pattern `NOTIFYX_API_<SECTION>__<KEY>`. Examples:

| Environment variable | Overrides |
| --- | --- |
| `NOTIFYX_API_HTTP__ADDR` | `http.addr` |
| `NOTIFYX_API_OAUTH__ISSUER` | `oauth.issuer` |
| `NOTIFYX_API_OAUTH__JWKS` | `oauth.jwks` |
| `NOTIFYX_API_OAUTH__AUDIENCES` | `oauth.audiences` (comma-separated list) |

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

**Windows (PowerShell):**
```powershell
curl -X POST "http://localhost:8081/default/token" `
  -H "Content-Type: application/x-www-form-urlencoded" `
  -d "grant_type=client_credentials" `
  -d "client_id=my-client" `
  -d "client_secret=secret" `
  -d "scope=notify:read notify:write"
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

**Windows (PowerShell):**
```powershell
$token = "YOUR_ACCESS_TOKEN_HERE"
curl -X GET "http://localhost:8080/api/v1/groups?page=0&pageSize=10&sortOrder=asc" `
  -H "accept: application/json" `
  -H "Authorization: Bearer $token"
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

