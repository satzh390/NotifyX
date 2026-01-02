package storage

import (
	"context"
	"errors"

	"github.com/notifyx/core/domain"
)

var ErrNotFound = errors.New("storage: record not found")

type SubscriberStore interface {
	Put(ctx context.Context, sub domain.Subscriber) error
	Get(ctx context.Context, customerID, subscriberID string) (domain.Subscriber, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Subscriber], error)
	Delete(ctx context.Context, customerID, subscriberID string) error
}

type GroupStore interface {
	Put(ctx context.Context, group domain.Group) error
	Get(ctx context.Context, customerID, groupID string) (domain.Group, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Group], error)
	Delete(ctx context.Context, customerID, groupID string) error
}

type RuleStore interface {
	Put(ctx context.Context, rule domain.Rule) error
	Get(ctx context.Context, customerID, eventType string) (domain.Rule, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Rule], error)
	Delete(ctx context.Context, customerID, eventType string) error
}

type TemplateStore interface {
	Put(ctx context.Context, tpl domain.Template) error
	Get(ctx context.Context, customerID, templateID string) (domain.Template, error)
	GetByLanguage(ctx context.Context, customerID, templateID, language string) (domain.Template, error)
	Delete(ctx context.Context, customerID, templateID string) error
}

type OrganizationStore interface {
	Put(ctx context.Context, org domain.Organization) error
	Get(ctx context.Context, orgID string) (domain.Organization, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Organization], error)
	Delete(ctx context.Context, orgID string) error
}

type CustomerStore interface {
	Put(ctx context.Context, customer domain.Customer) error
	Get(ctx context.Context, customerID string) (domain.Customer, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Customer], error)
	Delete(ctx context.Context, customerID string) error
}

type DeliveryTaskStore interface {
	Put(ctx context.Context, task domain.DeliveryTask) error
	Get(ctx context.Context, taskID string) (domain.DeliveryTask, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.DeliveryTask], error)
}

type DeliveryLogStore interface {
	Put(ctx context.Context, log domain.DeliveryLog) error
	Get(ctx context.Context, taskID string) (domain.DeliveryLog, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.DeliveryLog], error)
}

type AppConfigStore interface {
	Put(ctx context.Context, appConfig domain.AppConfig) error
	Get(ctx context.Context, orgID, appID string) (domain.AppConfig, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.AppConfig], error)
	Delete(ctx context.Context, orgID, appID string) error
}

type Stores struct {
	Subscribers    SubscriberStore
	Groups         GroupStore
	Rules          RuleStore
	Templates      TemplateStore
	Organizations  OrganizationStore
	Customers      CustomerStore
	DeliveryTasks  DeliveryTaskStore
	DeliveryLogs   DeliveryLogStore
	AppConfigs     AppConfigStore
}
