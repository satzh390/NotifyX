package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/notifyx/core/storage"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Options struct {
	URI              string
	Database         string
	CollectionPrefix string
}

// NewStoreSet connects to MongoDB and returns storage implementations plus a cleanup function.
func NewStoreSet(ctx context.Context, opts Options) (storage.Stores, func(context.Context) error, error) {
	if opts.URI == "" {
		return storage.Stores{}, nil, errors.New("mongo: URI is required")
	}
	if opts.Database == "" {
		return storage.Stores{}, nil, errors.New("mongo: database is required")
	}

	client, err := mongoDriver.Connect(ctx, options.Client().ApplyURI(opts.URI))
	if err != nil {
		return storage.Stores{}, nil, fmt.Errorf("mongo: connect: %w", err)
	}

	db := client.Database(opts.Database)
	prefix := opts.CollectionPrefix
	if prefix != "" && prefix[len(prefix)-1] != '_' {
		prefix += "_"
	}

	stores := storage.Stores{
		Subscribers:   &SubscriberRepository{collection: db.Collection(prefix + "subscribers")},
		Groups:        &GroupRepository{collection: db.Collection(prefix + "groups")},
		Rules:         &RuleRepository{collection: db.Collection(prefix + "rules")},
		Templates:     &TemplateRepository{collection: db.Collection(prefix + "templates")},
		Organizations: &OrganizationRepository{collection: db.Collection(prefix + "organizations")},
		Customers:     &CustomerRepository{collection: db.Collection(prefix + "customers")},
		DeliveryTasks: &DeliveryTaskRepository{collection: db.Collection(prefix + "deliveryTasks")},
		DeliveryLogs:   &DeliveryLogRepository{collection: db.Collection(prefix + "deliveryLogs")},
	}

	cleanup := func(ctx context.Context) error {
		return client.Disconnect(ctx)
	}

	return stores, cleanup, nil
}
