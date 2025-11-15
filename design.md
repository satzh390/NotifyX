# NotifyX — High-Level Architecture

NotifyX is an event-driven, highly scalable, real-time notification system designed to be modular, extensible, and easy to integrate with any platform. It supports user-specific subscriptions, dynamic templates, and pluggable delivery channels (Email, SMS, etc.).

---

## 🔷 Core Services Overview

| Service | Better Name | Purpose | Notes |
|--------|-------------|---------|-------|
| **API + user/tenant config service** | `notifyx-api` | Public-facing REST APIs | Handles event ingestion, user/tenant management, rule configuration, and authentication |
| **Event ingestion + rule evaluation** | `notifyx-processor` | Consumes events, resolves rules, fanout | Real-time stream processing; evaluates notification rules and routes to appropriate channels |
| **Delivery worker (email, SMS, webhook)** | `notifyx-worker` | One binary; per-channel workers via config | Modular workers spawned based on config; supports pluggable providers (SendGrid, Twilio, etc.) |
| **Admin UI** | `notifyx-console` | Angular/React | Web-based dashboard for managing tenants, rules, templates, and monitoring delivery |

---

## 🔷 Architecture Overview

NotifyX is organized into four major functional layers. Each layer is horizontally scalable and pluggable so customers can bring their own storage, brokers, and notifiers.

### 1) Web Configuration Layer (User-facing configuration services)
- User Configuration
  - Manage user subscriptions
  - Preferences (channels, frequency, opt-in events)
  - User contact information (email, phone)
- Template & Rule Configuration
  - Template CRUD (Email, SMS)
  - Rule CRUD (EventType → Notification Config)
  - Rule versioning, preview, and test dispatch

### 2) Event Processing Layer (Ingest + Fanout merged)
A unified layer that consumes events, applies rules, and fans out notifications to recipients.
Responsibilities:
- Consume events from a message broker (Kafka default)
  - Broker is pluggable: Kafka, RabbitMQ, NATS, Kinesis, SQS
- Load rule and template metadata (from configured template store)
- Identify interested recipients using:
  - User subscriptions
  - Per-user preferences (channels, language, do-not-disturb)
  - Opt-in event types
- Generate idempotency keys (ex: `<org_id>:<event_id>`)
- Produce channel-specific delivery tasks to worker queues (per-channel queues possible)

### 3) Message Delivery Worker Layer
Asynchronous workers that build and send notifications.
Responsibilities:
- Fetch & parse templates from template store (S3 or configured store)
- Merge event data with user data and preferences
- Build final messages (HTML email, SMS text, JSON payload for webhooks)
- Send via external providers (SMTP, Twilio, Webhook endpoints)
- Implement retry policy with exponential backoff
- Send failed messages to DLQ (dead-letter queue) after max retries
- Emit delivery status updates/events back into the system for logging and admin UI

### 4) Web Application (Admin UI)
A web-based console for operators to manage configuration and inspect system state:
- Manage user subscriptions
- Manage templates & rules with preview and test dispatch
- View delivery logs, retry status, and audit history
- Search/filter by `org_id`, event type, user, delivery status

---


## 🔷 Core Design Principles

- Configurable Components: everything is pluggable and configured via `application.yml`/env vars.
  - Ingest Consumer: Kafka default; can swap to NATS, RabbitMQ, SQS.
  - User DB: start with DynamoDB or Cassandra; pluggable to Postgres/MongoDB.
  - Template Storage: default S3; pluggable to GCS/Azure Blob or DB-backed storage.
- Scalable & Real-Time: event ingestion → fanout → delivery is asynchronous and horizontally scalable.
- Minimal Coupling: components communicate via well-defined messages/queues; internal state is isolated per `org_id`.
- Extensible: add new channels (Push, WhatsApp), brokers, or storage backends via well-documented SPI / interfaces.

---

## 🔷 Authentication & Authorization

- NotifyX does not provide its own identity provider.
- Supports external OAuth2 / OIDC providers.
- API validates tokens against a customer's identity provider (or via a cached introspection endpoint).
- Authorization is scope-based; API-level permissions are configurable (e.g., `notify:read`, `notify:write`, `notify:admin`).
- JWTs / tokens should include `org_id` in a claim (or an API header must set `org_id`) so all operations are scoped to an organization.

---

## 🔷 Core Entities

- Event
  - `type` (string)
  - `data` (object)
  - CloudEvents-compatible recommended
- Template
  - Email or SMS
  - Subject, body, variables, optional localization
- UserConfig
  - `emailId`, `phoneNumber`
  - `preferences` (channels, language, DND windows)
  - `eventTypes[]` (subscribed events)
- Rule
  - Maps `eventType` → `notifiers[]`
- Notifier
  - `channel`: `sms` | `email` | `webhook` | etc.
  - `templateRef`: pointer to template

---

## 🔷 Typical Data Flow

1. Producer emits an event:
   - Option A: HTTP API `POST /events`
   - Option B: Publish to broker topic (e.g., Kafka `notifications.events`)
2. Event Processing Layer consumes the event and validates it (CloudEvents schema optional).
3. Load rule(s) for `event.type` and query subscription store for interested users (filter by `org_id`).
4. For each recipient: build per-channel delivery tasks and write to delivery queue.
5. Delivery workers pick tasks, fetch templates and user data, render messages, and call notifiers.
6. Delivery status (success/failure/retry) is emitted and persisted to delivery logs; failed tasks may go to DLQ.
7. Admin UI queries delivery logs and shows status by `org_id`.

---

## 🔷 Configurability & Multi-Tenancy Notes

- `org_id` must be included with all requests and event messages to maintain strict multitenancy.
- Systems should use `(org_id, event_id)` as the idempotence key.
- Each org may supply its own storage and notifier credentials; configuration is per-org.
- Provide a default global configuration and allow org-level overrides (runtime config DB or per-tenant config files).

---

## 🔷 Reliability & Operational Concerns

- Delivery guarantees: support configurable semantics (at-least-once by default; exactly-once where broker/storage supports it).
- Retries and DLQ: retry with exponential backoff and a capped max attempts; move to DLQ on permanent failures.
- Observability: metrics (events/sec, deliveries/sec, retry rate), traces (event_id → delivery task traces), and structured logs tagged with `org_id`.
- Backpressure: when delivery backends are slow, materialize queue depth metrics and auto-scale workers.
- Security: secrets (SMTP, Twilio API keys) stored in a secret manager (Vault / AWS Secrets Manager).

---

## 🔷 Example Minimal Configuration (conceptual)

- `messageBroker.type: kafka`
- `userStore.type: dynamodb`
- `templateStore.type: s3`
- `notifiers.smtp.host`, `notifiers.twilio.*` from secret manager
- `worker.maxRetries: 3`, `worker.backoffBaseSeconds: 2`

---

## 🔷 Next Steps / Suggested Implementation Plan

Phase 0 — Preparation
- Finalize tech choices (Java Spring Boot or Go for services; Angular for Admin UI).
- Define interfaces for stores, notifiers, and broker consumers.

Phase 1 — MVP
- Implement API service with endpoints: `POST /events`, `POST /users`, `POST /subscriptions`, `POST /templates`, `POST /rules`.
- Implement Event Processing Layer that ingests HTTP events and fans out to an in-memory delivery queue.
- Implement Delivery Worker that performs email delivery via SMTP and records delivery events in an in-memory log.
- Ship a minimal Angular Admin UI to CRUD templates and view delivery logs.

Phase 2 — Productionization
- Swap in persistent stores (DynamoDB/Cassandra, S3, Redis), add Kafka consumer, and implement per-org configuration loading.
- Add retries, DLQ, and metrics (Prometheus) + tracing (OpenTelemetry).

Phase 3 — Scale & Extensibility
- Implement additional notifier adapters (SMS, webhooks), quota enforcement, anti-spam heuristics, and operator tooling (k8s manifests, helm charts).

---

