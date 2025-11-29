package recipients

import (
	"context"
	"fmt"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/processor/internal/cache"
)

const (
	defaultPageSize = 500
)

type Resolver struct {
	stores   storage.Stores
	cache    cache.SubscriberCache
	PageSize int // Public field - users can set it directly
}

func NewResolver(stores storage.Stores, cacheProvider cache.SubscriberCache) *Resolver {
	if cacheProvider == nil {
		cacheProvider = cache.NoopSubscriberCache{}
	}
	return &Resolver{
		stores:   stores,
		cache:    cacheProvider,
		PageSize: defaultPageSize,
	}
}

func (resolver *Resolver) Stream(ctx context.Context, customerID, eventType string, recipients domain.Recipients, visitor func(domain.Subscriber) error) error {
	explicitRecipients := hasExplicitRecipients(recipients)

	if err := resolver.streamSubscriberIDs(ctx, customerID, recipients.SubscriberIDs, visitor); err != nil {
		return err
	}
	if err := resolver.streamGroups(ctx, customerID, recipients.Groups, visitor); err != nil {
		return err
	}
	if recipients.Broadcast {
		if err := resolver.streamBroadcast(ctx, customerID, visitor); err != nil {
			return err
		}
	}
	if err := resolver.streamDirectEmails(ctx, customerID, recipients.DirectEmails, visitor); err != nil {
		return err
	}
	if err := resolver.streamDirectPhones(ctx, customerID, recipients.DirectPhones, visitor); err != nil {
		return err
	}

	if !explicitRecipients && eventType != "" {
		if err := resolver.streamInterestedSubscribers(ctx, customerID, eventType, visitor); err != nil {
			return err
		}
		if err := resolver.streamInterestedGroups(ctx, customerID, eventType, visitor); err != nil {
			return err
		}
	}
	return nil
}

func hasExplicitRecipients(recipients domain.Recipients) bool {
	return len(recipients.SubscriberIDs) > 0 ||
		len(recipients.Groups) > 0 ||
		recipients.Broadcast ||
		len(recipients.DirectEmails) > 0 ||
		len(recipients.DirectPhones) > 0
}

func (resolver *Resolver) streamSubscriberIDs(ctx context.Context, customerID string, ids []string, visitor func(domain.Subscriber) error) error {
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		if id == "" {
			continue
		}

		subscriber, ok, err := resolver.getSubscriber(ctx, customerID, id)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if err := visitor(subscriber); err != nil {
			return err
		}
	}
	return nil
}

func (resolver *Resolver) streamGroups(ctx context.Context, customerID string, groupIDs []string, visitor func(domain.Subscriber) error) error {
	for _, groupID := range groupIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if groupID == "" {
			continue
		}
		group, err := resolver.stores.Groups.Get(ctx, customerID, groupID)
		if err != nil {
			if err == storage.ErrNotFound {
				continue
			}
			return fmt.Errorf("recipients: group %s: %w", groupID, err)
		}
		if err := resolver.streamSubscriberIDs(ctx, customerID, group.Subscribers, visitor); err != nil {
			return err
		}
	}
	return nil
}

func (resolver *Resolver) streamBroadcast(ctx context.Context, customerID string, visitor func(domain.Subscriber) error) error {
	return resolver.streamPaginatedSubscribers(ctx, map[string]string{
		"customerId": customerID,
	}, visitor)
}

func (resolver *Resolver) getSubscriber(ctx context.Context, customerID, subscriberID string) (domain.Subscriber, bool, error) {
	if cached, ok := resolver.cache.Get(ctx, customerID, subscriberID); ok {
		return cached, true, nil
	}

	subscriber, err := resolver.stores.Subscribers.Get(ctx, customerID, subscriberID)
	if err != nil {
		if err == storage.ErrNotFound {
			return domain.Subscriber{}, false, nil
		}
		return domain.Subscriber{}, false, fmt.Errorf("recipients: subscriber %s: %w", subscriberID, err)
	}

	_ = resolver.cache.Set(ctx, subscriber)
	return subscriber, true, nil
}

func (resolver *Resolver) streamDirectEmails(ctx context.Context, customerID string, emails []string, visitor func(domain.Subscriber) error) error {
	for _, email := range emails {
		if err := ctx.Err(); err != nil {
			return err
		}
		if email == "" {
			continue
		}
		if err := visitor(createDirectEmailSubscriber(customerID, email)); err != nil {
			return err
		}
	}
	return nil
}

func (resolver *Resolver) streamDirectPhones(ctx context.Context, customerID string, phones []string, visitor func(domain.Subscriber) error) error {
	for _, phone := range phones {
		if err := ctx.Err(); err != nil {
			return err
		}
		if phone == "" {
			continue
		}
		if err := visitor(createDirectPhoneSubscriber(customerID, phone)); err != nil {
			return err
		}
	}
	return nil
}

func (resolver *Resolver) streamInterestedSubscribers(ctx context.Context, customerID, eventType string, visitor func(domain.Subscriber) error) error {
	return resolver.streamPaginatedSubscribers(ctx, map[string]string{
		"customerId":           customerID,
		"subscribedEventTypes": eventType,
	}, visitor)
}

func (resolver *Resolver) streamInterestedGroups(ctx context.Context, customerID, eventType string, visitor func(domain.Subscriber) error) error {
	page := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		list, err := resolver.stores.Groups.List(ctx, domain.ListOptions{
			Pagination: domain.PaginationParams{
				Page:     page,
				PageSize: resolver.PageSize,
			},
			Filter: map[string]string{
				"customerId":           customerID,
				"subscribedEventTypes": eventType,
			},
		})
		if err != nil {
			return fmt.Errorf("recipients: interested groups: %w", err)
		}

		if len(list.Items) == 0 {
			break
		}

		for _, group := range list.Items {
			if err := resolver.streamSubscriberIDs(ctx, customerID, group.Subscribers, visitor); err != nil {
				return err
			}
		}

		page++
		if page >= list.Pagination.TotalPages {
			break
		}
	}
	return nil
}

// streamPaginatedSubscribers streams subscribers from paginated list results
func (resolver *Resolver) streamPaginatedSubscribers(ctx context.Context, filter map[string]string, visitor func(domain.Subscriber) error) error {
	page := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		list, err := resolver.stores.Subscribers.List(ctx, domain.ListOptions{
			Pagination: domain.PaginationParams{
				Page:     page,
				PageSize: resolver.PageSize,
			},
			Filter: filter,
		})
		if err != nil {
			return fmt.Errorf("recipients: subscriber list: %w", err)
		}

		if len(list.Items) == 0 {
			break
		}

		for _, subscriber := range list.Items {
			_ = resolver.cache.Set(ctx, subscriber)
			if err := visitor(subscriber); err != nil {
				return err
			}
		}

		page++
		if page >= list.Pagination.TotalPages {
			break
		}
	}
	return nil
}

func createDirectEmailSubscriber(customerID, email string) domain.Subscriber {
	return domain.Subscriber{
		ID:         "", // Empty ID for direct recipients, deduplication by email
		CustomerID: customerID,
		Email:      email,
		Preferences: domain.SubscriberPrefs{
			Channels: map[domain.ChannelType]bool{
				domain.ChannelEmail: true,
				domain.ChannelSMS:   false,
				domain.ChannelPush:  false,
			},
		},
		CreatedAt: time.Now().UTC(),
	}
}

func createDirectPhoneSubscriber(customerID, phone string) domain.Subscriber {
	return domain.Subscriber{
		ID:         "", // Empty ID for direct recipients, deduplication by phone
		CustomerID: customerID,
		Phone:      phone,
		Preferences: domain.SubscriberPrefs{
			Channels: map[domain.ChannelType]bool{
				domain.ChannelEmail: false,
				domain.ChannelSMS:   true,
				domain.ChannelPush:  false,
			},
		},
		CreatedAt: time.Now().UTC(),
	}
}
