# NotifyX Processor Service

This service consumes events from Kafka, resolves recipients (subscribers/groups), applies rules, filters by preferences and custom filters, and publishes delivery tasks to worker topics.

## Quick Start

1. **Start local infrastructure:**
   ```bash
   docker compose -f ../../docker-compose.local.yaml up -d
   ```

2. **Run the processor:**
   ```bash
   go run ./cmd
   ```

The processor will connect to Kafka and start consuming events from the configured input topic.

## How It Works

1. **Event Consumption**: Reads events from Kafka input topic
2. **Rule Resolution**: Looks up rules for the event type and customer
3. **Recipient Resolution**: 
   - If event has explicit recipients → uses them
   - If event has no recipients → uses subscribers/groups subscribed to the event type
4. **Filtering**: Applies subscriber preferences (disabled channels, DND windows, unsubscribed event types) and optional custom filters
5. **Task Metadata**: For push notifications, extracts `appId` from `Rule.Metadata["appId"]` and includes it in `DeliveryTask.Metadata`
6. **Fanout**: Publishes delivery tasks to per-channel worker topics
7. **Error Handling**: Failed events are sent to DLQ topic

### Push Notification Multi-App Support

For push notifications, the processor extracts the `appId` from `Rule.Metadata["appId"]` and includes it in the `DeliveryTask.Metadata`. This allows the push worker to:
- Identify which app configuration to use
- Select the correct push token from `Subscriber.PushTokens[appId]`
- Route notifications to the appropriate provider

**Example Rule with App ID:**
```json
{
  "eventType": "order.created",
  "channels": ["push"],
  "templateRefs": {
    "push": "order-notification-template"
  },
  "metadata": {
    "appId": "my-ios-app"
  }
}
```

## Configuration

Configuration is loaded from `config/config.yaml` by default. To use a different config file, set the `NOTIFYX_PROCESSOR_CONFIG` environment variable:

**Windows:**
```cmd
set NOTIFYX_PROCESSOR_CONFIG=C:\path\to\custom.yaml
```

**Linux/Mac:**
```bash
export NOTIFYX_PROCESSOR_CONFIG=/path/to/custom.yaml
```

### Overriding Config Values with Environment Variables

Any configuration value can be overridden using environment variables. The pattern is:
- **Prefix**: `NOTIFYX_PROCESSOR_`
- **Key mapping**: Dots (`.`) in config keys become double underscores (`__`)
- **Case**: Environment variable names are case-sensitive (use uppercase)

**Examples:**

| Config File Path | Environment Variable | Example Value |
|-----------------|---------------------|---------------|
| `kafka.brokers` | `NOTIFYX_PROCESSOR_KAFKA__BROKERS` | `localhost:9092,localhost:9093` (comma-separated) |
| `kafka.groupId` | `NOTIFYX_PROCESSOR_KAFKA__GROUPID` | `notifyx-processor-prod` |
| `kafka.inputTopic` | `NOTIFYX_PROCESSOR_KAFKA__INPUTTOPIC` | `notifyx-events-prod` |
| `kafka.dlqTopic` | `NOTIFYX_PROCESSOR_KAFKA__DLQTOPIC` | `notifyx-events-dlq-prod` |
| `storage.mongo.uri` | `NOTIFYX_PROCESSOR_STORAGE__MONGO__URI` | `mongodb://prod:27017` |
| `storage.mongo.database` | `NOTIFYX_PROCESSOR_STORAGE__MONGO__DATABASE` | `notifyx_prod` |
| `cache.redis.enabled` | `NOTIFYX_PROCESSOR_CACHE__REDIS__ENABLED` | `true` |
| `cache.redis.addr` | `NOTIFYX_PROCESSOR_CACHE__REDIS__ADDR` | `redis://prod:6379` |
| `cache.redis.ttl` | `NOTIFYX_PROCESSOR_CACHE__REDIS__TTL` | `10m` |

**How to use:**

**Windows:**
```cmd
set NOTIFYX_PROCESSOR_KAFKA__BROKERS=prod-kafka-1:9092,prod-kafka-2:9092
set NOTIFYX_PROCESSOR_STORAGE__MONGO__URI=mongodb://production:27017
set NOTIFYX_PROCESSOR_CACHE__REDIS__ENABLED=true
go run ./cmd
```

**Linux/Mac:**
```bash
export NOTIFYX_PROCESSOR_KAFKA__BROKERS="prod-kafka-1:9092,prod-kafka-2:9092"
export NOTIFYX_PROCESSOR_STORAGE__MONGO__URI="mongodb://production:27017"
export NOTIFYX_PROCESSOR_CACHE__REDIS__ENABLED="true"
go run ./cmd
```

**Important Notes:**
- Environment variables take precedence over config file values
- Use double underscores (`__`) to represent dots (`.`) in nested config paths
- Array values (like `brokers`) should be comma-separated strings
- Duration values (like `ttl`) should be in Go duration format (e.g., `5m`, `10s`, `1h`)
- The config file is still required - environment variables only override specific values

## Custom Filters

The processor supports custom filters that allow you to implement business-specific filtering logic. Custom filters are automatically registered when the processor starts.

**Key Features:**
- Custom filters are applied in addition to the default preferences filter
- Custom filters are **NOT** applied to direct emails and phone numbers (only to registered subscribers)
- Filters are automatically discovered and registered from the `internal/filter/custom/` directory
- Each filter is registered using its struct type name

For detailed information on creating and using custom filters, see the [Custom Filters Documentation](internal/filter/custom/README.md).

## Using the Local Docker Stack

`docker-compose.local.yaml` spins up MongoDB, Kafka, Redis, LocalStack, and a mock OAuth server. After running:

```bash
docker compose -f ../../docker-compose.local.yaml up -d
```

Configure `config/config.yaml` to use:
- Kafka brokers: `localhost:9092`
- MongoDB: `mongodb://localhost:27017`
- Redis: `localhost:6379` (if cache enabled)

The bundled config already points to the Compose services; update it if you run services elsewhere.
