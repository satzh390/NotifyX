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

func (mockStore *MockTemplateStore) Put(ctx context.Context, template domain.Template) error {
	args := mockStore.Called(ctx, template)
	return args.Error(0)
}

func (mockStore *MockTemplateStore) Get(ctx context.Context, orgID, templateID string) (domain.Template, error) {
	args := mockStore.Called(ctx, orgID, templateID)
	return args.Get(0).(domain.Template), args.Error(1)
}

func (mockStore *MockTemplateStore) GetByLanguage(ctx context.Context, orgID, templateID, language string) (domain.Template, error) {
	args := mockStore.Called(ctx, orgID, templateID, language)
	return args.Get(0).(domain.Template), args.Error(1)
}

func (mockStore *MockTemplateStore) Delete(ctx context.Context, orgID, templateID string) error {
	args := mockStore.Called(ctx, orgID, templateID)
	return args.Error(0)
}
