# NotifyX Workers Architecture

The NotifyX worker system is split into separate projects for each channel type, sharing common functionality through a shared library.

## Structure

```
app/
├── workerx/             # Shared library for all workers
│   ├── worker/          # Base worker implementation
│   ├── delivery/        # Delivery result handlers
│   ├── template/        # Template rendering
│   └── config/          # Base configuration
│
└── workers/             # All worker projects
    ├── worker-sms/      # SMS worker project
    │   ├── cmd/         # Main entry point
    │   ├── config/      # SMS-specific config
    │   ├── internal/
    │   │   ├── provider/    # SMS providers (SNS, etc.)
    │   │   └── worker/      # SMS worker implementation
    │   └── go.mod
    │
    ├── worker-email/    # Email worker project
    │   ├── cmd/         # Main entry point
    │   ├── config/      # Email-specific config
    │   ├── internal/
    │   │   ├── provider/    # Email providers (SMTP, etc.)
    │   │   └── worker/      # Email worker implementation
    │   └── go.mod
    │
    ├── worker-push/     # Push notification worker project
    │   ├── cmd/         # Main entry point
    │   ├── config/      # Push-specific config (base config only)
    │   ├── internal/
    │   │   ├── provider/    # Push providers (APNS, Firebase, ProviderManager)
    │   │   └── worker/      # Push worker implementation
    │   └── go.mod
```

## Worker Library (`workerx`)

The shared library provides:

- **Base Worker**: Common Kafka consumption, task processing, error handling
- **Delivery Handlers**: MongoDB and Kafka result storage
- **Template Rendering**: Shared template rendering logic
- **Configuration**: Base configuration structure

## Channel-Specific Workers

Each worker project:

1. **Listens to a specific Kafka topic** (e.g., `notifyx-worker-sms`)
2. **Implements channel-specific providers** (SNS for SMS, SMTP for Email, etc.)
3. **Uses the shared worker library** for common functionality
4. **Has its own configuration** for provider-specific settings

## Running Workers

Each worker is a separate process:

```bash
# SMS Worker
cd app/workers/worker-sms
go run cmd/main.go

# Email Worker
cd app/workers/worker-email
go run cmd/main.go

# Push Worker
cd app/workers/worker-push
go run cmd/main.go
```

## Configuration

Each worker has its own config file:

- `workers/worker-sms/config/config.yaml` - SMS worker config
- `workers/worker-email/config/config.yaml` - Email worker config
- `workers/worker-push/config/config.yaml` - Push worker base config (providers via AppConfig API)

Configuration can be overridden via environment variables with prefix:
- `NOTIFYX_WORKER_SMS__*` for SMS worker
- `NOTIFYX_WORKER_EMAIL__*` for Email worker
- `NOTIFYX_WORKER_PUSH__*` for Push worker (base config only)

### Push Worker Multi-App Architecture

The push worker implements a multi-app architecture:
- **AppConfig API**: Push providers are managed via `/api/v1/app-configs` endpoints
- **ProviderManager**: Dynamically loads providers from AppConfig store and caches them per app
- **Multiple Tokens**: Subscribers can have multiple push tokens via `pushTokens` map (appId → token)
- **Rule Metadata**: Rules include `metadata.appId` to route push notifications to the correct app

For more details, see:
- [Multi-App Push Design](../../docs/MULTI_APP_PUSH_DESIGN.md)
- [Multi-App Push Implementation](../../docs/MULTI_APP_PUSH_IMPLEMENTATION.md)

## Kafka Topics

Each worker listens to its specific topic:

- SMS: `notifyx-worker-sms`
- Email: `notifyx-worker-email`
- Push: `notifyx-worker-push`

These topics are configured in the processor's `config.yaml` under `worker.topics`.

## Benefits of This Architecture

1. **Separation of Concerns**: Each channel is independent
2. **Independent Scaling**: Scale each worker type independently
3. **Independent Deployment**: Deploy updates to one channel without affecting others
4. **Code Reuse**: Common functionality shared via library
5. **Provider Flexibility**: Easy to add new providers per channel
6. **Configuration Isolation**: Each worker has its own config

