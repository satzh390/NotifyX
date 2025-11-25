package pipeline

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/processor/internal/fanout"
	"github.com/notifyx/processor/internal/filter"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

// MockKafkaReader is a mock implementation of kafka.Reader
type MockKafkaReader struct {
	mock.Mock
}

func (m *MockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockKafkaWriter is a mock implementation of kafka.Writer
type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockResolver is a mock implementation of recipients.Resolver
type MockResolver struct {
	mock.Mock
}

func (m *MockResolver) Stream(ctx context.Context, customerID string, recipients domain.Recipients, visitor func(domain.Subscriber) error) error {
	args := m.Called(ctx, customerID, recipients, visitor)
	return args.Error(0)
}

// MockPublisher is a mock implementation of fanout.Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, envelopes []fanout.Envelope) error {
	args := m.Called(ctx, envelopes)
	return args.Error(0)
}

func (m *MockPublisher) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockPreferencesFilter is a mock implementation of filter.PreferencesFilter
type MockPreferencesFilter struct {
	mock.Mock
}

func (m *MockPreferencesFilter) Apply(subscribers []domain.Subscriber, rule domain.Rule) []filter.FilteredSubscriber {
	args := m.Called(subscribers, rule)
	return args.Get(0).([]filter.FilteredSubscriber)
}

// MockStores is a mock implementation of storage.Stores
type MockStores struct {
	mock.Mock
	Rules         storage.RuleStore
	Subscribers   storage.SubscriberStore
	Groups        storage.GroupStore
	Templates     storage.TemplateStore
	Organizations storage.OrganizationStore
	Customers     storage.CustomerStore
}

// MockRuleStore is a mock implementation of storage.RuleStore
type MockRuleStore struct {
	mock.Mock
}

func (m *MockRuleStore) Put(ctx context.Context, rule domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleStore) Get(ctx context.Context, customerID, eventType string) (domain.Rule, error) {
	args := m.Called(ctx, customerID, eventType)
	return args.Get(0).(domain.Rule), args.Error(1)
}

func (m *MockRuleStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Rule], error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Rule]), args.Error(1)
}

func (m *MockRuleStore) Delete(ctx context.Context, customerID, eventType string) error {
	args := m.Called(ctx, customerID, eventType)
	return args.Error(0)
}

// MockSubscriberCache is a mock implementation of cache.SubscriberCache
type MockSubscriberCache struct {
	mock.Mock
}

func (m *MockSubscriberCache) Get(ctx context.Context, customerID, subscriberID string) (domain.Subscriber, bool) {
	args := m.Called(ctx, customerID, subscriberID)
	return args.Get(0).(domain.Subscriber), args.Bool(1)
}

func (m *MockSubscriberCache) Set(ctx context.Context, subscriber domain.Subscriber) error {
	args := m.Called(ctx, subscriber)
	return args.Error(0)
}
