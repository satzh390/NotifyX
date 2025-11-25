package customer

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/mock"
)

// MockCustomerStore is a mock implementation of storage.CustomerStore
type MockCustomerStore struct {
	mock.Mock
}

func (m *MockCustomerStore) Put(ctx context.Context, customer domain.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerStore) Get(ctx context.Context, customerID string) (domain.Customer, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).(domain.Customer), args.Error(1)
}

func (m *MockCustomerStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Customer], error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Customer]), args.Error(1)
}

func (m *MockCustomerStore) Delete(ctx context.Context, customerID string) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

