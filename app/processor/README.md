## notifyx-processor

This service consumes events from Kafka, resolves recipients (subscribers/groups), applies rules, filters by preferences, and publishes delivery tasks to worker topics.

### Quick Start

1. **Start local infrastructure:**
   ```cmd
   docker compose -f ../../docker-compose.local.yaml up -d
   ```

2. **Run the processor:**
   ```cmd
   go run ./cmd
   ```

The processor will connect to Kafka and start consuming events from the configured input topic.

### Running locally

Edit `config/config.yaml` with your Kafka and MongoDB details, then run:

```
go run ./cmd
```

To load a different config path, set `NOTIFYX_PROCESSOR_CONFIG=/path/to/custom.yaml`.

### Configuration

Configuration is loaded from `config/config.yaml` by default. To use a different config file, set the `NOTIFYX_PROCESSOR_CONFIG` environment variable:

**Windows (Command Prompt):**
```cmd
set NOTIFYX_PROCESSOR_CONFIG=C:\path\to\custom.yaml
```

**Linux/Mac:**
```bash
export NOTIFYX_PROCESSOR_CONFIG=/path/to/custom.yaml
```

#### Overriding Config Values with Environment Variables

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

**Windows (Command Prompt):**
```cmd
REM Override Kafka brokers
set NOTIFYX_PROCESSOR_KAFKA__BROKERS=prod-kafka-1:9092,prod-kafka-2:9092

REM Override MongoDB URI
set NOTIFYX_PROCESSOR_STORAGE__MONGO__URI=mongodb://production:27017

REM Enable Redis cache
set NOTIFYX_PROCESSOR_CACHE__REDIS__ENABLED=true
set NOTIFYX_PROCESSOR_CACHE__REDIS__ADDR=redis://prod:6379

REM Run the processor
go run ./cmd
```

**Linux/Mac:**
```bash
# Override Kafka brokers
export NOTIFYX_PROCESSOR_KAFKA__BROKERS="prod-kafka-1:9092,prod-kafka-2:9092"

# Override MongoDB URI
export NOTIFYX_PROCESSOR_STORAGE__MONGO__URI="mongodb://production:27017"

# Enable Redis cache
export NOTIFYX_PROCESSOR_CACHE__REDIS__ENABLED="true"
export NOTIFYX_PROCESSOR_CACHE__REDIS__ADDR="redis://prod:6379"

# Run the processor
go run ./cmd
```

**Using a .env file:**

**Windows (Command Prompt):**
```cmd
REM Create .env file with:
REM NOTIFYX_PROCESSOR_KAFKA__BROKERS=prod-kafka-1:9092,prod-kafka-2:9092
REM NOTIFYX_PROCESSOR_STORAGE__MONGO__URI=mongodb://production:27017
REM NOTIFYX_PROCESSOR_CACHE__REDIS__ENABLED=true
REM NOTIFYX_PROCESSOR_CACHE__REDIS__ADDR=redis://prod:6379

REM Load and run (requires a .env loader or manual set commands)
```

**Linux/Mac:**
```bash
# .env file
NOTIFYX_PROCESSOR_KAFKA__BROKERS=prod-kafka-1:9092,prod-kafka-2:9092
NOTIFYX_PROCESSOR_STORAGE__MONGO__URI=mongodb://production:27017
NOTIFYX_PROCESSOR_CACHE__REDIS__ENABLED=true
NOTIFYX_PROCESSOR_CACHE__REDIS__ADDR=redis://prod:6379

# Load and run
source .env
go run ./cmd
```

**Important Notes:**
- Environment variables take precedence over config file values
- Use double underscores (`__`) to represent dots (`.`) in nested config paths
- Array values (like `brokers`) should be comma-separated strings
- Duration values (like `ttl`) should be in Go duration format (e.g., `5m`, `10s`, `1h`)
- The config file is still required - environment variables only override specific values

### How It Works

1. **Event Consumption**: Reads events from Kafka input topic
2. **Rule Resolution**: Looks up rules for the event type and customer
3. **Recipient Resolution**: 
   - If event has explicit recipients → uses them
   - If event has no recipients → uses subscribers/groups subscribed to the event type
4. **Filtering**: Applies subscriber preferences (disabled channels, DND windows, unsubscribed event types)
5. **Fanout**: Publishes delivery tasks to per-channel worker topics
6. **Error Handling**: Failed events are sent to DLQ topic

### Using the local docker stack

`docker-compose.local.yaml` spins up MongoDB, Kafka, Redis, LocalStack, and a mock OAuth server. After running:

```
docker compose -f ../../docker-compose.local.yaml up -d
```

configure `config/config.yaml` to use:
- Kafka brokers: `localhost:9092`
- MongoDB: `mongodb://localhost:27017`
- Redis: `localhost:6379` (if cache enabled)

The bundled config already points to the Compose services; update it if you run services elsewhere.

