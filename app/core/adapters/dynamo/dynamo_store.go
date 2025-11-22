package dynamo

import (
	"context"
	"errors"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
)

var errNotImplemented = errors.New("dynamo adapter: not implemented")

type SubscriberRepository struct{}

func (repo *SubscriberRepository) Put(ctx context.Context, sub domain.Subscriber) error {
	return errNotImplemented
}

func (repo *SubscriberRepository) Get(ctx context.Context, orgID, subscriberID string) (domain.Subscriber, error) {
	return domain.Subscriber{}, errNotImplemented
}

func (repo *SubscriberRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Subscriber], error) {
	return domain.ListResult[domain.Subscriber]{}, errNotImplemented
}

func (repo *SubscriberRepository) Delete(ctx context.Context, orgID, subscriberID string) error {
	return errNotImplemented
}

type GroupRepository struct{}

func (repo *GroupRepository) Put(ctx context.Context, group domain.Group) error {
	return errNotImplemented
}

func (repo *GroupRepository) Get(ctx context.Context, orgID, groupID string) (domain.Group, error) {
	return domain.Group{}, errNotImplemented
}

func (repo *GroupRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Group], error) {
	return domain.ListResult[domain.Group]{}, errNotImplemented
}

func (repo *GroupRepository) Delete(ctx context.Context, orgID, groupID string) error {
	return errNotImplemented
}

type RuleRepository struct{}

func (repo *RuleRepository) Put(ctx context.Context, rule domain.Rule) error {
	return errNotImplemented
}

func (repo *RuleRepository) Get(ctx context.Context, orgID, eventType string) (domain.Rule, error) {
	return domain.Rule{}, errNotImplemented
}

func (repo *RuleRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Rule], error) {
	return domain.ListResult[domain.Rule]{}, errNotImplemented
}

func (repo *RuleRepository) Delete(ctx context.Context, orgID, eventType string) error {
	return errNotImplemented
}

// NewStoreSet returns storage interfaces backed by DynamoDB-specific repositories.
func NewStoreSet() storage.Stores {
	return storage.Stores{
		Subscribers: &SubscriberRepository{},
		Groups:      &GroupRepository{},
		Rules:       &RuleRepository{},
		// Templates: not implemented for DynamoDB
	}
}

