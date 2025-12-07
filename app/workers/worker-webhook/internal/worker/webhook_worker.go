package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/worker-webhook/internal/provider"
	"github.com/notifyx/workerx/worker"
)

// WebhookWorker processes webhook delivery tasks
type WebhookWorker struct {
	*worker.BaseWorker
	provider provider.Provider
}

// NewWebhookWorker creates a new webhook worker instance
func NewWebhookWorker(base *worker.BaseWorker, webhookProvider provider.Provider) *WebhookWorker {
	w := &WebhookWorker{
		BaseWorker: base,
		provider:   webhookProvider,
	}
	base.SetProcessTask(w.ProcessTask)
	return w
}

// ProcessTask processes a webhook delivery task
func (w *WebhookWorker) ProcessTask(ctx context.Context, task domain.DeliveryTask) error {
	// Get and process template
	tpl, err := worker.GetTemplate(ctx, w.GetTemplateStore(), task.CustomerID, task.TemplateRef)
	if err != nil {
		return fmt.Errorf("webhook: %w", err)
	}

	// Get localized content
	language := worker.GetSubscriberLanguage(task.Subscriber)
	content := worker.GetLocalizedContent(tpl, language)

	// Build webhook payload by merging template payload with task payload
	webhookPayload := w.buildWebhookPayload(content.Payload, task.Payload)

	// Render any template strings in the payload
	renderedPayload, err := w.renderPayload(webhookPayload, task.Payload)
	if err != nil {
		return fmt.Errorf("webhook: render payload: %w", err)
	}

	// Validate webhook URL
	if task.Subscriber.WebhookURL == "" {
		return fmt.Errorf("webhook: subscriber has no webhook URL")
	}

	// Send webhook
	if err := w.provider.Send(ctx, task.Subscriber.WebhookURL, renderedPayload, task.Metadata); err != nil {
		return fmt.Errorf("webhook: send: %w", err)
	}

	// Log success
	w.GetLogger().Info("webhook: sent successfully",
		slog.String("taskId", task.TaskID),
		slog.String("url", task.Subscriber.WebhookURL),
	)

	return nil
}

// buildWebhookPayload merges template payload with task payload
// Task payload values override template payload values
func (w *WebhookWorker) buildWebhookPayload(
	templatePayload map[string]interface{},
	taskPayload map[string]interface{},
) map[string]interface{} {
	result := make(map[string]interface{})

	// Start with template payload if it exists
	if templatePayload != nil {
		for k, v := range templatePayload {
			result[k] = v
		}
	}

	// Merge task payload (overrides template payload keys)
	for k, v := range taskPayload {
		result[k] = v
	}

	return result
}

// renderPayload renders any string values in the payload that might contain templates
func (w *WebhookWorker) renderPayload(
	payload map[string]interface{},
	templateData map[string]interface{},
) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for k, v := range payload {
		switch val := v.(type) {
		case string:
			// Try to render as template
			rendered, err := worker.RenderTemplate(val, templateData)
			if err != nil {
				// If rendering fails, use original value
				result[k] = val
			} else {
				result[k] = rendered
			}
		default:
			result[k] = v
		}
	}

	return result, nil
}

