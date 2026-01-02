package rule

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/mock"
)

// MockRuleStore is a mock implementation of storage.RuleStore
type MockRuleStore struct {
	mock.Mock
}

func (mockStore *MockRuleStore) Put(ctx context.Context, rule domain.Rule) error {
	args := mockStore.Called(ctx, rule)
	return args.Error(0)
}

func (mockStore *MockRuleStore) Get(ctx context.Context, customerID, eventType string) (domain.Rule, error) {
	args := mockStore.Called(ctx, customerID, eventType)
	return args.Get(0).(domain.Rule), args.Error(1)
}

func (mockStore *MockRuleStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Rule], error) {
	args := mockStore.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Rule]), args.Error(1)
}

func (mockStore *MockRuleStore) Delete(ctx context.Context, customerID, eventType string) error {
	args := mockStore.Called(ctx, customerID, eventType)
	return args.Error(0)
}
