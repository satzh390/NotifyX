package worker

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/mock"
)

// MockSMSProvider is a mock SMS provider for testing
type MockSMSProvider struct {
	mock.Mock
}

func (m *MockSMSProvider) Send(ctx context.Context, phone, message string, metadata map[string]string) error {
	args := m.Called(ctx, phone, message, metadata)
	return args.Error(0)
}

func (m *MockSMSProvider) Name() string {
	return "mock"
}

// MockTemplateStore is a mock template store for testing
type MockTemplateStore struct {
	mock.Mock
}

func (m *MockTemplateStore) Get(ctx context.Context, customerID, templateID string) (domain.Template, error) {
	args := m.Called(ctx, customerID, templateID)
	return args.Get(0).(domain.Template), args.Error(1)
}

func (m *MockTemplateStore) GetByLanguage(ctx context.Context, customerID, templateID, language string) (domain.Template, error) {
	args := m.Called(ctx, customerID, templateID, language)
	return args.Get(0).(domain.Template), args.Error(1)
}

func (m *MockTemplateStore) Put(ctx context.Context, tpl domain.Template) error {
	args := m.Called(ctx, tpl)
	return args.Error(0)
}

func (m *MockTemplateStore) Delete(ctx context.Context, customerID, templateID string) error {
	args := m.Called(ctx, customerID, templateID)
	return args.Error(0)
}

