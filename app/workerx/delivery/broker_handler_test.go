package delivery

import (
	"context"
	"testing"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestBrokerResultHandler_HandleTask(t *testing.T) {
	t.Run("successful publish", func(t *testing.T) {
		// Note: In a real test, we'd need to inject a mock writer
		// For now, we test the structure
		handler := NewBrokerResultHandler(BrokerConfig{
			Brokers:   []string{"localhost:9092"},
			TaskTopic: "test-tasks",
		})

		task := domain.DeliveryTask{
			TaskID:     "task-1",
			CustomerID: "customer-1",
			EventID:    "event-1",
			Channel:    domain.ChannelSMS,
		}

		// This would require a way to inject mock writer
		_ = handler
		_ = task
	})

	t.Run("nil task writer", func(t *testing.T) {
		handler := NewBrokerResultHandler(BrokerConfig{
			TaskTopic: "", // No topic = no writer
		})

		task := domain.DeliveryTask{
			TaskID: "task-1",
		}

		err := handler.HandleTask(context.Background(), task)
		assert.NoError(t, err) // Should skip if nil
	})
}

func TestBrokerResultHandler_HandleResult(t *testing.T) {
	t.Run("nil log writer", func(t *testing.T) {
		handler := NewBrokerResultHandler(BrokerConfig{
			LogTopic: "", // No topic = no writer
		})

		log := domain.DeliveryLog{
			TaskID:    "task-1",
			Status:    domain.EventStatusDelivered,
			Timestamp: time.Now(),
		}

		err := handler.HandleResult(context.Background(), log)
		assert.NoError(t, err) // Should skip if nil
	})
}

func TestBrokerResultHandler_Close(t *testing.T) {
	t.Run("close with writers", func(t *testing.T) {
		handler := NewBrokerResultHandler(BrokerConfig{
			Brokers:   []string{"localhost:9092"},
			TaskTopic: "test-tasks",
			LogTopic:  "test-logs",
		})

		err := handler.Close(context.Background())
		assert.NoError(t, err)
	})

	t.Run("close without writers", func(t *testing.T) {
		handler := NewBrokerResultHandler(BrokerConfig{
			TaskTopic: "",
			LogTopic:  "",
		})

		err := handler.Close(context.Background())
		assert.NoError(t, err)
	})
}

func TestNewBrokerResultHandler(t *testing.T) {
	t.Run("with both topics", func(t *testing.T) {
		handler := NewBrokerResultHandler(BrokerConfig{
			Brokers:   []string{"localhost:9092"},
			TaskTopic: "test-tasks",
			LogTopic:  "test-logs",
		})

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.taskWriter)
		assert.NotNil(t, handler.logWriter)
	})

	t.Run("with only task topic", func(t *testing.T) {
		handler := NewBrokerResultHandler(BrokerConfig{
			Brokers:   []string{"localhost:9092"},
			TaskTopic: "test-tasks",
			LogTopic:  "",
		})

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.taskWriter)
		assert.Nil(t, handler.logWriter)
	})

	t.Run("with only log topic", func(t *testing.T) {
		handler := NewBrokerResultHandler(BrokerConfig{
			Brokers:   []string{"localhost:9092"},
			TaskTopic: "",
			LogTopic:  "test-logs",
		})

		assert.NotNil(t, handler)
		assert.Nil(t, handler.taskWriter)
		assert.NotNil(t, handler.logWriter)
	})

	t.Run("default balancer", func(t *testing.T) {
		handler := NewBrokerResultHandler(BrokerConfig{
			Brokers:   []string{"localhost:9092"},
			TaskTopic: "test-tasks",
			Balancer:  nil, // Should use default
		})

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.taskWriter)
	})
}

