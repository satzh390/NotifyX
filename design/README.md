# NotifyX — High-Level Architecture (Subscriber + Multi-Group Model)

NotifyX is an event-driven, highly scalable, real-time notification platform designed to be modular, extensible, and easy to integrate into any infrastructure. It supports multi-channel delivery (Email, SMS, Webhooks, Push), dynamic templates, subscriber-level preferences, and flexible group-based fanout.

---

**🔷 Core Services Overview**

| Service | Name | Purpose | Notes |
|---|---:|---|---|
| API + subscriber/tenant config service | `notifyx-api` | Public REST APIs | Event ingestion, organization/customer/subscriber/group/tenant management, templates, rules, authentication |
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

**Organization & Customer Management**

- **Organizations**: Top-level entities representing companies or SaaS providers
  - `id`, `name`, `type` (company | saasProvider)
  - Used for multi-tenant SaaS scenarios where one organization manages multiple customers
- **Customers**: Business units or tenants within an organization
  - `id`, `orgId` (links to organization), `name`, `logo`, `metadata`
  - Customers belong to organizations (hierarchy: Organization → Customer → Subscribers)
  - For single-tenant scenarios, one organization may have one customer

**Subscriber Configuration**

- Tenant-scoped subscribers: create/update subscribers
- Stored per-subscriber:
  - Contact info: `email`, `phone`, `webhookUrl`, `pushTokens` (map of appId → token)
  - Preferences: enabled channels, language, DND window
  - Subscribed event types (opt-ins)
  - Metadata
- Subscribers can belong to multiple groups
- Subscribers are scoped to a `customerId` (which may be nested under an `orgId`)
- **Push Tokens**: `pushTokens` is a map of `appId → pushToken` to support multiple apps per subscriber
  - Legacy `pushToken` field is deprecated but still supported for backward compatibility

**Group Management**

- Create groups, add/remove subscribers
- Many-to-many membership (subscriber ↔ groups)
- Group semantics defined by tenant (e.g., `operators`, `admins`, `warehouse`)
- Groups are scoped to a `customerId`

**Template & Rule Configuration**

- CRUD templates (Email, SMS, Push, Webhook JSON)
- Rules map `eventType` → channels + template
- Rules support `metadata` field (e.g., `appId` for push notifications)
- Supported recipient types: `subscriberIds[]`, `groups[]`, `broadcast` (all subscribers), `directEmails[]`, `directPhones[]`
- Versioning, preview, test-send

**App Configuration (Push Notifications)**

- **AppConfig**: Push notification configuration bound to Organization
  - Each app has its own provider configuration (APNS, Firebase, etc.)
  - Managed via `/api/v1/app-configs` endpoints
  - App identification comes from `Rule.Metadata["appId"]`
  - Providers are cached per app in the push worker

### 2) Event Processing Layer (Ingest + Fanout)

Responsibilities:

- Receive events via HTTP `/events` or broker (Kafka/NATS/RabbitMQ/SQS)
- Load rule metadata
- **Recipients are optional in events** - if not provided, system uses:
  1. Subscribers/groups that have subscribed to the event type (`subscribedEventTypes`)
- Expand recipients using `subscriberIds[]`, `groups[]`, `broadcast`, `directEmails[]`, `directPhones[]`
- Merge and deduplicate subscriber list
- Filter subscribers by preferences (disabled channels, DND, unsubscribed event types)
- Generate idempotency keys: `<customerId>:<eventId>:<channel>`
- Produce delivery tasks to per-channel queues

**Recipient Resolution Priority:**
1. Explicit recipients in event (if provided)
2. Subscribers/groups subscribed to the event type (if no explicit recipients)

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

- Manage organizations and customers
- Manage subscribers and multi-membership groups
- Manage templates (preview + versioning)
- Manage rules and test-send
- View event → fanout → delivery logs, traces, retries
- Search/filter by `orgId`, `customerId`, event type, `subscriberId`, `groupId`, channel, status

---

## 🔷 Core Design Principles

- Complete pluggability (storage, broker, providers)
- NotifyX stores subscriber profiles only — zero coupling with tenant user DBs
- Tenants own subscriber lifecycle, group definitions, and auth policies
- Multi-group membership supported (N:N)
- Horizontally scalable and asynchronous workflows
- **Hierarchical multi-tenancy**: Organization → Customer → Subscribers (supports both single-tenant and multi-tenant SaaS models)

---

## 🔷 Authentication & Authorization

- Tenants use external OAuth2/OIDC; NotifyX validates tokens
- Required metadata: `customerId` (token claim or header, optionally paired with an `orgId` claim for SaaS providers)
- Scopes: `notify:read`, `notify:write`, `notify:admin`
- API keys supported for server-to-server integrations

---

## 🔷 Core Entities (Updated)

- **Organization**
  - `id` (immutable)
  - `name`, `type` (company | saasProvider)
  - `createdAt`, `updatedAt`
  - Top-level entity for multi-tenant SaaS scenarios

- **Customer**
  - `id` (immutable)
  - `orgId` (immutable, links to organization)
  - `name`, `logo`, `metadata`
  - `createdAt`, `updatedAt`
  - Business unit or tenant within an organization

- **Subscriber**
  - `subscriberId` (external eg: UserId, immutable)
  - `customerId` (scopes the subscriber to a tenant or business unit, immutable)
  - `email`, `phone`, `pushTokens` (map of appId → token), `webhookUrl`
  - `pushToken` (deprecated, use `pushTokens` instead)
  - `preferences` { channels: { email/sms/push/webhook: true/false }, language, allowedDays[], notificationWindow { start, end } }
  - `subscribedEventTypes: [string]` (opt-in event types)
  - `groups: [groupId]` (many-to-many membership)
  - `metadata` (optional)
  - `createdAt` (immutable)

- **Group**
  - `groupId` (immutable)
  - `customerId` (immutable)
  - `name`, `description`
  - `subscribers: [subscriberId]` (many-to-many)
  - `subscribedEventTypes: [string]` (opt-in event types)
  - `metadata` (optional)

- **Event**
  - `eventId`, `customerId`, `type`, `payload/data`
  - `recipients` { `subscriberIds?`, `groups?`, `broadcast?`, `directEmails?`, `directPhones?` } **— OPTIONAL**
    - If not provided, system uses rule defaults or interested subscribers/groups
  - `meta` (optional metadata)

- **Rule**
  - `eventType` (immutable)
  - `customerId` (immutable)
  - `channels[]`
  - `templateRefs` (map of channel → template ID)
  - `metadata` (optional, e.g., `appId` for push notifications)
  - `createdAt` (immutable), `updatedAt`

- **AppConfig** (Push Notifications)
  - `id` (app identifier, immutable)
  - `orgId` (immutable, links to organization)
  - `name` (app name)
  - `provider` ("apns", "firebase", or "mock")
  - `apns` (APNS configuration if provider is "apns")
  - `firebase` (Firebase configuration if provider is "firebase")
  - `createdAt` (immutable), `updatedAt`

- **Template**
  - `id` (immutable)
  - `customerId` (immutable)
  - `name`, `channel` (immutable), `version`
  - `content` (channel-specific: subject/body for email, body for SMS, title/body for push, payload for webhook)
  - `translations` (optional, map of language code → content)
  - `metadata` (optional)
  - `createdAt` (immutable), `updatedAt`

---

## 🔷 Typical Data Flow (Subscriber + Group Fanout)

1) Client system emits event via `POST /events` or Kafka (example payloads):

**With explicit recipients:**
```json
{
  "customerId":"acme-store",
  "type":"ORDER_DELAYED",
  "recipients":{
    "subscriberIds":["s1"],
    "groups":["operators","warehouse"],
    "broadcast":false
  },
  "payload":{"orderId":123}
}
```

**Without recipients (uses interested subscribers):**
```json
{
  "customerId":"acme-store",
  "type":"ORDER_DELAYED",
  "payload":{"orderId":123}
}
```

**With direct email/phone (no subscriber required):**
```json
{
  "customerId":"acme-store",
  "type":"ORDER_DELAYED",
  "recipients":{
    "directEmails":["customer@example.com"],
    "directPhones":["+1234567890"]
  },
  "payload":{"orderId":123}
}
```

2) Processor loads rule + resolves recipients:
   - If event has explicit recipients → use them
   - If event has no recipients → use subscribers/groups that have `subscribedEventTypes` matching the event type
3) Expand recipients: explicit `subscriberIds` + members of groups + direct emails/phones
4) Deduplicate subscriberIds
5) For each subscriber → expand per-channel tasks (respecting preferences and DND)
6) Workers send notifications + log status

---

## 🔷 Multi-Tenancy Notes

- **Hierarchical structure**: Organization → Customer → Subscribers/Groups/Templates/Rules
- All entities partitioned by `customerId` (nested under `orgId` for SaaS providers)
- Tenant isolation enforced at API, storage, and delivery pipeline
- Storage options: schema-per-tenant, table-per-tenant, or shared table with a `customerId` (and optional `orgId`) partition (start with shared table + partition)
- **Single-tenant model**: One organization with one customer
- **Multi-tenant SaaS model**: One organization managing multiple customers

---

## 🔷 Reliability & Ops

- Delivery semantics: at-least-once default; exactly-once optional with idempotent stores
- Quotas per-org and per-channel rate limiting
- Retries + DLQ + DLQ inspection UI
- Observability: Prometheus metrics, OpenTelemetry traces, structured logs with `customerId` and `orgId`
- Secrets stored in secret manager (Vault/AWS Secrets Manager)

---

## 🔷 Minimal Config Example

- `messageBroker.type: kafka`
- `subscriberStore.type: mongodb`
- `groupStore.type: mongodb`
- `organizationStore.type: mongodb`
- `customerStore.type: mongodb`
- `templateStore.type: mongodb`
- `ruleStore.type: mongodb`
- `retry.maxAttempts: 3`
- `retry.backoffBaseMs: 2000`
- `notifier.mail.provider: smtp-server`
- `notifier.sms.provider: sns`

---

## **Execution Plan (Improved & Actionable)**

This section breaks the work into concrete milestones, deliverables, acceptance criteria, and suggested timeline for a small team (2–4 engineers).

**Phase 0 — Foundations (1–2 weeks)**
- Deliverables: tech choices, API sketch, SPI definitions, data model (organization/customer/subscriber/group/event/rule)
- Acceptance: documented SPI interfaces, example `application.yml`, basic ERD, and a `README.md` for each service
- Tasks:
  - Decide language(s) and frameworks (Go/Java/Node) and message broker
  - Design REST routes and minimal JSON contracts for organizations, customers, subscribers, groups, events, templates, rules
  - Create repository skeletons (monorepo or multiple repos) and CI templates

**Phase 1 — MVP (3–6 weeks)**
- Goal: complete end-to-end flow for email notifications and admin UI to manage subscribers/rules
- Deliverables: running `notifyx-api`, `notifyx-processor`, `notifyx-worker` (email), and `notifyx-console` with minimal features
- Acceptance criteria:
  - `POST /events` ingests events and results in delivery attempts in worker logs
  - Organizations, customers, subscribers and groups CRUD working with persistent store
  - Templates can be created and used by rules
  - Events can be sent with or without explicit recipients (falls back to rule defaults or interested subscribers)
  - Delivery logs show per-`customerId` (and `orgId` when relevant) status
- Tasks (incremental):
  1. Implement `notifyx-api` endpoints for Organization, Customer, Subscriber, Group, Template, Rule, Event ingestion
  2. Implement `notifyx-processor` that reads events, resolves recipients (explicit → interested subscribers), expands recipients (subscriberIds/groups/broadcast/directEmails/directPhones), filters preferences, and enqueues tasks
  3. Implement `notifyx-worker` for email using SMTP/SendGrid with template rendering and retry logic (DLQ)
  4. Build minimal `notifyx-console` to manage organizations, customers, subscribers, groups, templates, and view logs
  5. Add unit tests and a simple integration test to exercise event→delivery path

**Phase 2 — Hardening (2–4 weeks)**
- Goal: make system production-ready for a pilot tenant
- Deliverables: Kafka ingestion, persistent stores (MongoDB/DynamoDB/Postgres), secrets integration, metrics, tracing
- Acceptance criteria:
  - Kafka consumer capable of high-throughput ingestion
  - Persistent storage of organizations, customers, subscribers/groups and delivery logs
  - Metrics (Prometheus) and traces (OpenTelemetry) visible for end-to-end flows
  - Retry policies, exponential backoff, DLQ persisted and inspectable
- Tasks:
  - Swap in Kafka for processor ingestion and SQS/Redis for delivery queues where appropriate
  - Implement persistent stores and migration scripts for organization, customer, and subscriber data
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
- End-to-end smoke tests: create organization/customer → create subscriber/group/rule → send event (with and without recipients) → assert delivery log
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

