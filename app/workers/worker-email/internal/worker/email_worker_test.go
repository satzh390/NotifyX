package worker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/workerx/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestEmailWorker() (*EmailWorker, *MockEmailProvider, *MockTemplateStore) {
	provider := new(MockEmailProvider)
	templateStore := new(MockTemplateStore)
	
	baseWorker := worker.NewBaseWorker(worker.BaseWorkerOptions{
		TemplateStore: templateStore,
		Channel:       domain.ChannelEmail,
		Logger:        slog.Default(),
	})

	emailWorker := NewEmailWorker(baseWorker, provider)
	return emailWorker, provider, templateStore
}

func TestEmailWorker_processTask(t *testing.T) {
	t.Run("successful processing", func(t *testing.T) {
		worker, provider, templateStore := setupTestEmailWorker()

		task := domain.DeliveryTask{
			TaskID:      "task-1",
			CustomerID:  "customer-1",
			TemplateRef: "template-1",
			Subscriber: domain.Subscriber{
				Email: "test@example.com",
				Preferences: domain.SubscriberPrefs{
					Language: "en",
				},
			},
			Payload: map[string]interface{}{
				"name": "John",
			},
		}

		template := domain.Template{
			ID:         "template-1",
			CustomerID: "customer-1",
			Channel:    domain.ChannelEmail,
			Content: domain.TemplateContent{
				Subject: "Hello {{.name}}",
				Body:    "<p>Hello {{.name}}</p>",
			},
		}

		templateStore.On("Get", mock.Anything, "customer-1", "template-1").Return(template, nil)
		provider.On("Send", mock.Anything, "test@example.com", "Hello John", "<p>Hello John</p>", task.Metadata).Return(nil)

		err := worker.ProcessTask(context.Background(), task)
		assert.NoError(t, err)
		templateStore.AssertExpectations(t)
		provider.AssertExpectations(t)
	})

	t.Run("template not found", func(t *testing.T) {
		worker, _, templateStore := setupTestEmailWorker()

		task := domain.DeliveryTask{
			CustomerID:  "customer-1",
			TemplateRef: "template-1",
			Subscriber: domain.Subscriber{
				Email: "test@example.com",
			},
		}

		templateStore.On("Get", mock.Anything, "customer-1", "template-1").Return(domain.Template{}, storage.ErrNotFound)

		err := worker.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get template")
		templateStore.AssertExpectations(t)
	})

	t.Run("missing email address", func(t *testing.T) {
		worker, _, templateStore := setupTestEmailWorker()

		task := domain.DeliveryTask{
			CustomerID:  "customer-1",
			TemplateRef: "template-1",
			Subscriber: domain.Subscriber{
				Email: "", // Missing email
			},
		}

		template := domain.Template{
			ID:         "template-1",
			CustomerID: "customer-1",
			Channel:    domain.ChannelEmail,
			Content: domain.TemplateContent{
				Subject: "Hello",
				Body:    "Hello",
			},
		}

		templateStore.On("Get", mock.Anything, "customer-1", "template-1").Return(template, nil)

		err := worker.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no email address")
		templateStore.AssertExpectations(t)
	})

	t.Run("provider send error", func(t *testing.T) {
		worker, provider, templateStore := setupTestEmailWorker()

		task := domain.DeliveryTask{
			CustomerID:  "customer-1",
			TemplateRef: "template-1",
			Subscriber: domain.Subscriber{
				Email: "test@example.com",
			},
			Payload: map[string]interface{}{},
		}

		template := domain.Template{
			ID:         "template-1",
			CustomerID: "customer-1",
			Channel:    domain.ChannelEmail,
			Content: domain.TemplateContent{
				Subject: "Hello",
				Body:    "Hello",
			},
		}

		sendErr := errors.New("provider error")
		templateStore.On("Get", mock.Anything, "customer-1", "template-1").Return(template, nil)
		provider.On("Send", mock.Anything, "test@example.com", "Hello", "Hello", task.Metadata).Return(sendErr)

		err := worker.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "send")
		templateStore.AssertExpectations(t)
		provider.AssertExpectations(t)
	})
}

func TestEmailWorker_NewEmailWorker(t *testing.T) {
	provider := new(MockEmailProvider)
	templateStore := new(MockTemplateStore)
	
	baseWorker := worker.NewBaseWorker(worker.BaseWorkerOptions{
		TemplateStore: templateStore,
		Channel:       domain.ChannelEmail,
	})

	emailWorker := NewEmailWorker(baseWorker, provider)
	assert.NotNil(t, emailWorker)
	assert.Equal(t, provider, emailWorker.provider)
	assert.Equal(t, baseWorker, emailWorker.BaseWorker)
}

