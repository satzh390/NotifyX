package delivery

import (
	"context"

	"github.com/notifyx/core/domain"
)

// CompositeResultHandler combines multiple result handlers
type CompositeResultHandler struct {
	handlers []ResultHandler
}

func NewCompositeResultHandler(handlers ...ResultHandler) *CompositeResultHandler {
	return &CompositeResultHandler{
		handlers: handlers,
	}
}

func (h *CompositeResultHandler) HandleTask(ctx context.Context, task domain.DeliveryTask) error {
	var firstErr error
	for _, handler := range h.handlers {
		if err := handler.HandleTask(ctx, task); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (h *CompositeResultHandler) HandleResult(ctx context.Context, log domain.DeliveryLog) error {
	var firstErr error
	for _, handler := range h.handlers {
		if err := handler.HandleResult(ctx, log); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (h *CompositeResultHandler) Close(ctx context.Context) error {
	var firstErr error
	for _, handler := range h.handlers {
		if err := handler.Close(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

