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

func (mockStore *MockCustomerStore) Put(ctx context.Context, customer domain.Customer) error {
	args := mockStore.Called(ctx, customer)
	return args.Error(0)
}

func (mockStore *MockCustomerStore) Get(ctx context.Context, customerID string) (domain.Customer, error) {
	args := mockStore.Called(ctx, customerID)
	return args.Get(0).(domain.Customer), args.Error(1)
}

func (mockStore *MockCustomerStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Customer], error) {
	args := mockStore.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Customer]), args.Error(1)
}

func (mockStore *MockCustomerStore) Delete(ctx context.Context, customerID string) error {
	args := mockStore.Called(ctx, customerID)
	return args.Error(0)
}

