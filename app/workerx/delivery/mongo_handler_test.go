package delivery

import (
	"context"
	"testing"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeliveryTaskStore is a mock delivery task store
type MockDeliveryTaskStore struct {
	mock.Mock
}

func (m *MockDeliveryTaskStore) Put(ctx context.Context, task domain.DeliveryTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockDeliveryTaskStore) Get(ctx context.Context, taskID string) (domain.DeliveryTask, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).(domain.DeliveryTask), args.Error(1)
}

func (m *MockDeliveryTaskStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.DeliveryTask], error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.DeliveryTask]), args.Error(1)
}

// MockDeliveryLogStore is a mock delivery log store
type MockDeliveryLogStore struct {
	mock.Mock
}

func (m *MockDeliveryLogStore) Put(ctx context.Context, log domain.DeliveryLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockDeliveryLogStore) Get(ctx context.Context, taskID string) (domain.DeliveryLog, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).(domain.DeliveryLog), args.Error(1)
}

func (m *MockDeliveryLogStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.DeliveryLog], error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(domain.ListResult[domain.DeliveryLog]), args.Error(1)
}

func TestMongoResultHandler_HandleTask(t *testing.T) {
	t.Run("successful store", func(t *testing.T) {
		taskStore := new(MockDeliveryTaskStore)
		logStore := new(MockDeliveryLogStore)
		handler := NewMongoResultHandler(taskStore, logStore)

		task := domain.DeliveryTask{
			TaskID:     "task-1",
			CustomerID: "customer-1",
			EventID:    "event-1",
			Channel:    domain.ChannelSMS,
		}

		taskStore.On("Put", mock.Anything, task).Return(nil)

		err := handler.HandleTask(context.Background(), task)
		assert.NoError(t, err)
		taskStore.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		taskStore := new(MockDeliveryTaskStore)
		logStore := new(MockDeliveryLogStore)
		handler := NewMongoResultHandler(taskStore, logStore)

		task := domain.DeliveryTask{
			TaskID: "task-1",
		}

		storeErr := storage.ErrNotFound
		taskStore.On("Put", mock.Anything, task).Return(storeErr)

		err := handler.HandleTask(context.Background(), task)
		assert.Error(t, err)
		assert.Equal(t, storeErr, err)
		taskStore.AssertExpectations(t)
	})

	t.Run("nil task store", func(t *testing.T) {
		handler := NewMongoResultHandler(nil, nil)

		task := domain.DeliveryTask{
			TaskID: "task-1",
		}

		err := handler.HandleTask(context.Background(), task)
		assert.NoError(t, err) // Should skip if nil
	})
}

func TestMongoResultHandler_HandleResult(t *testing.T) {
	t.Run("successful store", func(t *testing.T) {
		taskStore := new(MockDeliveryTaskStore)
		logStore := new(MockDeliveryLogStore)
		handler := NewMongoResultHandler(taskStore, logStore)

		log := domain.DeliveryLog{
			TaskID:     "task-1",
			CustomerID: "customer-1",
			EventID:    "event-1",
			Channel:    domain.ChannelSMS,
			Status:     domain.EventStatusDelivered,
			Timestamp:  time.Now(),
		}

		logStore.On("Put", mock.Anything, log).Return(nil)

		err := handler.HandleResult(context.Background(), log)
		assert.NoError(t, err)
		logStore.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		taskStore := new(MockDeliveryTaskStore)
		logStore := new(MockDeliveryLogStore)
		handler := NewMongoResultHandler(taskStore, logStore)

		log := domain.DeliveryLog{
			TaskID: "task-1",
		}

		storeErr := storage.ErrNotFound
		logStore.On("Put", mock.Anything, log).Return(storeErr)

		err := handler.HandleResult(context.Background(), log)
		assert.Error(t, err)
		assert.Equal(t, storeErr, err)
		logStore.AssertExpectations(t)
	})

	t.Run("nil log store", func(t *testing.T) {
		handler := NewMongoResultHandler(nil, nil)

		log := domain.DeliveryLog{
			TaskID: "task-1",
		}

		err := handler.HandleResult(context.Background(), log)
		assert.NoError(t, err) // Should skip if nil
	})
}

func TestMongoResultHandler_Close(t *testing.T) {
	handler := NewMongoResultHandler(nil, nil)
	err := handler.Close(context.Background())
	assert.NoError(t, err)
}

