package subscriber

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/mock"
)

// MockSubscriberStore is a mock implementation of storage.SubscriberStore
type MockSubscriberStore struct {
	mock.Mock
}

func (m *MockSubscriberStore) Put(ctx context.Context, sub domain.Subscriber) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriberStore) Get(ctx context.Context, orgID, subscriberID string) (domain.Subscriber, error) {
	args := m.Called(ctx, orgID, subscriberID)
	return args.Get(0).(domain.Subscriber), args.Error(1)
}

func (m *MockSubscriberStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Subscriber], error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Subscriber]), args.Error(1)
}

func (m *MockSubscriberStore) Delete(ctx context.Context, orgID, subscriberID string) error {
	args := m.Called(ctx, orgID, subscriberID)
	return args.Error(0)
}
