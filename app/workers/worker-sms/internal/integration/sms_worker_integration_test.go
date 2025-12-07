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
	"github.com/notifyx/worker-sms/internal/provider"
	"github.com/notifyx/worker-sms/internal/worker"
	"github.com/notifyx/workerx/delivery"
	workerlib "github.com/notifyx/workerx/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSMSWorkerIntegration(t *testing.T) (*worker.SMSWorker, func(), provider.Provider) {
	ctx := context.Background()

	// Connect to test MongoDB
	stores, cleanup, err := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})
	require.NoError(t, err)

	// Create SNS provider with LocalStack endpoint
	snsProvider, err := provider.NewSNSProvider(provider.SNSConfig{
		Region:    "us-east-1",
		AccessKey: "test",
		SecretKey: "test",
		Endpoint:  "http://localhost:4566", // LocalStack endpoint
	})
	require.NoError(t, err)

	// Create base worker
	baseWorker := workerlib.NewBaseWorker(workerlib.BaseWorkerOptions{
		TemplateStore: stores.Templates,
		ResultHandler: delivery.NewMongoResultHandler(stores.DeliveryTasks, stores.DeliveryLogs),
		Channel:       domain.ChannelSMS,
		Logger:        slog.Default(),
	})

	// Create SMS worker
	smsWorker := worker.NewSMSWorker(baseWorker, snsProvider)

	return smsWorker, func() {
		if cleanup != nil {
			_ = cleanup(ctx)
		}
	}, snsProvider
}

func TestSMSWorker_Integration(t *testing.T) {
	t.Run("process task with template", func(t *testing.T) {
		smsWorker, cleanup, _ := setupSMSWorkerIntegration(t)
		defer cleanup()

		ctx := context.Background()

		// Create a template in the database
		template := domain.Template{
			ID:         "test-template-1",
			CustomerID: "test-customer-1",
			Channel:    domain.ChannelSMS,
			Content: domain.TemplateContent{
				Body: "Hello {{.name}}, your order {{.orderId}} is ready",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := smsWorker.GetTemplateStore().Put(ctx, template)
		require.NoError(t, err)

		// Create a delivery task
		task := domain.DeliveryTask{
			TaskID:      "test-task-1",
			CustomerID:  "test-customer-1",
			EventID:     "test-event-1",
			Channel:     domain.ChannelSMS,
			TemplateRef: "test-template-1",
			Subscriber: domain.Subscriber{
				ID:    "sub-1",
				Phone: "+1234567890",
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

		// Process the task - should succeed with LocalStack SNS
		err = smsWorker.ProcessTask(ctx, task)
		assert.NoError(t, err)
	})

	t.Run("process task with translation", func(t *testing.T) {
		smsWorker, cleanup, _ := setupSMSWorkerIntegration(t)
		defer cleanup()

		ctx := context.Background()

		// Create a template with translation
		template := domain.Template{
			ID:         "test-template-2",
			CustomerID: "test-customer-1",
			Channel:    domain.ChannelSMS,
			Content: domain.TemplateContent{
				Body: "Hello {{.name}}",
			},
			Translations: map[string]domain.TemplateContent{
				"es": {
					Body: "Hola {{.name}}",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := smsWorker.GetTemplateStore().Put(ctx, template)
		require.NoError(t, err)

		// Create a delivery task with Spanish preference
		task := domain.DeliveryTask{
			TaskID:      "test-task-2",
			CustomerID:  "test-customer-1",
			EventID:     "test-event-2",
			Channel:     domain.ChannelSMS,
			TemplateRef: "test-template-2",
			Subscriber: domain.Subscriber{
				ID:    "sub-2",
				Phone: "+1234567890",
				Preferences: domain.SubscriberPrefs{
					Language: "es",
				},
			},
			Payload: map[string]interface{}{
				"name": "Juan",
			},
			CreatedAt: time.Now(),
		}

		// Process the task - should succeed with LocalStack SNS
		err = smsWorker.ProcessTask(ctx, task)
		assert.NoError(t, err)
	})
}
