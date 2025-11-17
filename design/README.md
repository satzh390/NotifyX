# NotifyX — High-Level Architecture (Subscriber + Multi-Group Model)

NotifyX is an event-driven, highly scalable, real-time notification platform designed to be modular, extensible, and easy to integrate into any infrastructure. It supports multi-channel delivery (Email, SMS, Webhooks, Push), dynamic templates, subscriber-level preferences, and flexible group-based fanout.

---

**🔷 Core Services Overview**

| Service | Name | Purpose | Notes |
|---|---:|---|---|
| API + subscriber/tenant config service | `notifyx-api` | Public REST APIs | Event ingestion, subscriber/group/tenant management, templates, rules, authentication |
| Event ingestion + rule evaluation | `notifyx-processor` | Consume events → apply rules → determine subscribers | Real-time fanout based on subscriber preferences & group membership |
| Delivery worker (email, SMS, webhook, push) | `notifyx-worker` | Channel-specific delivery | Modular plugins for SendGrid, Twilio, SMTP, Push, custom adapters |
| Admin UI | `notifyx-console` | Angular/React Dashboard | Manage templates, groups, rules, subscriber profiles, logs |

---

## 🔷 Architecture Overview

NotifyX is separated into four horizontally scalable layers, each pluggable:

- Web Configuration Layer
- Event Processing Layer
- Delivery Worker Layer
- Admin UI Layer

### 1) Web Configuration Layer

Subscriber Configuration

- Tenant-scoped subscribers: create/update subscribers
- Stored per-subscriber:
  - Contact info: `email`, `phone`, `webhookUrl`, `pushToken`
  - Preferences: enabled channels, language, DND window
  - Subscribed event types (opt-ins)
  - Metadata
- Subscribers can belong to multiple groups

Group Management

- Create groups, add/remove subscribers
- Many-to-many membership (subscriber ↔ groups)
- Group semantics defined by tenant (e.g., `operators`, `admins`, `warehouse`)

Template & Rule Configuration

- CRUD templates (Email, SMS, Push, Webhook JSON)
- Rules map `eventType` → channels + template + default recipients
- Supported recipient types: `subscriberIds[]`, `groups[]`, `broadcast` (all subscribers)
- Versioning, preview, test-send

### 2) Event Processing Layer (Ingest + Fanout)

Responsibilities:

- Receive events via HTTP `/events` or broker (Kafka/NATS/RabbitMQ/SQS)
- Load rule metadata and default recipients
- Expand recipients using `subscriberIds[]`, `groups[]`, `broadcast`, and rule defaults
- Merge and deduplicate subscriber list
- Filter subscribers by preferences (disabled channels, DND, unsubscribed event types)
- Generate idempotency keys: `<org_id>:<event_id>:<channel>`
- Produce delivery tasks to per-channel queues

### 3) Message Delivery Worker Layer

Workers perform:

- Template retrieval (S3/Blob/custom store)
- Merge payload + subscriber attributes
- Channel-specific formatting (email HTML + subject, SMS body, push payload, webhook JSON)
- Provider communication and adapter abstraction
- Retries with exponential backoff and DLQ routing for terminal failures
- Emit delivery logs/status events

### 4) Admin UI

UI features:

- Manage subscribers and multi-membership groups
- Manage templates (preview + versioning)
- Manage rules and test-send
- View event → fanout → delivery logs, traces, retries
- Search/filter by `org_id`, event type, `subscriberId`, `groupId`, channel, status

---

## 🔷 Core Design Principles

- Complete pluggability (storage, broker, providers)
- NotifyX stores subscriber profiles only — zero coupling with tenant user DBs
- Tenants own subscriber lifecycle, group definitions, and auth policies
- Multi-group membership supported (N:N)
- Horizontally scalable and asynchronous workflows

---

## 🔷 Authentication & Authorization

- Tenants use external OAuth2/OIDC; NotifyX validates tokens
- Required metadata: `org_id` (token claim or header)
- Scopes: `notify:read`, `notify:write`, `notify:admin`
- API keys supported for server-to-server integrations

---

## 🔷 Core Entities (Updated)

- Subscriber
  - `subscriberId` (external eg: UserId)
  - `org_id` (For Non SAAS product like ecommerce, org_id can be some unique identifier like flipkart,  myntra)
  - `email`, `phone`, `pushToken`, `webhookUrl`
  - `preferences` { channels: { email/sms/push/webhook: true/false }, language, allowedDays[], notificationWindow { start, end } }
  - `subscribedEventTypes: [string]`(Will add this preference later)
  - `groups: [groupId]`
  - `metadata` (optional)
- Group
  - `groupId`, `org_id`, `name`, `description`
  - `subscribers: [subscriberId]` (many-to-many)
  - `metadata` (optional)
- Event
  - `eventId`, `org_id`, `type`, `payload/data`
  - `recipients` { `subscriberIds?`, `groups?`, `broadcast?` }
- Rule
  - `eventType`, `org_id`, `channels[]`, `defaultRecipients` (subscriberIds/groups/broadcast)
- Channel
  - `type`: `email` | `sms` | `push` | `webhook` | `custom`
  - `templateRef`

---

## 🔷 Typical Data Flow (Subscriber + Group Fanout)

1) Client system emits event via `POST /events` (example payload):

```json
{
  "org_id":"acme",
  "type":"ORDER_DELAYED",
  "recipients":{
    "subscriberIds":["s1"],
    "groups":["operators","warehouse"],
    "broadcast":false
  },
  "payload":{"orderId":123}
}
```

2) Processor loads rule + expands subscribers: explicit `subscriberIds` + members of `operators` + members of `warehouse` + optional rule defaults
3) Deduplicate subscriberIds
4) For each subscriber → expand per-channel tasks (respecting preferences and DND)
5) Workers send notifications + log status

---

## 🔷 Multi-Tenancy Notes

- All entities partitioned by `org_id`
- Tenant isolation enforced at API, storage, and delivery pipeline
- Storage options: schema-per-tenant, table-per-tenant, or shared table with `org_id` partition(start with shared table with `org_id` partition)

---

## 🔷 Reliability & Ops

- Delivery semantics: at-least-once default; exactly-once optional with idempotent stores
- Quotas per-org and per-channel rate limiting
- Retries + DLQ + DLQ inspection UI
- Observability: Prometheus metrics, OpenTelemetry traces, structured logs with `org_id`
- Secrets stored in secret manager (Vault/AWS Secrets Manager)

---

## 🔷 Minimal Config Example

- `messageBroker.type: kafka`
- `subscriberStore.type: dynamodb`
- `groupStore.type: dynamodb`
- `templateStore.type: s3`
- `retry.maxAttempts: 3`
- `retry.backoffBaseMs: 2000`
- `notifier.mail.provider: smtp-server`
- `notifier.sms.provider: sns`

---

## **Execution Plan (Improved & Actionable)**

This section breaks the work into concrete milestones, deliverables, acceptance criteria, and suggested timeline for a small team (2–4 engineers).

**Phase 0 — Foundations (1–2 weeks)**
- Deliverables: tech choices, API sketch, SPI definitions, data model (subscriber/group/event/rule)
- Acceptance: documented SPI interfaces, example `application.yml`, basic ERD, and a `README.md` for each service
- Tasks:
  - Decide language(s) and frameworks (Go/Java/Node) and message broker
  - Design REST routes and minimal JSON contracts for subscribers, groups, events, templates, rules
  - Create repository skeletons (monorepo or multiple repos) and CI templates

**Phase 1 — MVP (3–6 weeks)**
- Goal: complete end-to-end flow for email notifications and admin UI to manage subscribers/rules
- Deliverables: running `notifyx-api`, `notifyx-processor`, `notifyx-worker` (email), and `notifyx-console` with minimal features
- Acceptance criteria:
  - `POST /events` ingests events and results in delivery attempts in worker logs
  - Subscribers and groups CRUD working with in-memory or simple persistent store
  - Templates can be created and used by rules
  - Delivery logs show per-`org_id` status
- Tasks (incremental):
  1. Implement `notifyx-api` endpoints for Subscriber, Group, Template, Rule, Event ingestion
  2. Implement `notifyx-processor` that reads events, expands recipients (subscriberIds/groups/broadcast), filters preferences, and enqueues tasks (in-memory queue for MVP)
  3. Implement `notifyx-worker` for email using SMTP/SendGrid with template rendering and retry logic (in-memory DLQ)
  4. Build minimal `notifyx-console` to manage subscribers, groups, templates, and view logs
  5. Add unit tests and a simple integration test to exercise event→delivery path

**Phase 2 — Hardening (2–4 weeks)**
- Goal: make system production-ready for a pilot tenant
- Deliverables: Kafka ingestion, persistent stores (DynamoDB/Postgres), S3 template store, secrets integration, metrics, tracing
- Acceptance criteria:
  - Kafka consumer capable of high-throughput ingestion
  - Persistent storage of subscribers/groups and delivery logs
  - Metrics (Prometheus) and traces (OpenTelemetry) visible for end-to-end flows
  - Retry policies, exponential backoff, DLQ persisted and inspectable
- Tasks:
  - Swap in Kafka for processor ingestion and SQS/Redis for delivery queues where appropriate
  - Implement persistent stores and migration scripts for subscriber data
  - Integrate secret manager for provider credentials
  - Implement rate limiting and per-org quotas

**Phase 3 — Scale & Extensibility (ongoing)**
- Add channels (SMS, Push, Webhook), provider adapters, operator tooling (Helm charts), and anti-spam/quota enforcement

---

## **Rollout & Release Guidance**

- Start with a single pilot tenant and internal testing cluster
- Canary release: enable traffic for a percentage of tenant events and monitor delivery success and error rates
- Rollback: keep manual DLQ retry runbooks and automated feature flags to disable channels
- Migration: export/import subscriber and group data; provide scripts and validation checks

---

## **Testing & CI**

- Unit tests for SPI implementations and template rendering
- Integration tests: processor + worker + API using test Kafka or in-memory broker
- End-to-end smoke tests: create subscriber/group/rule → send event → assert delivery log
- CI: run tests and lint on pull requests; build container images for services

---

## **Observability & SLOs**

- Metrics: events/sec, deliveries/sec, avg latency (event→delivery), retry rate, DLQ rate
- Traces: sample 1%–10% of requests; ensure trace IDs flow from API → processor → worker → provider
- Alerts: high DLQ growth, sustained delivery failures, high retry rate, backpressure on queues

---

## **Security & Compliance**

- Secrets in Vault/AWS Secrets Manager; no secrets checked into repo
- Tenant data partitioning and encryption-at-rest for stored subscriber data
- Audit logs for template/rule changes and admin actions

---

## **Next Steps (short checklist)**

- Merge this updated `design.md` into `main` as the canonical design doc
- Create repo skeletons for the four core services with `README.md` and CI
- Implement Phase 0 tasks and mark task 2 (`Update design document`) as completed in the plan

---

