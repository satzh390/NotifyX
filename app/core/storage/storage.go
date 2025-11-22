package storage

import (
	"context"
	"errors"

	"github.com/notifyx/core/domain"
)

var ErrNotFound = errors.New("storage: record not found")

type SubscriberStore interface {
	Put(ctx context.Context, sub domain.Subscriber) error
	Get(ctx context.Context, orgID, subscriberID string) (domain.Subscriber, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Subscriber], error)
	Delete(ctx context.Context, orgID, subscriberID string) error
}

type GroupStore interface {
	Put(ctx context.Context, group domain.Group) error
	Get(ctx context.Context, orgID, groupID string) (domain.Group, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Group], error)
	Delete(ctx context.Context, orgID, groupID string) error
}

type RuleStore interface {
	Put(ctx context.Context, rule domain.Rule) error
	Get(ctx context.Context, orgID, eventType string) (domain.Rule, error)
	List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Rule], error)
	Delete(ctx context.Context, orgID, eventType string) error
}

type TemplateStore interface {
	Put(ctx context.Context, tpl domain.Template) error
	Get(ctx context.Context, orgID, templateID string) (domain.Template, error)
	GetByLanguage(ctx context.Context, orgID, templateID, language string) (domain.Template, error)
	Delete(ctx context.Context, orgID, templateID string) error
}

type Stores struct {
	Subscribers SubscriberStore
	Groups      GroupStore
	Rules       RuleStore
	Templates   TemplateStore
}
