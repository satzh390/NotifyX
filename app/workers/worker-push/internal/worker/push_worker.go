package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/worker-push/internal/provider"
	"github.com/notifyx/workerx/worker"
)

const (
	// maxTokenDisplayLength limits the length of push token in logs for security
	maxTokenDisplayLength = 20
)

// PushWorker processes push notification delivery tasks
type PushWorker struct {
	*worker.BaseWorker
	provider provider.Provider
}

// NewPushWorker creates a new push worker instance
func NewPushWorker(base *worker.BaseWorker, pushProvider provider.Provider) *PushWorker {
	w := &PushWorker{
		BaseWorker: base,
		provider:   pushProvider,
	}
	base.SetProcessTask(w.ProcessTask)
	return w
}

// ProcessTask processes a push notification delivery task
func (w *PushWorker) ProcessTask(ctx context.Context, task domain.DeliveryTask) error {
	// Get and process template
	tpl, err := worker.GetTemplate(ctx, w.GetTemplateStore(), task.CustomerID, task.TemplateRef)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}

	// Get localized content
	language := worker.GetSubscriberLanguage(task.Subscriber)
	content := worker.GetLocalizedContent(tpl, language)

	// Render template fields
	title, err := worker.RenderTemplate(content.Title, task.Payload)
	if err != nil {
		return fmt.Errorf("push: render title: %w", err)
	}

	body, err := worker.RenderTemplate(content.Body, task.Payload)
	if err != nil {
		return fmt.Errorf("push: render body: %w", err)
	}

	// Validate push token
	if task.Subscriber.PushToken == "" {
		return fmt.Errorf("push: subscriber has no push token")
	}

	// Send push notification
	if err := w.provider.Send(ctx, task.Subscriber.PushToken, title, body, task.Metadata); err != nil {
		return fmt.Errorf("push: send: %w", err)
	}

	// Log success (truncate token for security)
	tokenDisplay := task.Subscriber.PushToken
	if len(tokenDisplay) > maxTokenDisplayLength {
		tokenDisplay = tokenDisplay[:maxTokenDisplayLength] + "..."
	}

	w.GetLogger().Info("push: sent successfully",
		slog.String("taskId", task.TaskID),
		slog.String("pushToken", tokenDisplay),
	)

	return nil
}
