package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/notifyx/core/domain"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKafkaReader is a mock Kafka reader
type MockKafkaReader struct {
	mock.Mock
}

func (m *MockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockKafkaWriter is a mock Kafka writer
type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockTemplateStore is a mock template store
type MockTemplateStore struct {
	mock.Mock
}

func (m *MockTemplateStore) Get(ctx context.Context, customerID, templateID string) (domain.Template, error) {
	args := m.Called(ctx, customerID, templateID)
	return args.Get(0).(domain.Template), args.Error(1)
}

func (m *MockTemplateStore) GetByLanguage(ctx context.Context, customerID, templateID, language string) (domain.Template, error) {
	args := m.Called(ctx, customerID, templateID, language)
	return args.Get(0).(domain.Template), args.Error(1)
}

func (m *MockTemplateStore) Put(ctx context.Context, tpl domain.Template) error {
	args := m.Called(ctx, tpl)
	return args.Error(0)
}

func (m *MockTemplateStore) Delete(ctx context.Context, customerID, templateID string) error {
	args := m.Called(ctx, customerID, templateID)
	return args.Error(0)
}

// MockResultHandler is a mock result handler
type MockResultHandler struct {
	mock.Mock
}

func (m *MockResultHandler) HandleTask(ctx context.Context, task domain.DeliveryTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockResultHandler) HandleResult(ctx context.Context, log domain.DeliveryLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockResultHandler) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupTestBaseWorker() (*BaseWorker, *MockKafkaReader, *MockKafkaWriter, *MockTemplateStore, *MockResultHandler) {
	reader := new(MockKafkaReader)
	dlq := new(MockKafkaWriter)
	templateStore := new(MockTemplateStore)
	resultHandler := new(MockResultHandler)
	logger := slog.Default()

	worker := NewBaseWorker(BaseWorkerOptions{
		Reader:        &kafka.Reader{}, // Will be replaced in tests
		DLQ:           &kafka.Writer{}, // Will be replaced in tests
		TemplateStore: templateStore,
		ResultHandler: resultHandler,
		Logger:        logger,
		Channel:       domain.ChannelSMS,
	})

	// Use reflection or direct field access for testing
	// For now, we'll test through the public interface

	return worker, reader, dlq, templateStore, resultHandler
}

func TestBaseWorker_handleMessage(t *testing.T) {
	t.Run("successful processing", func(t *testing.T) {
		worker, _, _, _, resultHandler := setupTestBaseWorker()
		
		task := domain.DeliveryTask{
			TaskID:     "task-1",
			CustomerID: "customer-1",
			EventID:    "event-1",
			Channel:    domain.ChannelSMS,
			Metadata:   map[string]string{"key": "value"},
		}

		taskJSON, _ := json.Marshal(task)
		msg := kafka.Message{
			Key:   []byte("task-1"),
			Value: taskJSON,
		}

		// Set up process task function
		processCalled := false
		worker.SetProcessTask(func(ctx context.Context, taskParam domain.DeliveryTask) error {
			processCalled = true
			assert.Equal(t, task.TaskID, taskParam.TaskID)
			return nil
		})

		resultHandler.On("HandleTask", mock.Anything, task).Return(nil)
		resultHandler.On("HandleResult", mock.Anything, mock.MatchedBy(func(log domain.DeliveryLog) bool {
			return log.TaskID == task.TaskID && log.Status == domain.EventStatusDelivered
		})).Return(nil)

		err := worker.handleMessage(context.Background(), msg)
		assert.NoError(t, err)
		assert.True(t, processCalled)
		resultHandler.AssertExpectations(t)
	})

	t.Run("channel mismatch", func(t *testing.T) {
		worker, _, _, _, _ := setupTestBaseWorker()
		
		task := domain.DeliveryTask{
			TaskID:     "task-1",
			CustomerID: "customer-1",
			EventID:    "event-1",
			Channel:    domain.ChannelEmail, // Different channel
		}

		taskJSON, _ := json.Marshal(task)
		msg := kafka.Message{
			Key:   []byte("task-1"),
			Value: taskJSON,
		}

		err := worker.handleMessage(context.Background(), msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "channel mismatch")
	})

	t.Run("processing error", func(t *testing.T) {
		worker, _, _, _, resultHandler := setupTestBaseWorker()
		
		task := domain.DeliveryTask{
			TaskID:     "task-1",
			CustomerID: "customer-1",
			EventID:    "event-1",
			Channel:    domain.ChannelSMS,
		}

		taskJSON, _ := json.Marshal(task)
		msg := kafka.Message{
			Key:   []byte("task-1"),
			Value: taskJSON,
		}

		processErr := errors.New("processing failed")
		worker.SetProcessTask(func(ctx context.Context, t domain.DeliveryTask) error {
			return processErr
		})

		resultHandler.On("HandleTask", mock.Anything, task).Return(nil)
		resultHandler.On("HandleResult", mock.Anything, mock.MatchedBy(func(log domain.DeliveryLog) bool {
			return log.TaskID == task.TaskID && log.Status == domain.EventStatusFailed && log.Error == processErr.Error()
		})).Return(nil)

		err := worker.handleMessage(context.Background(), msg)
		assert.Error(t, err)
		assert.Equal(t, processErr, err)
		resultHandler.AssertExpectations(t)
	})

	t.Run("invalid message", func(t *testing.T) {
		worker, _, _, _, _ := setupTestBaseWorker()
		
		msg := kafka.Message{
			Key:   []byte("task-1"),
			Value: []byte("invalid json"),
		}

		err := worker.handleMessage(context.Background(), msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
}

func TestBaseWorker_publishDLQ(t *testing.T) {
	t.Run("without DLQ writer", func(t *testing.T) {
		worker := NewBaseWorker(BaseWorkerOptions{
			DLQ: nil, // No DLQ
		})

		msg := kafka.Message{
			Key:   []byte("key"),
			Value: []byte("value"),
		}
		procErr := errors.New("processing error")

		err := worker.publishDLQ(context.Background(), msg, procErr)
		assert.NoError(t, err) // Should return nil when DLQ is nil
	})
}

func TestBaseWorker_SetProcessTask(t *testing.T) {
	worker := NewBaseWorker(BaseWorkerOptions{
		Channel: domain.ChannelSMS,
	})

	called := false
	taskFn := func(ctx context.Context, task domain.DeliveryTask) error {
		called = true
		return nil
	}

	worker.SetProcessTask(taskFn)
	assert.NotNil(t, worker.processTask)

	// Test that it's set correctly
	_ = worker.processTask(context.Background(), domain.DeliveryTask{})
	assert.True(t, called)
}

func TestBaseWorker_GetTemplateStore(t *testing.T) {
	templateStore := new(MockTemplateStore)
	worker := NewBaseWorker(BaseWorkerOptions{
		TemplateStore: templateStore,
	})

	assert.Equal(t, templateStore, worker.GetTemplateStore())
}

func TestBaseWorker_GetLogger(t *testing.T) {
	logger := slog.Default()
	worker := NewBaseWorker(BaseWorkerOptions{
		Logger: logger,
	})

	assert.Equal(t, logger, worker.GetLogger())
}

