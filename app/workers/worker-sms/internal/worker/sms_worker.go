package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/workerx/worker"
	"github.com/notifyx/worker-sms/internal/provider"
)

// SMSWorker processes SMS delivery tasks
type SMSWorker struct {
	*worker.BaseWorker
	provider provider.Provider
}

// NewSMSWorker creates a new SMS worker instance
func NewSMSWorker(base *worker.BaseWorker, smsProvider provider.Provider) *SMSWorker {
	w := &SMSWorker{
		BaseWorker: base,
		provider:   smsProvider,
	}
	base.SetProcessTask(w.ProcessTask)
	return w
}

// ProcessTask processes an SMS delivery task
func (w *SMSWorker) ProcessTask(ctx context.Context, task domain.DeliveryTask) error {
	// Get and process template
	tpl, err := worker.GetTemplate(ctx, w.GetTemplateStore(), task.CustomerID, task.TemplateRef)
	if err != nil {
		return fmt.Errorf("sms: %w", err)
	}

	// Get localized content
	language := worker.GetSubscriberLanguage(task.Subscriber)
	content := worker.GetLocalizedContent(tpl, language)

	// Render template body
	body, err := worker.RenderTemplate(content.Body, task.Payload)
	if err != nil {
		return fmt.Errorf("sms: render template: %w", err)
	}

	// Validate phone number
	if task.Subscriber.Phone == "" {
		return fmt.Errorf("sms: subscriber has no phone number")
	}

	// Send SMS
	if err := w.provider.Send(ctx, task.Subscriber.Phone, body, task.Metadata); err != nil {
		return fmt.Errorf("sms: send: %w", err)
	}

	// Log success
	w.GetLogger().Info("sms: sent successfully",
		slog.String("taskId", task.TaskID),
		slog.String("phone", task.Subscriber.Phone),
	)

	return nil
}

