package recipients

import (
	"context"
	"testing"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/processor/internal/cache"
)

func TestResolver_Stream_FallsBackToSubscribedEventTypes(t *testing.T) {
	customerID := "cust-1"
	eventType := "order.created"

	subscriberStore := &fakeSubscriberStore{
		byID: map[string]domain.Subscriber{
			"sub-2": {
				ID:         "sub-2",
				CustomerID: customerID,
			},
		},
		listByEvent: map[string][]domain.Subscriber{
			eventType: {
				{
					ID:         "sub-1",
					CustomerID: customerID,
					Email:      "a@example.com",
				},
			},
		},
	}

	groupStore := &fakeGroupStore{
		listByEvent: map[string][]domain.Group{
			eventType: {
				{
					ID:          "group-1",
					CustomerID:  customerID,
					Subscribers: []string{"sub-2"},
				},
			},
		},
		byID: map[string]domain.Group{
			"group-1": {
				ID:          "group-1",
				CustomerID:  customerID,
				Subscribers: []string{"sub-2"},
			},
		},
	}

	stores := storage.Stores{
		Subscribers: subscriberStore,
		Groups:      groupStore,
	}

	resolver := NewResolver(stores, cache.NoopSubscriberCache{})

	var collected []string
	err := resolver.Stream(context.Background(), customerID, eventType, domain.Recipients{}, func(sub domain.Subscriber) error {
		collected = append(collected, sub.ID)
		return nil
	})
	if err != nil {
		t.Fatalf("resolver.Stream returned error: %v", err)
	}

	expected := map[string]struct{}{
		"sub-1": {},
		"sub-2": {},
	}

	if len(collected) != len(expected) {
		t.Fatalf("expected %d subscribers, got %d", len(expected), len(collected))
	}
	for _, id := range collected {
		if _, ok := expected[id]; !ok {
			t.Fatalf("unexpected subscriber %s", id)
		}
	}
}

type fakeSubscriberStore struct {
	byID        map[string]domain.Subscriber
	listByEvent map[string][]domain.Subscriber
}

func (store *fakeSubscriberStore) Put(ctx context.Context, sub domain.Subscriber) error {
	return nil
}

func (store *fakeSubscriberStore) Get(ctx context.Context, customerID, subscriberID string) (domain.Subscriber, error) {
	if sub, ok := store.byID[subscriberID]; ok {
		return sub, nil
	}
	return domain.Subscriber{}, storage.ErrNotFound
}

func (store *fakeSubscriberStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Subscriber], error) {
	eventType := opts.Filter["subscribedEventTypes"]
	items := store.listByEvent[eventType]
	return domain.ListResult[domain.Subscriber]{
		Items: items,
		Pagination: domain.PaginationResult{
			Page:       opts.Pagination.Page,
			PageSize:   opts.Pagination.PageSize,
			TotalPages: 1,
		},
	}, nil
}

func (store *fakeSubscriberStore) Delete(ctx context.Context, customerID, subscriberID string) error {
	return nil
}

type fakeGroupStore struct {
	byID        map[string]domain.Group
	listByEvent map[string][]domain.Group
}

func (store *fakeGroupStore) Put(ctx context.Context, group domain.Group) error {
	return nil
}

func (store *fakeGroupStore) Get(ctx context.Context, customerID, groupID string) (domain.Group, error) {
	if group, ok := store.byID[groupID]; ok {
		return group, nil
	}
	return domain.Group{}, storage.ErrNotFound
}

func (store *fakeGroupStore) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Group], error) {
	eventType := opts.Filter["subscribedEventTypes"]
	items := store.listByEvent[eventType]
	return domain.ListResult[domain.Group]{
		Items: items,
		Pagination: domain.PaginationResult{
			Page:       opts.Pagination.Page,
			PageSize:   opts.Pagination.PageSize,
			TotalPages: 1,
		},
	}, nil
}

func (store *fakeGroupStore) Delete(ctx context.Context, customerID, groupID string) error {
	return nil
}
