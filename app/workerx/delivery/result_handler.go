package delivery

import (
	"context"

	"github.com/notifyx/core/domain"
)

// ResultHandler handles delivery results (either to MongoDB or broker)
type ResultHandler interface {
	HandleResult(ctx context.Context, log domain.DeliveryLog) error
	HandleTask(ctx context.Context, task domain.DeliveryTask) error
	Close(ctx context.Context) error
}

