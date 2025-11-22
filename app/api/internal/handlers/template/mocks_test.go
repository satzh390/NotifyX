package template

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/mock"
)

// MockTemplateStore is a mock implementation of storage.TemplateStore
type MockTemplateStore struct {
	mock.Mock
}

func (m *MockTemplateStore) Put(ctx context.Context, template domain.Template) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockTemplateStore) Get(ctx context.Context, orgID, templateID string) (domain.Template, error) {
	args := m.Called(ctx, orgID, templateID)
	return args.Get(0).(domain.Template), args.Error(1)
}

func (m *MockTemplateStore) GetByLanguage(ctx context.Context, orgID, templateID, language string) (domain.Template, error) {
	args := m.Called(ctx, orgID, templateID, language)
	return args.Get(0).(domain.Template), args.Error(1)
}

func (m *MockTemplateStore) Delete(ctx context.Context, orgID, templateID string) error {
	args := m.Called(ctx, orgID, templateID)
	return args.Error(0)
}
