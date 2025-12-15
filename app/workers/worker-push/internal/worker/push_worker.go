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
	providerManager *provider.ProviderManager
}

// NewPushWorker creates a new push worker instance
func NewPushWorker(base *worker.BaseWorker, providerManager *provider.ProviderManager) *PushWorker {
	w := &PushWorker{
		BaseWorker:      base,
		providerManager: providerManager,
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

	// Extract appId from task metadata (set by processor from Rule.Metadata)
	appID := task.Metadata["appId"]
	if appID == "" {
		return fmt.Errorf("push: appId is required in task metadata (set in Rule.Metadata[\"appId\"])")
	}

	// Get provider for this app
	pushProvider, err := w.providerManager.GetProvider(ctx, task.CustomerID, appID)
	if err != nil {
		return fmt.Errorf("push: get provider: %w", err)
	}

	// Get push token for this app
	// Support both PushTokens map and legacy PushToken field for backward compatibility
	var pushToken string
	if task.Subscriber.PushTokens != nil {
		pushToken = task.Subscriber.PushTokens[appID]
	}
	// Fallback to legacy PushToken if PushTokens map doesn't have the appId
	if pushToken == "" {
		pushToken = task.Subscriber.PushToken
	}
	if pushToken == "" {
		return fmt.Errorf("push: subscriber has no push token for appId %s", appID)
	}

	// Send push notification
	if err := pushProvider.Send(ctx, pushToken, title, body, task.Metadata); err != nil {
		return fmt.Errorf("push: send: %w", err)
	}

	// Log success (truncate token for security)
	tokenDisplay := pushToken
	if len(tokenDisplay) > maxTokenDisplayLength {
		tokenDisplay = tokenDisplay[:maxTokenDisplayLength] + "..."
	}

	w.GetLogger().Info("push: sent successfully",
		slog.String("taskId", task.TaskID),
		slog.String("appId", appID),
		slog.String("pushToken", tokenDisplay),
	)

	return nil
}
