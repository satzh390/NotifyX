# NotifyX Worker Library

Shared library for NotifyX worker implementations. This library provides common functionality that all channel-specific workers use.

## Components

### Worker Base (`worker/base.go`)
- Base worker implementation with Kafka message consumption
- Task processing framework
- Delivery logging
- DLQ (Dead Letter Queue) support

### Delivery Handlers (`delivery/`)
- **MongoResultHandler**: Stores delivery tasks and logs in MongoDB
- **BrokerResultHandler**: Publishes delivery tasks and logs to Kafka
- **CompositeResultHandler**: Combines multiple handlers

### Template Rendering (`template/render.go`)
- Template rendering with Go template syntax
- Payload data injection

### Configuration (`config/config.go`)
- Base configuration structure
- Kafka, MongoDB, and delivery configuration
- Environment variable support

## Usage

Each channel-specific worker (SMS, Email, Push) imports this library and:

1. Creates a `BaseWorker` with channel-specific configuration
2. Implements channel-specific task processing
3. Uses shared delivery handlers for result storage
4. Uses shared template rendering

## Example

```go
import (
    "github.com/notifyx/workerx/worker"
    "github.com/notifyx/workerx/delivery"
    "github.com/notifyx/workerx/template"
)

// Create base worker
baseWorker := worker.NewBaseWorker(worker.BaseWorkerOptions{
    Reader:        kafkaReader,
    DLQ:           dlqWriter,
    TemplateStore: templateStore,
    ResultHandler: resultHandler,
    Logger:        logger,
    Channel:       domain.ChannelSMS,
})

// Set custom processing
baseWorker.SetProcessTask(func(ctx context.Context, task domain.DeliveryTask) error {
    // Channel-specific processing
    return nil
})
```

