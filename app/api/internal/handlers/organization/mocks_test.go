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

func (m *MockOrganizationStore) Put(ctx context.Context, org domain.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationStore) Get(ctx context.Context, orgID string) (domain.Organization, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(domain.Organization), args.Error(1)
}

func (m *MockOrganizationStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Organization], error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Organization]), args.Error(1)
}

func (m *MockOrganizationStore) Delete(ctx context.Context, orgID string) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}
