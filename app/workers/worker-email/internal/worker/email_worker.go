package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/workerx/worker"
	"github.com/notifyx/worker-email/internal/provider"
)

// EmailWorker processes email delivery tasks
type EmailWorker struct {
	*worker.BaseWorker
	provider provider.Provider
}

// NewEmailWorker creates a new email worker instance
func NewEmailWorker(base *worker.BaseWorker, emailProvider provider.Provider) *EmailWorker {
	w := &EmailWorker{
		BaseWorker: base,
		provider:   emailProvider,
	}
	base.SetProcessTask(w.ProcessTask)
	return w
}

// ProcessTask processes an email delivery task
func (w *EmailWorker) ProcessTask(ctx context.Context, task domain.DeliveryTask) error {
	// Get and process template
	tpl, err := worker.GetTemplate(ctx, w.GetTemplateStore(), task.CustomerID, task.TemplateRef)
	if err != nil {
		return fmt.Errorf("email: %w", err)
	}

	// Get localized content
	language := worker.GetSubscriberLanguage(task.Subscriber)
	content := worker.GetLocalizedContent(tpl, language)

	// Render template fields
	subject, err := worker.RenderTemplate(content.Subject, task.Payload)
	if err != nil {
		return fmt.Errorf("email: render subject: %w", err)
	}

	body, err := worker.RenderTemplate(content.Body, task.Payload)
	if err != nil {
		return fmt.Errorf("email: render body: %w", err)
	}

	// Validate email address
	if task.Subscriber.Email == "" {
		return fmt.Errorf("email: subscriber has no email address")
	}

	// Send email
	if err := w.provider.Send(ctx, task.Subscriber.Email, subject, body, task.Metadata); err != nil {
		return fmt.Errorf("email: send: %w", err)
	}

	// Log success
	w.GetLogger().Info("email: sent successfully",
		slog.String("taskId", task.TaskID),
		slog.String("email", task.Subscriber.Email),
	)

	return nil
}

