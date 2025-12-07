//go:build integration
// +build integration

package integration

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/notifyx/core/domain"
	mongoadapter "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/worker-email/internal/provider"
	"github.com/notifyx/worker-email/internal/worker"
	"github.com/notifyx/workerx/delivery"
	workerlib "github.com/notifyx/workerx/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEmailWorkerIntegration(t *testing.T) (*worker.EmailWorker, func(), provider.Provider) {
	ctx := context.Background()
	
	// Connect to test MongoDB
	stores, cleanup, err := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})
	require.NoError(t, err)

	// Use simple mock provider for testing (no external SMTP server required)
	emailProvider := provider.NewMockEmailProvider()

	// Create base worker
	baseWorker := workerlib.NewBaseWorker(workerlib.BaseWorkerOptions{
		TemplateStore: stores.Templates,
		ResultHandler: delivery.NewMongoResultHandler(stores.DeliveryTasks, stores.DeliveryLogs),
		Channel:       domain.ChannelEmail,
		Logger:        slog.Default(),
	})

	// Create Email worker
	emailWorker := worker.NewEmailWorker(baseWorker, emailProvider)

	return emailWorker, func() {
		if cleanup != nil {
			_ = cleanup(ctx)
		}
	}, emailProvider
}

func TestEmailWorker_Integration(t *testing.T) {
	t.Run("process task with template", func(t *testing.T) {
		emailWorker, cleanup, _ := setupEmailWorkerIntegration(t)
		defer cleanup()

		ctx := context.Background()

		// Create a template in the database
		template := domain.Template{
			ID:         "test-template-1",
			CustomerID: "test-customer-1",
			Channel:    domain.ChannelEmail,
			Content: domain.TemplateContent{
				Subject: "Order {{.orderId}} Confirmation",
				Body:    "<p>Hello {{.name}}, your order {{.orderId}} is confirmed.</p>",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := emailWorker.GetTemplateStore().Put(ctx, template)
		require.NoError(t, err)

		// Create a delivery task
		task := domain.DeliveryTask{
			TaskID:      "test-task-1",
			CustomerID:  "test-customer-1",
			EventID:     "test-event-1",
			Channel:     domain.ChannelEmail,
			TemplateRef: "test-template-1",
			Subscriber: domain.Subscriber{
				ID:    "sub-1",
				Email: "test@example.com",
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
		err = emailWorker.ProcessTask(ctx, task)
		assert.NoError(t, err)
	})

	t.Run("process task with translation", func(t *testing.T) {
		emailWorker, cleanup, _ := setupEmailWorkerIntegration(t)
		defer cleanup()

		ctx := context.Background()

		// Create a template with translation
		template := domain.Template{
			ID:         "test-template-2",
			CustomerID: "test-customer-1",
			Channel:    domain.ChannelEmail,
			Content: domain.TemplateContent{
				Subject: "Hello {{.name}}",
				Body:    "<p>Hello {{.name}}</p>",
			},
			Translations: map[string]domain.TemplateContent{
				"es": {
					Subject: "Hola {{.name}}",
					Body:    "<p>Hola {{.name}}</p>",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := emailWorker.GetTemplateStore().Put(ctx, template)
		require.NoError(t, err)

		// Create a delivery task with Spanish preference
		task := domain.DeliveryTask{
			TaskID:      "test-task-2",
			CustomerID:  "test-customer-1",
			EventID:     "test-event-2",
			Channel:     domain.ChannelEmail,
			TemplateRef: "test-template-2",
			Subscriber: domain.Subscriber{
				ID:    "sub-2",
				Email: "test@example.com",
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
		err = emailWorker.ProcessTask(ctx, task)
		assert.NoError(t, err)
	})
}

