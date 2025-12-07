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

func setupTestSMSWorker() (*SMSWorker, *MockSMSProvider, *MockTemplateStore) {
	provider := new(MockSMSProvider)
	templateStore := new(MockTemplateStore)
	
	baseWorker := worker.NewBaseWorker(worker.BaseWorkerOptions{
		TemplateStore: templateStore,
		Channel:       domain.ChannelSMS,
		Logger:        slog.Default(),
	})

	smsWorker := NewSMSWorker(baseWorker, provider)
	return smsWorker, provider, templateStore
}

func TestSMSWorker_processTask(t *testing.T) {
	t.Run("successful processing", func(t *testing.T) {
		worker, provider, templateStore := setupTestSMSWorker()

		task := domain.DeliveryTask{
			TaskID:      "task-1",
			CustomerID:  "customer-1",
			TemplateRef: "template-1",
			Subscriber: domain.Subscriber{
				Phone: "+1234567890",
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
			Channel:    domain.ChannelSMS,
			Content: domain.TemplateContent{
				Body: "Hello {{.name}}",
			},
		}

		templateStore.On("Get", mock.Anything, "customer-1", "template-1").Return(template, nil)
		provider.On("Send", mock.Anything, "+1234567890", "Hello John", task.Metadata).Return(nil)

		err := worker.ProcessTask(context.Background(), task)
		assert.NoError(t, err)
		templateStore.AssertExpectations(t)
		provider.AssertExpectations(t)
	})

	t.Run("template not found", func(t *testing.T) {
		worker, _, templateStore := setupTestSMSWorker()

		task := domain.DeliveryTask{
			CustomerID:  "customer-1",
			TemplateRef: "template-1",
			Subscriber: domain.Subscriber{
				Phone: "+1234567890",
			},
		}

		templateStore.On("Get", mock.Anything, "customer-1", "template-1").Return(domain.Template{}, storage.ErrNotFound)

		err := worker.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get template")
		templateStore.AssertExpectations(t)
	})

	t.Run("missing phone number", func(t *testing.T) {
		worker, _, templateStore := setupTestSMSWorker()

		task := domain.DeliveryTask{
			CustomerID:  "customer-1",
			TemplateRef: "template-1",
			Subscriber: domain.Subscriber{
				Phone: "", // Missing phone
			},
		}

		template := domain.Template{
			ID:         "template-1",
			CustomerID: "customer-1",
			Channel:    domain.ChannelSMS,
			Content: domain.TemplateContent{
				Body: "Hello",
			},
		}

		templateStore.On("Get", mock.Anything, "customer-1", "template-1").Return(template, nil)

		err := worker.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no phone number")
		templateStore.AssertExpectations(t)
	})

	t.Run("provider send error", func(t *testing.T) {
		worker, provider, templateStore := setupTestSMSWorker()

		task := domain.DeliveryTask{
			CustomerID:  "customer-1",
			TemplateRef: "template-1",
			Subscriber: domain.Subscriber{
				Phone: "+1234567890",
			},
			Payload: map[string]interface{}{},
		}

		template := domain.Template{
			ID:         "template-1",
			CustomerID: "customer-1",
			Channel:    domain.ChannelSMS,
			Content: domain.TemplateContent{
				Body: "Hello",
			},
		}

		sendErr := errors.New("provider error")
		templateStore.On("Get", mock.Anything, "customer-1", "template-1").Return(template, nil)
		provider.On("Send", mock.Anything, "+1234567890", "Hello", task.Metadata).Return(sendErr)

		err := worker.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "send")
		templateStore.AssertExpectations(t)
		provider.AssertExpectations(t)
	})

	t.Run("template with translation", func(t *testing.T) {
		worker, provider, templateStore := setupTestSMSWorker()

		task := domain.DeliveryTask{
			CustomerID:  "customer-1",
			TemplateRef: "template-1",
			Subscriber: domain.Subscriber{
				Phone: "+1234567890",
				Preferences: domain.SubscriberPrefs{
					Language: "es",
				},
			},
			Payload: map[string]interface{}{},
		}

		template := domain.Template{
			ID:         "template-1",
			CustomerID: "customer-1",
			Channel:    domain.ChannelSMS,
			Content: domain.TemplateContent{
				Body: "Hello",
			},
			Translations: map[string]domain.TemplateContent{
				"es": {
					Body: "Hola",
				},
			},
		}

		templateStore.On("Get", mock.Anything, "customer-1", "template-1").Return(template, nil)
		provider.On("Send", mock.Anything, "+1234567890", "Hola", task.Metadata).Return(nil)

		err := worker.ProcessTask(context.Background(), task)
		assert.NoError(t, err)
		templateStore.AssertExpectations(t)
		provider.AssertExpectations(t)
	})
}

func TestSMSWorker_NewSMSWorker(t *testing.T) {
	provider := new(MockSMSProvider)
	templateStore := new(MockTemplateStore)
	
	baseWorker := worker.NewBaseWorker(worker.BaseWorkerOptions{
		TemplateStore: templateStore,
		Channel:       domain.ChannelSMS,
	})

	smsWorker := NewSMSWorker(baseWorker, provider)
	assert.NotNil(t, smsWorker)
	assert.Equal(t, provider, smsWorker.provider)
	assert.Equal(t, baseWorker, smsWorker.BaseWorker)
}

