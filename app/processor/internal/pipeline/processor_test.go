package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/notifyx/core/domain"
	resolverpkg "github.com/notifyx/core/resolver"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/processor/internal/event"
	"github.com/notifyx/processor/internal/fanout"
	filterpkg "github.com/notifyx/processor/internal/filter"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestProcessor() (*Processor, *MockKafkaReader, *MockKafkaWriter, *MockResolver, *MockPublisher, *MockPreferencesFilter, *MockRuleStore) {
	reader := new(MockKafkaReader)
	dlq := new(MockKafkaWriter)
	resolverMock := new(MockResolver)
	publisher := new(MockPublisher)
	filter := new(MockPreferencesFilter)
	ruleStore := new(MockRuleStore)

	ruleResolver := resolverpkg.NewRuleResolver(resolverpkg.Options{
		Store: ruleStore,
		Cache: resolverpkg.NoopRuleCache{},
	})

	processor := NewProcessor(Options{
		Reader:       reader,
		DLQ:          dlq,
		Resolver:     resolverMock,
		Filter:       filter,
		Fanout:       publisher,
		RuleResolver: ruleResolver,
		Stores: storage.Stores{
			Rules: ruleStore,
		},
	})

	return processor, reader, dlq, resolverMock, publisher, filter, ruleStore
}

func TestProcessor_HandleMessage_Success(t *testing.T) {
	processor, _, _, resolver, publisher, filter, ruleStore := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"
	rule := domain.Rule{
		EventType:  eventType,
		CustomerID: customerID,
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
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
		ID:          "event-1",
		Source:      "test-source",
		SpecVersion: "1.0",
		CustomerID:  customerID,
		Type:        eventType,
		Data:        map[string]any{"orderId": "123"},
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

	// RuleResolver first tries to load global rule (customerID = ""), then customer-specific rule
	ruleStore.On("Get", mock.Anything, "", eventType).Return(domain.Rule{}, storage.ErrNotFound).Once()
	ruleStore.On("Get", mock.Anything, customerID, eventType).Return(rule, nil).Once()
	resolver.On("Stream", mock.Anything, customerID, eventType, mock.Anything, mock.AnythingOfType("func(domain.Subscriber) error")).
		Run(func(args mock.Arguments) {
			visitor := args.Get(4).(func(domain.Subscriber) error)
			_ = visitor(subscriber)
		}).Return(nil).Once()

	filter.On("Apply", mock.MatchedBy(func(subs []domain.Subscriber) bool {
		return len(subs) == 1 && subs[0].ID == "sub-1"
	}), rule, mock.Anything).Return([]filterpkg.FilteredSubscriber{
		{
			Subscriber: subscriber,
			Channels:   []domain.ChannelType{domain.ChannelEmail},
		},
	}).Once()

	publisher.On("Publish", mock.Anything, mock.MatchedBy(func(envs []fanout.Envelope) bool {
		return len(envs) == 1 && envs[0].Channel == domain.ChannelEmail
	})).Return(nil).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.NoError(t, err)
	ruleStore.AssertExpectations(t)
	resolver.AssertExpectations(t)
	filter.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestProcessor_HandleMessage_InvalidEnvelope(t *testing.T) {
	processor, _, _, _, _, _, _ := setupTestProcessor()

	msg := kafka.Message{
		Value: []byte("invalid json"),
	}

	err := processor.handleMessage(context.Background(), msg)

	assert.Error(t, err)
}

func TestProcessor_HandleMessage_MissingCustomerID(t *testing.T) {
	processor, _, _, _, _, _, _ := setupTestProcessor()

	envelope := event.CloudEventEnvelope[map[string]any]{
		ID:          "event-1",
		Source:      "test-source",
		SpecVersion: "1.0",
		CustomerID:  "", // Missing customerID
		Type:        "order.created",
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
}

func TestProcessor_HandleMessage_RuleNotFound(t *testing.T) {
	processor, _, _, _, _, _, ruleStore := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"

	envelope := event.CloudEventEnvelope[map[string]any]{
		ID:          "event-1",
		Source:      "test-source",
		SpecVersion: "1.0",
		CustomerID:  customerID,
		Type:        eventType,
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
		Time: time.Now(),
	}
	envelopeJSON, _ := json.Marshal(envelope)

	msg := kafka.Message{
		Value: envelopeJSON,
	}

	// RuleResolver first tries to load global rule (customerID = ""), then customer-specific rule
	ruleStore.On("Get", mock.Anything, "", eventType).Return(domain.Rule{}, storage.ErrNotFound).Once()
	ruleStore.On("Get", mock.Anything, customerID, eventType).Return(domain.Rule{}, storage.ErrNotFound).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.Error(t, err)
	ruleStore.AssertExpectations(t)
}

func TestProcessor_HandleMessage_NoEligibleSubscribers(t *testing.T) {
	processor, _, _, resolver, publisher, filter, ruleStore := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"
	rule := domain.Rule{
		EventType:  eventType,
		CustomerID: customerID,
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
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
		ID:          "event-1",
		Source:      "test-source",
		SpecVersion: "1.0",
		CustomerID:  customerID,
		Type:        eventType,
		Data:        map[string]any{"orderId": "123"},
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

	// RuleResolver first tries to load global rule (customerID = ""), then customer-specific rule
	ruleStore.On("Get", mock.Anything, "", eventType).Return(domain.Rule{}, storage.ErrNotFound).Once()
	ruleStore.On("Get", mock.Anything, customerID, eventType).Return(rule, nil).Once()
	resolver.On("Stream", mock.Anything, customerID, eventType, mock.Anything, mock.AnythingOfType("func(domain.Subscriber) error")).
		Run(func(args mock.Arguments) {
			visitor := args.Get(4).(func(domain.Subscriber) error)
			_ = visitor(subscriber)
		}).Return(nil).Once()

	filter.On("Apply", mock.Anything, rule, mock.Anything).Return([]filterpkg.FilteredSubscriber{}).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.NoError(t, err)
	publisher.AssertNotCalled(t, "Publish")
}

func TestProcessor_HandleMessage_ResolverError(t *testing.T) {
	processor, _, _, resolver, _, _, ruleStore := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"
	rule := domain.Rule{
		EventType:  eventType,
		CustomerID: customerID,
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
		},
	}

	envelope := event.CloudEventEnvelope[map[string]any]{
		ID:          "event-1",
		Source:      "test-source",
		SpecVersion: "1.0",
		CustomerID:  customerID,
		Type:        eventType,
		Recipients: domain.Recipients{
			SubscriberIDs: []string{"sub-1"},
		},
		Time: time.Now(),
	}
	envelopeJSON, _ := json.Marshal(envelope)

	msg := kafka.Message{
		Value: envelopeJSON,
	}

	// RuleResolver first tries to load global rule (customerID = ""), then customer-specific rule
	ruleStore.On("Get", mock.Anything, "", eventType).Return(domain.Rule{}, storage.ErrNotFound).Once()
	ruleStore.On("Get", mock.Anything, customerID, eventType).Return(rule, nil).Once()
	resolver.On("Stream", mock.Anything, customerID, eventType, mock.Anything, mock.AnythingOfType("func(domain.Subscriber) error")).
		Return(errors.New("resolver error")).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.Error(t, err)
}

func TestProcessor_HandleMessage_PublisherError(t *testing.T) {
	processor, _, _, resolver, publisher, filter, ruleStore := setupTestProcessor()

	customerID := "test-customer"
	eventType := "order.created"
	rule := domain.Rule{
		EventType:  eventType,
		CustomerID: customerID,
		Channels:   []domain.ChannelType{domain.ChannelEmail},
		TemplateRefs: map[domain.ChannelType]string{
			domain.ChannelEmail: "template-1",
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
		ID:          "event-1",
		Source:      "test-source",
		SpecVersion: "1.0",
		CustomerID:  customerID,
		Type:        eventType,
		Data:        map[string]any{"orderId": "123"},
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

	// RuleResolver first tries to load global rule (customerID = ""), then customer-specific rule
	ruleStore.On("Get", mock.Anything, "", eventType).Return(domain.Rule{}, storage.ErrNotFound).Once()
	ruleStore.On("Get", mock.Anything, customerID, eventType).Return(rule, nil).Once()
	resolver.On("Stream", mock.Anything, customerID, eventType, mock.Anything, mock.AnythingOfType("func(domain.Subscriber) error")).
		Run(func(args mock.Arguments) {
			visitor := args.Get(4).(func(domain.Subscriber) error)
			_ = visitor(subscriber)
		}).Return(nil).Once()

	filter.On("Apply", mock.Anything, rule, mock.Anything).Return([]filterpkg.FilteredSubscriber{
		{
			Subscriber: subscriber,
			Channels:   []domain.ChannelType{domain.ChannelEmail},
		},
	}).Once()

	publisher.On("Publish", mock.Anything, mock.Anything).Return(errors.New("publisher error")).Once()

	err := processor.handleMessage(context.Background(), msg)

	assert.Error(t, err)
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
		name       string
		subscriber domain.Subscriber
		expected   string
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
