//go:build integration
// +build integration

package integration

import (
	"context"
	"log/slog"
	"testing"
	"time"

	mongoadapter "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/worker-webhook/internal/provider"
	"github.com/notifyx/worker-webhook/internal/worker"
	"github.com/notifyx/workerx/delivery"
	workerlib "github.com/notifyx/workerx/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWebhookWorkerIntegration(t *testing.T) (*worker.WebhookWorker, func(), provider.Provider) {
	ctx := context.Background()

	// Connect to test MongoDB
	stores, cleanup, err := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})
	require.NoError(t, err)

	// Use simple mock provider for testing (no external webhook receiver required)
	webhookProvider := provider.NewMockWebhookProvider()

	// Create base worker
	baseWorker := workerlib.NewBaseWorker(workerlib.BaseWorkerOptions{
		TemplateStore: stores.Templates,
		ResultHandler: delivery.NewMongoResultHandler(stores.DeliveryTasks, stores.DeliveryLogs),
		Channel:       domain.ChannelWebhook,
		Logger:        slog.Default(),
	})

	// Create Webhook worker
	webhookWorker := worker.NewWebhookWorker(baseWorker, webhookProvider)

	return webhookWorker, func() {
		if cleanup != nil {
			_ = cleanup(ctx)
		}
	}, webhookProvider
}

func TestWebhookWorker_Integration(t *testing.T) {
	t.Run("process task with template", func(t *testing.T) {
		webhookWorker, cleanup, _ := setupWebhookWorkerIntegration(t)
		defer cleanup()

		ctx := context.Background()

		// Create a template in the database
		template := domain.Template{
			ID:         "test-template-1",
			CustomerID: "test-customer-1",
			Channel:    domain.ChannelWebhook,
			Content: domain.TemplateContent{
				Payload: map[string]interface{}{
					"event":    "order_ready",
					"message":  "Hello {{.name}}, your order {{.orderId}} is ready",
					"orderId":  "{{.orderId}}",
					"customer": "{{.name}}",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := webhookWorker.GetTemplateStore().Put(ctx, template)
		require.NoError(t, err)

		// Create a delivery task
		task := domain.DeliveryTask{
			TaskID:      "test-task-1",
			CustomerID:  "test-customer-1",
			EventID:     "test-event-1",
			Channel:     domain.ChannelWebhook,
			TemplateRef: "test-template-1",
			Subscriber: domain.Subscriber{
				ID:         "sub-1",
				WebhookURL: "http://localhost:8888/test-webhook",
				Preferences: domain.SubscriberPrefs{
					Language: "en",
				},
			},
			Payload: map[string]interface{}{
				"name":    "John",
				"orderId": "12345",
			},
			CreatedAt: time.Now(),
		}

		// Process the task - should succeed with mock provider
		err = webhookWorker.ProcessTask(ctx, task)
		assert.NoError(t, err)
	})

	t.Run("process task with translation", func(t *testing.T) {
		webhookWorker, cleanup, _ := setupWebhookWorkerIntegration(t)
		defer cleanup()

		ctx := context.Background()

		// Create a template with translation
		template := domain.Template{
			ID:         "test-template-2",
			CustomerID: "test-customer-1",
			Channel:    domain.ChannelWebhook,
			Content: domain.TemplateContent{
				Payload: map[string]interface{}{
					"message": "Hello {{.name}}",
					"event":   "greeting",
				},
			},
			Translations: map[string]domain.TemplateContent{
				"es": {
					Payload: map[string]interface{}{
						"message": "Hola {{.name}}",
						"event":   "saludo",
					},
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := webhookWorker.GetTemplateStore().Put(ctx, template)
		require.NoError(t, err)

		// Create a delivery task with Spanish preference
		task := domain.DeliveryTask{
			TaskID:      "test-task-2",
			CustomerID:  "test-customer-1",
			EventID:     "test-event-2",
			Channel:     domain.ChannelWebhook,
			TemplateRef: "test-template-2",
			Subscriber: domain.Subscriber{
				ID:         "sub-2",
				WebhookURL: "http://localhost:8888/test-webhook/es",
				Preferences: domain.SubscriberPrefs{
					Language: "es",
				},
			},
			Payload: map[string]interface{}{
				"name": "Juan",
			},
			CreatedAt: time.Now(),
		}

		// Process the task - should succeed with mock provider
		err = webhookWorker.ProcessTask(ctx, task)
		assert.NoError(t, err)
	})
}

