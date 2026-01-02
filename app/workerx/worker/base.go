package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/workerx/delivery"
	"github.com/segmentio/kafka-go"
)

// BaseWorker provides common functionality for all channel workers
type BaseWorker struct {
	reader        *kafka.Reader
	dlq           *kafka.Writer
	templateStore storage.TemplateStore
	resultHandler delivery.ResultHandler
	logger        *slog.Logger
	channel       domain.ChannelType
	processTask   func(ctx context.Context, task domain.DeliveryTask) error
}

type BaseWorkerOptions struct {
	Reader        *kafka.Reader
	DLQ           *kafka.Writer
	TemplateStore storage.TemplateStore
	ResultHandler delivery.ResultHandler
	Logger        *slog.Logger
	Channel       domain.ChannelType
}

func NewBaseWorker(opts BaseWorkerOptions) *BaseWorker {
	return &BaseWorker{
		reader:        opts.Reader,
		dlq:           opts.DLQ,
		templateStore: opts.TemplateStore,
		resultHandler: opts.ResultHandler,
		logger:        opts.Logger,
		channel:       opts.Channel,
	}
}

func (w *BaseWorker) Run(ctx context.Context) error {
	for {
		msg, err := w.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("worker: fetch message: %w", err)
		}

		if err := w.handleMessage(ctx, msg); err != nil {
			w.logger.Error("worker: handle message failed",
				slog.String("channel", string(w.channel)),
				slog.String("error", err.Error()),
			)
			if dlqErr := w.publishDLQ(ctx, msg, err); dlqErr != nil {
				w.logger.Error("worker: publish to DLQ failed", slog.String("error", dlqErr.Error()))
			}
		} else {
			if err := w.reader.CommitMessages(ctx, msg); err != nil {
				w.logger.Error("worker: commit failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (w *BaseWorker) handleMessage(ctx context.Context, msg kafka.Message) error {
	var task domain.DeliveryTask
	if err := json.Unmarshal(msg.Value, &task); err != nil {
		return fmt.Errorf("worker: unmarshal task: %w", err)
	}

	// Validate channel matches
	if task.Channel != w.channel {
		return fmt.Errorf("worker: channel mismatch: expected %s, got %s", w.channel, task.Channel)
	}

	// Store task if configured
	if w.resultHandler != nil {
		if err := w.resultHandler.HandleTask(ctx, task); err != nil {
			w.logger.Warn("worker: failed to store task", slog.String("error", err.Error()))
		}
	}

	// Process the task (implemented by specific workers)
	if w.processTask == nil {
		return fmt.Errorf("worker: processTask not implemented")
	}
	if err := w.processTask(ctx, task); err != nil {
		// Log failure
		log := domain.DeliveryLog{
			TaskID:     task.TaskID,
			CustomerID: task.CustomerID,
			EventID:    task.EventID,
			Channel:    task.Channel,
			Status:     domain.EventStatusFailed,
			Error:      err.Error(),
			Timestamp:  time.Now(),
			Metadata:   task.Metadata,
		}
		if w.resultHandler != nil {
			_ = w.resultHandler.HandleResult(ctx, log)
		}
		return err
	}

	// Log success
	log := domain.DeliveryLog{
		TaskID:     task.TaskID,
		CustomerID: task.CustomerID,
		EventID:    task.EventID,
		Channel:    task.Channel,
		Status:     domain.EventStatusDelivered,
		Timestamp:  time.Now(),
		Metadata:   task.Metadata,
	}
	if w.resultHandler != nil {
		_ = w.resultHandler.HandleResult(ctx, log)
	}

	return nil
}

func (w *BaseWorker) publishDLQ(ctx context.Context, msg kafka.Message, procErr error) error {
	if w.dlq == nil {
		return nil
	}

	payload := map[string]any{
		"value":   string(msg.Value),
		"error":   procErr.Error(),
		"topic":   msg.Topic,
		"offset":  msg.Offset,
		"headers": msg.Headers,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return w.dlq.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: body,
	})
}

func (w *BaseWorker) Close() error {
	return w.reader.Close()
}

// SetProcessTask sets the task processing function
func (w *BaseWorker) SetProcessTask(fn func(ctx context.Context, task domain.DeliveryTask) error) {
	w.processTask = fn
}

// GetTemplateStore returns the template store
func (w *BaseWorker) GetTemplateStore() storage.TemplateStore {
	return w.templateStore
}

// GetLogger returns the logger
func (w *BaseWorker) GetLogger() *slog.Logger {
	return w.logger
}

