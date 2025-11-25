package group

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/mock"
)

// MockGroupStore is a mock implementation of storage.GroupStore
type MockGroupStore struct {
	mock.Mock
}

func (mockStore *MockGroupStore) Put(ctx context.Context, group domain.Group) error {
	args := mockStore.Called(ctx, group)
	return args.Error(0)
}

func (mockStore *MockGroupStore) Get(ctx context.Context, customerID, groupID string) (domain.Group, error) {
	args := mockStore.Called(ctx, customerID, groupID)
	return args.Get(0).(domain.Group), args.Error(1)
}

func (mockStore *MockGroupStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Group], error) {
	args := mockStore.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Group]), args.Error(1)
}

func (mockStore *MockGroupStore) Delete(ctx context.Context, customerID, groupID string) error {
	args := mockStore.Called(ctx, customerID, groupID)
	return args.Error(0)
}
