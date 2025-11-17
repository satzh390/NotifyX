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
   - `orgId` (used for tenant partitioning, no header overrides)
   - `scope` or `scp` (space-delimited string or string array)
3. Scopes drive route-level authorization:
   - `notify:write` for mutating resources (e.g., `POST /subscribers`)
   - `notify:read` for read-only endpoints (e.g., `GET /subscribers/:id`)

The service currently persists data via MongoDB through `core/adapters/mongo`. Configure `storage.mongo` in `config/config.yaml` (or override via env) to point at your local cluster or the Docker Compose Mongo service.

### Using the local docker stack

`docker-compose.local.yaml` spins up MongoDB, Kafka, LocalStack, and a mock OAuth server. After running:

```
docker compose -f ../docker-compose.local.yaml up -d
```

configure `config/config.yaml` to use issuer `http://localhost:8081/default` and JWKS `http://localhost:8081/default/jwks`, then start the API as usual.
The bundled config already points `storage.mongo.uri` at the Compose Mongo instance (`mongodb://localhost:27017`); update it if you run Mongo elsewhere.

