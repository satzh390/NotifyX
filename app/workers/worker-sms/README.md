# NotifyX SMS Worker

This service consumes SMS delivery tasks from Kafka, renders templates, and sends SMS notifications via configured providers (AWS SNS).

## Quick Start

1. **Start local infrastructure:**
   ```bash
   docker compose -f ../../../docker-compose.local.yaml up -d
   ```

2. **Run the SMS worker:**
   ```bash
   go run ./cmd
   ```

The worker will connect to Kafka and start consuming SMS delivery tasks from the configured topic (`notifyx-worker-sms`).

## How It Works

1. **Task Consumption**: Reads delivery tasks from Kafka topic `notifyx-worker-sms`
2. **Template Retrieval**: Loads template from MongoDB based on template reference
3. **Template Rendering**: Renders template with event payload and subscriber language preferences
4. **SMS Delivery**: Sends SMS via configured provider (AWS SNS)
5. **Result Logging**: Stores delivery results (success/failure) to MongoDB and/or Kafka
6. **Error Handling**: Failed tasks are sent to DLQ topic

## Configuration

Configuration is loaded from `config/config.yaml` by default. To use a different config file, set the `NOTIFYX_WORKER_SMS_CONFIG` environment variable:

**Windows:**
```cmd
set NOTIFYX_WORKER_SMS_CONFIG=C:\path\to\custom.yaml
```

**Linux/Mac:**
```bash
export NOTIFYX_WORKER_SMS_CONFIG=/path/to/custom.yaml
```

### Overriding Config Values with Environment Variables

Any configuration value can be overridden using environment variables. The pattern is:
- **Prefix**: `NOTIFYX_WORKER_SMS_`
- **Key mapping**: Dots (`.`) in config keys become double underscores (`__`)
- **Case**: Environment variable names are case-sensitive (use uppercase)

**Examples:**

| Config File Path | Environment Variable | Example Value |
|-----------------|---------------------|---------------|
| `kafka.brokers` | `NOTIFYX_WORKER_SMS_KAFKA__BROKERS` | `localhost:9092,localhost:9093` (comma-separated) |
| `kafka.topic` | `NOTIFYX_WORKER_SMS_KAFKA__TOPIC` | `notifyx-worker-sms-prod` |
| `storage.mongo.uri` | `NOTIFYX_WORKER_SMS_STORAGE__MONGO__URI` | `mongodb://prod:27017` |
| `sms.provider.sns.region` | `NOTIFYX_WORKER_SMS_SMS__PROVIDER__SNS__REGION` | `us-west-2` |
| `sms.provider.sns.accessKey` | `NOTIFYX_WORKER_SMS_SMS__PROVIDER__SNS__ACCESSKEY` | `AKIAIOSFODNN7EXAMPLE` |

**How to use:**

**Windows:**
```cmd
set NOTIFYX_WORKER_SMS_KAFKA__BROKERS=prod-kafka-1:9092,prod-kafka-2:9092
set NOTIFYX_WORKER_SMS_STORAGE__MONGO__URI=mongodb://production:27017
set NOTIFYX_WORKER_SMS_SMS__PROVIDER__SNS__REGION=us-west-2
go run ./cmd
```

**Linux/Mac:**
```bash
export NOTIFYX_WORKER_SMS_KAFKA__BROKERS="prod-kafka-1:9092,prod-kafka-2:9092"
export NOTIFYX_WORKER_SMS_STORAGE__MONGO__URI="mongodb://production:27017"
export NOTIFYX_WORKER_SMS_SMS__PROVIDER__SNS__REGION="us-west-2"
go run ./cmd
```

**Important Notes:**
- Environment variables take precedence over config file values
- Use double underscores (`__`) to represent dots (`.`) in nested config paths
- Array values (like `brokers`) should be comma-separated strings
- The config file is still required - environment variables only override specific values

## SMS Providers

### AWS SNS

The worker supports AWS SNS for sending SMS messages.

**Configuration:**
- `region`: AWS region (e.g., `us-east-1`)
- `accessKey`: AWS access key (optional, uses default credentials if empty)
- `secretKey`: AWS secret key (optional, uses default credentials if empty)

If `accessKey` and `secretKey` are not provided, the worker will use default AWS credentials (IAM role, environment variables, or credentials file).

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
- AWS SNS: Use LocalStack endpoint (if testing locally)

The bundled config already points to the Compose services; update it if you run services elsewhere.

