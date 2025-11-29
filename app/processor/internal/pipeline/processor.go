package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/resolver"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/processor/internal/event"
	"github.com/notifyx/processor/internal/fanout"
	"github.com/notifyx/processor/internal/filter"
)

type (
	KafkaReader interface {
		FetchMessage(ctx context.Context) (kafka.Message, error)
		CommitMessages(ctx context.Context, msgs ...kafka.Message) error
		Close() error
	}

	KafkaWriter interface {
		WriteMessages(ctx context.Context, msgs ...kafka.Message) error
		Close() error
	}

	RecipientResolver interface {
		Stream(ctx context.Context, customerID, eventType string, recipients domain.Recipients, visitor func(domain.Subscriber) error) error
	}

	SubscriberFilter interface {
		Apply(subscribers []domain.Subscriber, rule domain.Rule, message map[string]interface{}) []filter.FilteredSubscriber
	}
)

type Processor struct {
	reader       KafkaReader
	dlq          KafkaWriter
	resolver     RecipientResolver
	ruleResolver *resolver.RuleResolver
	filters      SubscriberFilter
	fanout       fanout.Publisher
	stores       storage.Stores
	logger       *slog.Logger
}

type Options struct {
	Reader       KafkaReader            `validate:"required"`
	DLQ          KafkaWriter            // Optional
	Resolver     RecipientResolver      `validate:"required"`
	RuleResolver *resolver.RuleResolver `validate:"required"`
	Filter       SubscriberFilter       `validate:"required"`
	Fanout       fanout.Publisher       `validate:"required"`
	Stores       storage.Stores         // Optional
	Logger       *slog.Logger           // Optional, defaults to slog.Default()
}

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func NewProcessor(opts Options) (*Processor, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, fmt.Errorf("processor: invalid options: %w", err)
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Processor{
		reader:       opts.Reader,
		dlq:          opts.DLQ,
		resolver:     opts.Resolver,
		ruleResolver: opts.RuleResolver,
		filters:      opts.Filter,
		fanout:       opts.Fanout,
		stores:       opts.Stores,
		logger:       logger,
	}, nil
}

const (
	subscriberChunkSize = 500
)

func (processor *Processor) Run(ctx context.Context) error {
	for {
		msg, err := processor.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("processor: fetch message: %w", err)
		}

		if err := processor.handleMessage(ctx, msg); err != nil {
			processor.logger.Error("processor: handle message failed", slog.String("error", err.Error()))
			if dlqErr := processor.publishDLQ(ctx, msg, err); dlqErr != nil {
				processor.logger.Error("processor: publish to DLQ failed", slog.String("error", dlqErr.Error()))
			}
		} else {
			if err := processor.reader.CommitMessages(ctx, msg); err != nil {
				processor.logger.Error("processor: commit failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (processor *Processor) handleMessage(ctx context.Context, msg kafka.Message) error {
	env, err := event.ParseEnvelope(msg.Value)
	if err != nil {
		return err
	}

	// Use RuleResolver to get merged rule (global + customer override)
	rule, err := processor.ruleResolver.Resolve(ctx, env.CustomerID, env.Type)
	if err != nil {
		return fmt.Errorf("rule lookup: %w", err)
	}

	recipients := env.Recipients
	if err := validate.Struct(recipients); err != nil {
		return fmt.Errorf("recipient limits: %w", err)
	}

	processed := map[string]struct{}{}
	chunk := make([]domain.Subscriber, 0, subscriberChunkSize)
	totalFanout := 0

	flushChunk := func() error {
		if len(chunk) == 0 {
			return nil
		}
		count, err := processor.processChunk(ctx, chunk, rule, env, env.CustomerID)
		if err != nil {
			return err
		}
		totalFanout += count
		chunk = chunk[:0]
		return nil
	}

	err = processor.resolver.Stream(ctx, env.CustomerID, env.Type, recipients, func(subscriber domain.Subscriber) error {
		// For direct recipients (empty ID), use email/phone for deduplication
		// For regular subscribers, use ID
		key := getSubscriberIdentifier(subscriber)
		if key == "unknown" {
			return nil // Skip invalid direct recipient
		}
		if _, exists := processed[key]; exists {
			return nil
		}
		processed[key] = struct{}{}

		chunk = append(chunk, subscriber)
		if len(chunk) >= subscriberChunkSize {
			return flushChunk()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("recipients stream: %w", err)
	}

	if err := flushChunk(); err != nil {
		return err
	}

	if totalFanout == 0 {
		processor.logger.Warn("processor: no eligible subscribers", slog.String("eventId", env.EventID()))
	}
	return nil
}

func (processor *Processor) publishDLQ(ctx context.Context, msg kafka.Message, procErr error) error {
	if processor.dlq == nil {
		return nil
	}

	payload := map[string]any{
		"value":   string(msg.Value),
		"error":   procErr.Error(),
		"topic":   msg.Topic,
		"offset":  msg.Offset,
		"headers": msg.Headers,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return processor.dlq.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: body,
	})
}

func (processor *Processor) processChunk(ctx context.Context, subscribers []domain.Subscriber, rule domain.Rule, env event.CloudEventEnvelope[map[string]any], customerID string) (int, error) {
	// Pass the event payload as the message for custom filters
	message := env.Payload
	if message == nil {
		message = make(map[string]interface{})
	}
	filtered := processor.filters.Apply(subscribers, rule, message)
	if len(filtered) == 0 {
		return 0, nil
	}

	envelopes := make([]fanout.Envelope, 0, len(filtered))
	for _, entry := range filtered {
		for _, channel := range entry.Channels {
			templateRef, ok := rule.TemplateRefs[channel]
			if !ok {
				continue
			}
			subscriberID := getSubscriberIdentifier(entry.Subscriber)
			taskID, idempotencyKey := generateTaskID(env.EventID(), customerID, subscriberID, channel)
			task := domain.DeliveryTask{
				TaskID:         taskID,
				IdempotencyKey: idempotencyKey,
				EventID:        env.EventID(),
				CustomerID:     customerID,
				Subscriber:     entry.Subscriber,
				Channel:        channel,
				TemplateRef:    templateRef,
				Payload:        env.Payload,
				Metadata:       env.Meta,
			}
			envelopes = append(envelopes, fanout.Envelope{Channel: channel, Task: task})
		}
	}

	if len(envelopes) == 0 {
		return 0, nil
	}

	if err := processor.fanout.Publish(ctx, envelopes); err != nil {
		return 0, fmt.Errorf("fanout: %w", err)
	}
	return len(envelopes), nil
}

// getSubscriberIdentifier returns a unique identifier for the subscriber.
// For regular subscribers, returns the subscriber ID.
// For direct recipients (empty ID), returns the email or phone number directly.
func getSubscriberIdentifier(subscriber domain.Subscriber) string {
	if subscriber.ID != "" {
		return subscriber.ID
	}
	if subscriber.Email != "" {
		return subscriber.Email
	}
	if subscriber.Phone != "" {
		return subscriber.Phone
	}
	// Fallback (should not happen in practice)
	return "unknown"
}

// generateTaskID creates a deterministic TaskID and idempotency key
// Format: eventId-customerId-subId-channel
// The idempotency key is a SHA256 hash of the same components for shorter storage
func generateTaskID(eventID, customerID, subscriberID string, channel domain.ChannelType) (taskID, idempotencyKey string) {
	// Create deterministic TaskID: eventId:customerId:subId:channel
	taskID = fmt.Sprintf("%s:%s:%s:%s", eventID, customerID, subscriberID, channel)

	// Create idempotency key as hash for shorter storage
	hash := sha256.Sum256([]byte(taskID))
	idempotencyKey = hex.EncodeToString(hash[:])

	return taskID, idempotencyKey
}
