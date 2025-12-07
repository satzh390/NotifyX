package delivery

import (
	"context"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
)

// MongoResultHandler stores delivery tasks and logs in MongoDB
type MongoResultHandler struct {
	taskStore storage.DeliveryTaskStore
	logStore  storage.DeliveryLogStore
}

func NewMongoResultHandler(taskStore storage.DeliveryTaskStore, logStore storage.DeliveryLogStore) *MongoResultHandler {
	return &MongoResultHandler{
		taskStore: taskStore,
		logStore:  logStore,
	}
}

func (h *MongoResultHandler) HandleTask(ctx context.Context, task domain.DeliveryTask) error {
	if h.taskStore == nil {
		return nil // Skip if not configured
	}
	return h.taskStore.Put(ctx, task)
}

func (h *MongoResultHandler) HandleResult(ctx context.Context, log domain.DeliveryLog) error {
	if h.logStore == nil {
		return nil // Skip if not configured
	}
	return h.logStore.Put(ctx, log)
}

func (h *MongoResultHandler) Close(ctx context.Context) error {
	return nil // No cleanup needed
}

