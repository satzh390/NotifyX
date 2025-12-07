package delivery

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/notifyx/core/domain"
	"github.com/segmentio/kafka-go"
)

// BrokerResultHandler publishes delivery tasks and logs to Kafka
type BrokerResultHandler struct {
	taskWriter *kafka.Writer
	logWriter  *kafka.Writer
}

type BrokerConfig struct {
	Brokers   []string
	TaskTopic string
	LogTopic  string
	Balancer  kafka.Balancer
}

func NewBrokerResultHandler(cfg BrokerConfig) *BrokerResultHandler {
	var taskWriter, logWriter *kafka.Writer

	if cfg.TaskTopic != "" {
		taskWriter = &kafka.Writer{
			Addr:     kafka.TCP(cfg.Brokers...),
			Topic:    cfg.TaskTopic,
			Balancer: cfg.Balancer,
		}
		if taskWriter.Balancer == nil {
			taskWriter.Balancer = &kafka.LeastBytes{}
		}
	}

	if cfg.LogTopic != "" {
		logWriter = &kafka.Writer{
			Addr:     kafka.TCP(cfg.Brokers...),
			Topic:    cfg.LogTopic,
			Balancer: cfg.Balancer,
		}
		if logWriter.Balancer == nil {
			logWriter.Balancer = &kafka.LeastBytes{}
		}
	}

	return &BrokerResultHandler{
		taskWriter: taskWriter,
		logWriter:  logWriter,
	}
}

func (h *BrokerResultHandler) HandleTask(ctx context.Context, task domain.DeliveryTask) error {
	if h.taskWriter == nil {
		return nil // Skip if not configured
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("broker: marshal task: %w", err)
	}

	return h.taskWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(task.TaskID),
		Value: payload,
	})
}

func (h *BrokerResultHandler) HandleResult(ctx context.Context, log domain.DeliveryLog) error {
	if h.logWriter == nil {
		return nil // Skip if not configured
	}

	payload, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("broker: marshal log: %w", err)
	}

	return h.logWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(log.TaskID),
		Value: payload,
	})
}

func (h *BrokerResultHandler) Close(ctx context.Context) error {
	var firstErr error
	if h.taskWriter != nil {
		if err := h.taskWriter.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if h.logWriter != nil {
		if err := h.logWriter.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

