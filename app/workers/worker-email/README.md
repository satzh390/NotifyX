# NotifyX Email Worker

This service consumes email delivery tasks from Kafka, renders templates, and sends email notifications via configured providers (SMTP).

## Quick Start

1. **Start local infrastructure:**
   ```bash
   docker compose -f ../../../docker-compose.local.yaml up -d
   ```

2. **Run the email worker:**
   ```bash
   go run ./cmd
   ```

The worker will connect to Kafka and start consuming email delivery tasks from the configured topic (`notifyx-worker-email`).

## How It Works

1. **Task Consumption**: Reads delivery tasks from Kafka topic `notifyx-worker-email`
2. **Template Retrieval**: Loads template from MongoDB based on template reference
3. **Template Rendering**: Renders template (subject and body) with event payload and subscriber language preferences
4. **Email Delivery**: Sends email via configured provider (SMTP)
5. **Result Logging**: Stores delivery results (success/failure) to MongoDB and/or Kafka
6. **Error Handling**: Failed tasks are sent to DLQ topic

## Configuration

Configuration is loaded from `config/config.yaml` by default. To use a different config file, set the `NOTIFYX_WORKER_EMAIL_CONFIG` environment variable:

**Windows:**
```cmd
set NOTIFYX_WORKER_EMAIL_CONFIG=C:\path\to\custom.yaml
```

**Linux/Mac:**
```bash
export NOTIFYX_WORKER_EMAIL_CONFIG=/path/to/custom.yaml
```

### Overriding Config Values with Environment Variables

Any configuration value can be overridden using environment variables. The pattern is:
- **Prefix**: `NOTIFYX_WORKER_EMAIL_`
- **Key mapping**: Dots (`.`) in config keys become double underscores (`__`)
- **Case**: Environment variable names are case-sensitive (use uppercase)

**Examples:**

| Config File Path | Environment Variable | Example Value |
|-----------------|---------------------|---------------|
| `kafka.brokers` | `NOTIFYX_WORKER_EMAIL_KAFKA__BROKERS` | `localhost:9092,localhost:9093` (comma-separated) |
| `kafka.topic` | `NOTIFYX_WORKER_EMAIL_KAFKA__TOPIC` | `notifyx-worker-email-prod` |
| `storage.mongo.uri` | `NOTIFYX_WORKER_EMAIL_STORAGE__MONGO__URI` | `mongodb://prod:27017` |
| `email.provider.smtp.host` | `NOTIFYX_WORKER_EMAIL_EMAIL__PROVIDER__SMTP__HOST` | `smtp.gmail.com` |
| `email.provider.smtp.port` | `NOTIFYX_WORKER_EMAIL_EMAIL__PROVIDER__SMTP__PORT` | `587` |

**How to use:**

**Windows:**
```cmd
set NOTIFYX_WORKER_EMAIL_KAFKA__BROKERS=prod-kafka-1:9092,prod-kafka-2:9092
set NOTIFYX_WORKER_EMAIL_STORAGE__MONGO__URI=mongodb://production:27017
set NOTIFYX_WORKER_EMAIL_EMAIL__PROVIDER__SMTP__HOST=smtp.gmail.com
go run ./cmd
```

**Linux/Mac:**
```bash
export NOTIFYX_WORKER_EMAIL_KAFKA__BROKERS="prod-kafka-1:9092,prod-kafka-2:9092"
export NOTIFYX_WORKER_EMAIL_STORAGE__MONGO__URI="mongodb://production:27017"
export NOTIFYX_WORKER_EMAIL_EMAIL__PROVIDER__SMTP__HOST="smtp.gmail.com"
go run ./cmd
```

**Important Notes:**
- Environment variables take precedence over config file values
- Use double underscores (`__`) to represent dots (`.`) in nested config paths
- Array values (like `brokers`) should be comma-separated strings
- The config file is still required - environment variables only override specific values

## Email Providers

### SMTP

The worker supports SMTP for sending emails.

**Configuration:**
- `host`: SMTP server hostname (e.g., `smtp.gmail.com`)
- `port`: SMTP server port (default: `587`)
- `username`: SMTP username
- `password`: SMTP password
- `from`: From email address

## Delivery Result Handling

The worker can store delivery results in multiple ways:

1. **MongoDB** (`delivery.mode: "mongo"`): Stores delivery tasks and logs in MongoDB collections
2. **Kafka Broker** (`delivery.mode: "broker"`): Publishes delivery tasks and logs to Kafka topics
3. **Both** (`delivery.mode: "both"`): Stores in both MongoDB and Kafka
4. **None** (`delivery.mode: "none"`): Skips storing results (only processes and sends)

## Using the Local Docker Stack

`docker-compose.local.yaml` spins up MongoDB, Kafka, LocalStack, and a mock OAuth server. After running:

```bash
docker compose -f ../../../docker-compose.local.yaml up -d
```

Configure `config/config.yaml` to use:
- Kafka brokers: `localhost:9092`
- MongoDB: `mongodb://localhost:27017`

The bundled config already points to the Compose services; update it if you run services elsewhere.

