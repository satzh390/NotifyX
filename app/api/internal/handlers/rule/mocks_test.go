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

func (m *MockRuleStore) Put(ctx context.Context, rule domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleStore) Get(ctx context.Context, orgID, eventType string) (domain.Rule, error) {
	args := m.Called(ctx, orgID, eventType)
	return args.Get(0).(domain.Rule), args.Error(1)
}

func (m *MockRuleStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Rule], error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.Rule]), args.Error(1)
}

func (m *MockRuleStore) Delete(ctx context.Context, orgID, eventType string) error {
	args := m.Called(ctx, orgID, eventType)
	return args.Error(0)
}
