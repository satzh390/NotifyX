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

func (mockStore *MockSubscriberStore) Put(ctx context.Context, sub domain.Subscriber) error {
	args := mockStore.Called(ctx, sub)
	return args.Error(0)
}

func (mockStore *MockSubscriberStore) Get(ctx context.Context, orgID, subscriberID string) (domain.Subscriber, error) {
	args := mockStore.Called(ctx, orgID, subscriberID)
	
	// Handle function return values
	if fn, ok := args.Get(0).(func(context.Context, string, string) domain.Subscriber); ok {
		return fn(ctx, orgID, subscriberID), args.Error(1)
	}
	
	return args.Get(0).(domain.Subscriber), args.Error(1)
}

func (mockStore *MockSubscriberStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Subscriber], error) {
	args := mockStore.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Subscriber]), args.Error(1)
}

func (mockStore *MockSubscriberStore) Delete(ctx context.Context, orgID, subscriberID string) error {
	args := mockStore.Called(ctx, orgID, subscriberID)
	return args.Error(0)
}
