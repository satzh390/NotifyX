package worker

import (
	"context"
	"fmt"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/workerx/template"
)

// GetTemplate retrieves a template from the store
func GetTemplate(
	ctx context.Context,
	templateStore storage.TemplateStore,
	customerID, templateRef string,
) (*domain.Template, error) {
	tpl, err := templateStore.Get(ctx, customerID, templateRef)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}
	return &tpl, nil
}

// GetLocalizedContent returns the appropriate template content based on language preference
func GetLocalizedContent(tpl *domain.Template, language string) domain.TemplateContent {
	if language == "" {
		language = DefaultLanguage
	}

	// Use default content if language is default or no translations available
	if language == DefaultLanguage || tpl.Translations == nil {
		return tpl.Content
	}

	// Try to get translated content
	if translated, ok := tpl.Translations[language]; ok {
		return translated
	}

	// Fallback to default content if translation not found
	return tpl.Content
}

// GetSubscriberLanguage extracts language preference from subscriber, with default fallback
func GetSubscriberLanguage(subscriber domain.Subscriber) string {
	if subscriber.Preferences.Language != "" {
		return subscriber.Preferences.Language
	}
	return DefaultLanguage
}

// RenderTemplate renders a template string with the given payload
func RenderTemplate(templateStr string, payload map[string]interface{}) (string, error) {
	return template.Render(templateStr, payload)
}

