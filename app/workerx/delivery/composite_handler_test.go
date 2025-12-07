package delivery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResultHandler is a mock result handler for testing
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

func TestCompositeResultHandler_HandleTask(t *testing.T) {
	t.Run("all handlers succeed", func(t *testing.T) {
		handler1 := new(MockResultHandler)
		handler2 := new(MockResultHandler)
		composite := NewCompositeResultHandler(handler1, handler2)

		task := domain.DeliveryTask{
			TaskID: "task-1",
		}

		handler1.On("HandleTask", mock.Anything, task).Return(nil)
		handler2.On("HandleTask", mock.Anything, task).Return(nil)

		err := composite.HandleTask(context.Background(), task)
		assert.NoError(t, err)
		handler1.AssertExpectations(t)
		handler2.AssertExpectations(t)
	})

	t.Run("first handler fails", func(t *testing.T) {
		handler1 := new(MockResultHandler)
		handler2 := new(MockResultHandler)
		composite := NewCompositeResultHandler(handler1, handler2)

		task := domain.DeliveryTask{
			TaskID: "task-1",
		}

		handler1Err := errors.New("handler1 error")
		handler1.On("HandleTask", mock.Anything, task).Return(handler1Err)
		handler2.On("HandleTask", mock.Anything, task).Return(nil)

		err := composite.HandleTask(context.Background(), task)
		assert.Error(t, err)
		assert.Equal(t, handler1Err, err)
		handler1.AssertExpectations(t)
		handler2.AssertExpectations(t)
	})

	t.Run("second handler fails", func(t *testing.T) {
		handler1 := new(MockResultHandler)
		handler2 := new(MockResultHandler)
		composite := NewCompositeResultHandler(handler1, handler2)

		task := domain.DeliveryTask{
			TaskID: "task-1",
		}

		handler2Err := errors.New("handler2 error")
		handler1.On("HandleTask", mock.Anything, task).Return(nil)
		handler2.On("HandleTask", mock.Anything, task).Return(handler2Err)

		err := composite.HandleTask(context.Background(), task)
		assert.Error(t, err)
		assert.Equal(t, handler2Err, err)
		handler1.AssertExpectations(t)
		handler2.AssertExpectations(t)
	})
}

func TestCompositeResultHandler_HandleResult(t *testing.T) {
	t.Run("all handlers succeed", func(t *testing.T) {
		handler1 := new(MockResultHandler)
		handler2 := new(MockResultHandler)
		composite := NewCompositeResultHandler(handler1, handler2)

		log := domain.DeliveryLog{
			TaskID:    "task-1",
			Status:    domain.EventStatusDelivered,
			Timestamp: time.Now(),
		}

		handler1.On("HandleResult", mock.Anything, log).Return(nil)
		handler2.On("HandleResult", mock.Anything, log).Return(nil)

		err := composite.HandleResult(context.Background(), log)
		assert.NoError(t, err)
		handler1.AssertExpectations(t)
		handler2.AssertExpectations(t)
	})

	t.Run("first handler fails", func(t *testing.T) {
		handler1 := new(MockResultHandler)
		handler2 := new(MockResultHandler)
		composite := NewCompositeResultHandler(handler1, handler2)

		log := domain.DeliveryLog{
			TaskID: "task-1",
		}

		handler1Err := errors.New("handler1 error")
		handler1.On("HandleResult", mock.Anything, log).Return(handler1Err)
		handler2.On("HandleResult", mock.Anything, log).Return(nil)

		err := composite.HandleResult(context.Background(), log)
		assert.Error(t, err)
		assert.Equal(t, handler1Err, err)
		handler1.AssertExpectations(t)
		handler2.AssertExpectations(t)
	})
}

func TestCompositeResultHandler_Close(t *testing.T) {
	t.Run("all handlers close successfully", func(t *testing.T) {
		handler1 := new(MockResultHandler)
		handler2 := new(MockResultHandler)
		composite := NewCompositeResultHandler(handler1, handler2)

		handler1.On("Close", mock.Anything).Return(nil)
		handler2.On("Close", mock.Anything).Return(nil)

		err := composite.Close(context.Background())
		assert.NoError(t, err)
		handler1.AssertExpectations(t)
		handler2.AssertExpectations(t)
	})

	t.Run("first handler fails to close", func(t *testing.T) {
		handler1 := new(MockResultHandler)
		handler2 := new(MockResultHandler)
		composite := NewCompositeResultHandler(handler1, handler2)

		closeErr := errors.New("close error")
		handler1.On("Close", mock.Anything).Return(closeErr)
		handler2.On("Close", mock.Anything).Return(nil)

		err := composite.Close(context.Background())
		assert.Error(t, err)
		assert.Equal(t, closeErr, err)
		handler1.AssertExpectations(t)
		handler2.AssertExpectations(t)
	})
}

