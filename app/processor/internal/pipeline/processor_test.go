package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/processor/internal/cache"
	"github.com/notifyx/processor/internal/event"
	"github.com/notifyx/processor/internal/fanout"
	"github.com/notifyx/processor/internal/filter"
	"github.com/notifyx/processor/internal/recipients"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestProcessor() (*Processor, *MockKafkaReader, *MockKafkaWriter, *MockResolver, *MockPublisher, *MockPreferencesFilter, *MockStores) {
	reader := new(MockKafkaReader)
	dlq := new(MockKafkaWriter)
	resolver := new(MockResolver)
	publisher := new(MockPublisher)
	filter := new(MockPreferencesFilter)
	stores := &MockStores{
		Rules: new(MockRuleStore),
	}

	processor := NewProcessor(Options{
		Reader:   reader,
		DLQ:      dlq,
		Resolver: resolver,
		Filter:   filter,
		Fanout:   publisher,
		Stores:   stores,
	})

	return processor, reader, dlq, resolver, publisher, filter, stores
}

func TestProcessor_HandleMessage_Success(t *testing.T) {
	processor, reader, _, resolver, publisher, filter, stores := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"
	rule := domain.Rule{
		EventType:  eventType,
		CustomerID: customerID,
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
		DefaultRecipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
	}

	subscriber := domain.Subscriber{
		ID:         "sub-1",
		CustomerID: customerID,
		Email:      "test@example.com",
		Preferences: domain.SubscriberPrefs{
			Channels: map[domain.ChannelType]bool{
				domain.ChannelEmail: true,
			},
		},
	}

	envelope := event.CloudEventEnvelope[map[string]any]{
		ID:         "event-1",
		CustomerID: customerID,
		Type:       eventType,
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
		Payload: map[string]any{"orderId": "123"},
		Time:    time.Now(),
	}
	envelopeJSON, _ := json.Marshal(envelope)

	msg := kafka.Message{
		Value: envelopeJSON,
	}

	stores.Rules.(*MockRuleStore).On("Get", mock.Anything, customerID, eventType).Return(rule, nil).Once()
	resolver.On("Stream", mock.Anything, customerID, mock.Anything, mock.AnythingOfType("func(domain.Subscriber) error")).
		Run(func(args mock.Arguments) {
			visitor := args.Get(3).(func(domain.Subscriber) error)
			_ = visitor(subscriber)
		}).Return(nil).Once()

	filter.On("Apply", mock.MatchedBy(func(subs []domain.Subscriber) bool {
		return len(subs) == 1 && subs[0].ID == "sub-1"
	}), rule).Return([]filter.FilteredSubscriber{
		{
			Subscriber: subscriber,
			Channels:   []domain.ChannelType{domain.ChannelEmail},
		},
	}).Once()

	publisher.On("Publish", mock.Anything, mock.MatchedBy(func(envs []fanout.Envelope) bool {
		return len(envs) == 1 && envs[0].Channel == domain.ChannelEmail
	})).Return(nil).Once()

	reader.On("CommitMessages", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.NoError(t, err)
	stores.Rules.(*MockRuleStore).AssertExpectations(t)
	resolver.AssertExpectations(t)
	filter.AssertExpectations(t)
	publisher.AssertExpectations(t)
	reader.AssertExpectations(t)
}

func TestProcessor_HandleMessage_InvalidEnvelope(t *testing.T) {
	processor, reader, _, _, _, _, _ := setupTestProcessor()

	msg := kafka.Message{
		Value: []byte("invalid json"),
	}

	err := processor.handleMessage(context.Background(), msg)

	assert.Error(t, err)
	reader.AssertNotCalled(t, "CommitMessages")
}

func TestProcessor_HandleMessage_MissingCustomerID(t *testing.T) {
	processor, reader, _, _, _, _, _ := setupTestProcessor()

	envelope := event.CloudEventEnvelope[map[string]any]{
		ID:         "event-1",
		CustomerID: "", // Missing customerID
		Type:       "order.created",
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
		Time: time.Now(),
	}
	envelopeJSON, _ := json.Marshal(envelope)

	msg := kafka.Message{
		Value: envelopeJSON,
	}

	err := processor.handleMessage(context.Background(), msg)

	assert.Error(t, err)
	reader.AssertNotCalled(t, "CommitMessages")
}

func TestProcessor_HandleMessage_RuleNotFound(t *testing.T) {
	processor, reader, dlq, _, _, _, stores := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"

	envelope := event.CloudEventEnvelope[map[string]any]{
		ID:         "event-1",
		CustomerID: customerID,
		Type:       eventType,
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
		Time: time.Now(),
	}
	envelopeJSON, _ := json.Marshal(envelope)

	msg := kafka.Message{
		Value: envelopeJSON,
	}

	stores.Rules.(*MockRuleStore).On("Get", mock.Anything, customerID, eventType).Return(domain.Rule{}, storage.ErrNotFound).Once()
	dlq.On("WriteMessages", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.Error(t, err)
	stores.Rules.(*MockRuleStore).AssertExpectations(t)
	dlq.AssertExpectations(t)
	reader.AssertNotCalled(t, "CommitMessages")
}

func TestProcessor_HandleMessage_NoEligibleSubscribers(t *testing.T) {
	processor, reader, _, resolver, publisher, filter, stores := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"
	rule := domain.Rule{
		EventType:  eventType,
		CustomerID: customerID,
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
		DefaultRecipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
	}

	subscriber := domain.Subscriber{
		ID:         "sub-1",
		CustomerID: customerID,
		Email:      "test@example.com",
		Preferences: domain.SubscriberPrefs{
			Channels: map[domain.ChannelType]bool{
				domain.ChannelEmail: false, // Disabled
			},
		},
	}

	envelope := event.CloudEventEnvelope[map[string]any]{
		ID:         "event-1",
		CustomerID: customerID,
		Type:       eventType,
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
		Payload: map[string]any{"orderId": "123"},
		Time:    time.Now(),
	}
	envelopeJSON, _ := json.Marshal(envelope)

	msg := kafka.Message{
		Value: envelopeJSON,
	}

	stores.Rules.(*MockRuleStore).On("Get", mock.Anything, customerID, eventType).Return(rule, nil).Once()
	resolver.On("Stream", mock.Anything, customerID, mock.Anything, mock.AnythingOfType("func(domain.Subscriber) error")).
		Run(func(args mock.Arguments) {
			visitor := args.Get(3).(func(domain.Subscriber) error)
			_ = visitor(subscriber)
		}).Return(nil).Once()

	filter.On("Apply", mock.Anything, rule).Return([]filter.FilteredSubscriber{}).Once()

	reader.On("CommitMessages", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.NoError(t, err)
	publisher.AssertNotCalled(t, "Publish")
	reader.AssertExpectations(t)
}

func TestProcessor_HandleMessage_ResolverError(t *testing.T) {
	processor, reader, _, resolver, _, _, stores := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"
	rule := domain.Rule{
		EventType:  eventType,
		CustomerID: customerID,
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
		DefaultRecipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
	}

	envelope := event.CloudEventEnvelope[map[string]any]{
		ID:         "event-1",
		CustomerID: customerID,
		Type:       eventType,
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
		Time: time.Now(),
	}
	envelopeJSON, _ := json.Marshal(envelope)

	msg := kafka.Message{
		Value: envelopeJSON,
	}

	stores.Rules.(*MockRuleStore).On("Get", mock.Anything, customerID, eventType).Return(rule, nil).Once()
	resolver.On("Stream", mock.Anything, customerID, mock.Anything, mock.AnythingOfType("func(domain.Subscriber) error")).
		Return(errors.New("resolver error")).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.Error(t, err)
	reader.AssertNotCalled(t, "CommitMessages")
}

func TestProcessor_HandleMessage_PublisherError(t *testing.T) {
	processor, reader, dlq, resolver, publisher, filter, stores := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"
	rule := domain.Rule{
		EventType:  eventType,
		CustomerID: customerID,
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
		DefaultRecipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
	}

	subscriber := domain.Subscriber{
		ID:         "sub-1",
		CustomerID: customerID,
		Email:      "test@example.com",
		Preferences: domain.SubscriberPrefs{
			Channels: map[domain.ChannelType]bool{
				domain.ChannelEmail: true,
			},
		},
	}

	envelope := event.CloudEventEnvelope[map[string]any]{
		ID:         "event-1",
		CustomerID: customerID,
		Type:       eventType,
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
		Payload: map[string]any{"orderId": "123"},
		Time:    time.Now(),
	}
	envelopeJSON, _ := json.Marshal(envelope)

	msg := kafka.Message{
		Value: envelopeJSON,
	}

	stores.Rules.(*MockRuleStore).On("Get", mock.Anything, customerID, eventType).Return(rule, nil).Once()
	resolver.On("Stream", mock.Anything, customerID, mock.Anything, mock.AnythingOfType("func(domain.Subscriber) error")).
		Run(func(args mock.Arguments) {
			visitor := args.Get(3).(func(domain.Subscriber) error)
			_ = visitor(subscriber)
		}).Return(nil).Once()

	filter.On("Apply", mock.Anything, rule).Return([]filter.FilteredSubscriber{
		{
			Subscriber: subscriber,
			Channels:   []domain.ChannelType{domain.ChannelEmail},
		},
	}).Once()

	publisher.On("Publish", mock.Anything, mock.Anything).Return(errors.New("publisher error")).Once()
	dlq.On("WriteMessages", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.Error(t, err)
	dlq.AssertExpectations(t)
	reader.AssertNotCalled(t, "CommitMessages")
}

func TestGenerateTaskID(t *testing.T) {
	eventID := "event-123"
	customerID := "customer-456"
	subscriberID := "sub-789"
	channel := domain.ChannelEmail

	taskID, idempotencyKey := generateTaskID(eventID, customerID, subscriberID, channel)

	expectedTaskID := "event-123:customer-456:sub-789:email"
	assert.Equal(t, expectedTaskID, taskID)
	assert.NotEmpty(t, idempotencyKey)
	assert.Len(t, idempotencyKey, 64) // SHA256 hex string length

	// Test idempotency - same inputs should produce same outputs
	taskID2, idempotencyKey2 := generateTaskID(eventID, customerID, subscriberID, channel)
	assert.Equal(t, taskID, taskID2)
	assert.Equal(t, idempotencyKey, idempotencyKey2)
}

func TestGetSubscriberIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		subscriber domain.Subscriber
		expected  string
	}{
		{
			name: "subscriber with ID",
			subscriber: domain.Subscriber{
				ID: "sub-123",
			},
			expected: "sub-123",
		},
		{
			name: "direct email recipient",
			subscriber: domain.Subscriber{
				ID:    "",
				Email: "test@example.com",
			},
			expected: "test@example.com",
		},
		{
			name: "direct phone recipient",
			subscriber: domain.Subscriber{
				ID:    "",
				Phone: "+1234567890",
			},
			expected: "+1234567890",
		},
		{
			name: "unknown recipient",
			subscriber: domain.Subscriber{
				ID:    "",
				Email: "",
				Phone: "",
			},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSubscriberIdentifier(tt.subscriber)
			assert.Equal(t, tt.expected, result)
		})
	}
}

