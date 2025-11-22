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

func (m *MockGroupStore) Put(ctx context.Context, group domain.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupStore) Get(ctx context.Context, orgID, groupID string) (domain.Group, error) {
	args := m.Called(ctx, orgID, groupID)
	return args.Get(0).(domain.Group), args.Error(1)
}

func (m *MockGroupStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Group], error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Group]), args.Error(1)
}

func (m *MockGroupStore) Delete(ctx context.Context, orgID, groupID string) error {
	args := m.Called(ctx, orgID, groupID)
	return args.Error(0)
}
