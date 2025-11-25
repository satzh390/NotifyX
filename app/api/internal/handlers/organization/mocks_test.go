package organization

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/mock"
)

// MockOrganizationStore is a mock implementation of storage.OrganizationStore
type MockOrganizationStore struct {
	mock.Mock
}

func (mockStore *MockOrganizationStore) Put(ctx context.Context, org domain.Organization) error {
	args := mockStore.Called(ctx, org)
	return args.Error(0)
}

func (mockStore *MockOrganizationStore) Get(ctx context.Context, orgID string) (domain.Organization, error) {
	args := mockStore.Called(ctx, orgID)
	return args.Get(0).(domain.Organization), args.Error(1)
}

func (mockStore *MockOrganizationStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Organization], error) {
	args := mockStore.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Organization]), args.Error(1)
}

func (mockStore *MockOrganizationStore) Delete(ctx context.Context, orgID string) error {
	args := mockStore.Called(ctx, orgID)
	return args.Error(0)
}
