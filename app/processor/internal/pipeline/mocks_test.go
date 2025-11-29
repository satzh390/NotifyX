package pipeline

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/processor/internal/fanout"
	"github.com/notifyx/processor/internal/filter"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

// MockKafkaReader is a mock implementation of kafka.Reader
type MockKafkaReader struct {
	mock.Mock
}

func (mockReader *MockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := mockReader.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (mockReader *MockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := mockReader.Called(ctx, msgs)
	return args.Error(0)
}

func (mockReader *MockKafkaReader) Close() error {
	args := mockReader.Called()
	return args.Error(0)
}

// MockKafkaWriter is a mock implementation of kafka.Writer
type MockKafkaWriter struct {
	mock.Mock
}

func (mockWriter *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := mockWriter.Called(ctx, msgs)
	return args.Error(0)
}

func (mockWriter *MockKafkaWriter) Close() error {
	args := mockWriter.Called()
	return args.Error(0)
}

// MockResolver is a mock implementation of recipients.Resolver
type MockResolver struct {
	mock.Mock
}

func (mockResolver *MockResolver) Stream(ctx context.Context, customerID, eventType string, recipients domain.Recipients, visitor func(domain.Subscriber) error) error {
	args := mockResolver.Called(ctx, customerID, eventType, recipients, visitor)
	return args.Error(0)
}

// MockPublisher is a mock implementation of fanout.Publisher
type MockPublisher struct {
	mock.Mock
}

func (mockPublisher *MockPublisher) Publish(ctx context.Context, envelopes []fanout.Envelope) error {
	args := mockPublisher.Called(ctx, envelopes)
	return args.Error(0)
}

func (mockPublisher *MockPublisher) Close(ctx context.Context) error {
	args := mockPublisher.Called(ctx)
	return args.Error(0)
}

// MockPreferencesFilter is a mock implementation of filter.PreferencesFilter
type MockPreferencesFilter struct {
	mock.Mock
}

func (mockPreferencesFilter *MockPreferencesFilter) Apply(subscribers []domain.Subscriber, rule domain.Rule, message map[string]interface{}) []filter.FilteredSubscriber {
	args := mockPreferencesFilter.Called(subscribers, rule, message)
	return args.Get(0).([]filter.FilteredSubscriber)
}

// MockRuleStore is a mock implementation of storage.RuleStore
type MockRuleStore struct {
	mock.Mock
}

func (mockRuleStore *MockRuleStore) Put(ctx context.Context, rule domain.Rule) error {
	args := mockRuleStore.Called(ctx, rule)
	return args.Error(0)
}

func (mockRuleStore *MockRuleStore) Get(ctx context.Context, customerID, eventType string) (domain.Rule, error) {
	args := mockRuleStore.Called(ctx, customerID, eventType)
	return args.Get(0).(domain.Rule), args.Error(1)
}

func (mockRuleStore *MockRuleStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Rule], error) {
	args := mockRuleStore.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Rule]), args.Error(1)
}

func (mockRuleStore *MockRuleStore) Delete(ctx context.Context, customerID, eventType string) error {
	args := mockRuleStore.Called(ctx, customerID, eventType)
	return args.Error(0)
}

// MockSubscriberCache is a mock implementation of cache.SubscriberCache
type MockSubscriberCache struct {
	mock.Mock
}

func (mockSubscriberCache *MockSubscriberCache) Get(ctx context.Context, customerID, subscriberID string) (domain.Subscriber, bool) {
	args := mockSubscriberCache.Called(ctx, customerID, subscriberID)
	return args.Get(0).(domain.Subscriber), args.Bool(1)
}

func (mockSubscriberCache *MockSubscriberCache) Set(ctx context.Context, subscriber domain.Subscriber) error {
	args := mockSubscriberCache.Called(ctx, subscriber)
	return args.Error(0)
}
