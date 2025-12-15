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
	"github.com/notifyx/worker-push/internal/provider"
	"github.com/notifyx/worker-push/internal/worker"
	"github.com/notifyx/workerx/delivery"
	workerlib "github.com/notifyx/workerx/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPushWorkerIntegration(t *testing.T) (*worker.PushWorker, func(), *provider.ProviderManager) {
	ctx := context.Background()

	// Connect to test MongoDB
	stores, cleanup, err := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})
	require.NoError(t, err)

	// Create a test organization and customer
	orgID := "test-org-1"
	customerID := "test-customer-1"
	customer := domain.Customer{
		ID:     customerID,
		OrgID:  orgID,
		Name:   "Test Customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = stores.Customers.Put(ctx, customer)
	require.NoError(t, err)

	// Create a test AppConfig with mock provider
	appID := "test-app-1"
	appConfig := domain.AppConfig{
		ID:        appID,
		OrgID:     orgID,
		Name:      "Test App",
		Provider:  domain.PushProviderMock,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = stores.AppConfigs.Put(ctx, appConfig)
	require.NoError(t, err)

	// Create ProviderManager
	providerManager := provider.NewProviderManager(stores.AppConfigs, stores.Customers, slog.Default())

	// Create base worker
	baseWorker := workerlib.NewBaseWorker(workerlib.BaseWorkerOptions{
		TemplateStore: stores.Templates,
		ResultHandler: delivery.NewMongoResultHandler(stores.DeliveryTasks, stores.DeliveryLogs),
		Channel:       domain.ChannelPush,
		Logger:        slog.Default(),
	})

	// Create Push worker
	pushWorker := worker.NewPushWorker(baseWorker, providerManager)

	return pushWorker, func() {
		if cleanup != nil {
			_ = cleanup(ctx)
		}
	}, providerManager
}

func TestPushWorker_Integration(t *testing.T) {
	t.Run("process task with template", func(t *testing.T) {
		pushWorker, cleanup, _ := setupPushWorkerIntegration(t)
		defer cleanup()

		ctx := context.Background()

		// Create a template in the database
		template := domain.Template{
			ID:         "test-template-1",
			CustomerID: "test-customer-1",
			Channel:    domain.ChannelPush,
			Content: domain.TemplateContent{
				Title: "Order {{.orderId}} Update",
				Body:  "Hello {{.name}}, your order {{.orderId}} is ready for pickup",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := pushWorker.GetTemplateStore().Put(ctx, template)
		require.NoError(t, err)

		// Create a delivery task with appId in metadata and PushTokens map
		task := domain.DeliveryTask{
			TaskID:      "test-task-1",
			CustomerID:  "test-customer-1",
			EventID:     "test-event-1",
			Channel:     domain.ChannelPush,
			TemplateRef: "test-template-1",
			Subscriber: domain.Subscriber{
				ID: "sub-1",
				PushTokens: map[string]string{
					"test-app-1": "fcm-token-123456789",
				},
				Preferences: domain.SubscriberPrefs{
					Language: "en",
				},
			},
			Metadata: map[string]string{
				"appId": "test-app-1",
			},
			Payload: map[string]interface{}{
				"name":    "John",
				"orderId": "12345",
			},
			CreatedAt: time.Now(),
		}

		// Process the task - should succeed with mock provider
		err = pushWorker.ProcessTask(ctx, task)
		assert.NoError(t, err)
	})

	t.Run("process task with translation", func(t *testing.T) {
		pushWorker, cleanup, _ := setupPushWorkerIntegration(t)
		defer cleanup()

		ctx := context.Background()

		// Create a template with translation
		template := domain.Template{
			ID:         "test-template-2",
			CustomerID: "test-customer-1",
			Channel:    domain.ChannelPush,
			Content: domain.TemplateContent{
				Title: "Hello {{.name}}",
				Body:  "Hello {{.name}}, welcome!",
			},
			Translations: map[string]domain.TemplateContent{
				"es": {
					Title: "Hola {{.name}}",
					Body:  "Hola {{.name}}, bienvenido!",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := pushWorker.GetTemplateStore().Put(ctx, template)
		require.NoError(t, err)

		// Create a delivery task with Spanish preference, appId in metadata and PushTokens map
		task := domain.DeliveryTask{
			TaskID:      "test-task-2",
			CustomerID:  "test-customer-1",
			EventID:     "test-event-2",
			Channel:     domain.ChannelPush,
			TemplateRef: "test-template-2",
			Subscriber: domain.Subscriber{
				ID: "sub-2",
				PushTokens: map[string]string{
					"test-app-1": "fcm-token-987654321",
				},
				Preferences: domain.SubscriberPrefs{
					Language: "es",
				},
			},
			Metadata: map[string]string{
				"appId": "test-app-1",
			},
			Payload: map[string]interface{}{
				"name": "Juan",
			},
			CreatedAt: time.Now(),
		}

		// Process the task - should succeed with mock provider
		err = pushWorker.ProcessTask(ctx, task)
		assert.NoError(t, err)
	})
}
