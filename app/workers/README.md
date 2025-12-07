# NotifyX Workers

This directory contains all channel-specific worker implementations. Each worker is a separate, independently deployable service.

## Structure

```
workers/
├── worker-sms/      # SMS worker - listens to notifyx-worker-sms topic
├── worker-email/    # Email worker - listens to notifyx-worker-email topic
├── worker-push/     # Push notification worker - listens to notifyx-worker-push topic
└── worker-webhook/  # Webhook worker - listens to notifyx-worker-webhook topic
```

## Shared Library

All workers use the shared `workerx` located at `app/workerx/` which provides:
- Base worker implementation
- Delivery result handlers (MongoDB, Kafka)
- Template rendering
- Base configuration

## Running Workers

Each worker can be run independently:

```bash
# SMS Worker
cd workers/worker-sms
go run ./cmd

# Email Worker
cd workers/worker-email
go run ./cmd

# Push Worker
cd workers/worker-push
go run ./cmd

# Webhook Worker
cd workers/worker-webhook
go run ./cmd
```

## Configuration

Each worker has its own configuration file:
- `worker-sms/config/config.yaml`
- `worker-email/config/config.yaml`
- `worker-push/config/config.yaml`
- `worker-webhook/config/config.yaml`

Configuration can be overridden via environment variables:
- `NOTIFYX_WORKER_SMS__*` for SMS worker
- `NOTIFYX_WORKER_EMAIL__*` for Email worker
- `NOTIFYX_WORKER_PUSH__*` for Push worker
- `NOTIFYX_WORKER_WEBHOOK__*` for Webhook worker

## Kafka Topics

Each worker listens to its specific topic:
- SMS: `notifyx-worker-sms`
- Email: `notifyx-worker-email`
- Push: `notifyx-worker-push`
- Webhook: `notifyx-worker-webhook`

These topics are configured in the processor's `config.yaml` under `worker.topics`.

