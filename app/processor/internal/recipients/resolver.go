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
	pageSize int
}

func NewResolver(stores storage.Stores, cacheProvider cache.SubscriberCache) *Resolver {
	if cacheProvider == nil {
		cacheProvider = cache.NoopSubscriberCache{}
	}
	return &Resolver{
		stores:   stores,
		cache:    cacheProvider,
		pageSize: defaultPageSize,
	}
}

func (resolver *Resolver) SetPageSize(size int) {
	if size > 0 {
		resolver.pageSize = size
	}
}

func (resolver *Resolver) Stream(ctx context.Context, orgID string, recipients domain.Recipients, visitor func(domain.Subscriber) error) error {
	if err := resolver.streamSubscriberIDs(ctx, orgID, recipients.SubscriberIDs, visitor); err != nil {
		return err
	}
	if err := resolver.streamGroups(ctx, orgID, recipients.Groups, visitor); err != nil {
		return err
	}
	if recipients.Broadcast {
		if err := resolver.streamBroadcast(ctx, orgID, visitor); err != nil {
			return err
		}
	}
	if err := resolver.streamDirectEmails(ctx, orgID, recipients.DirectEmails, visitor); err != nil {
		return err
	}
	if err := resolver.streamDirectPhones(ctx, orgID, recipients.DirectPhones, visitor); err != nil {
		return err
	}
	return nil
}

func (resolver *Resolver) streamSubscriberIDs(ctx context.Context, orgID string, ids []string, visitor func(domain.Subscriber) error) error {
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		if id == "" {
			continue
		}

		subscriber, ok, err := resolver.getSubscriber(ctx, orgID, id)
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

func (resolver *Resolver) streamGroups(ctx context.Context, orgID string, groupIDs []string, visitor func(domain.Subscriber) error) error {
	for _, groupID := range groupIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if groupID == "" {
			continue
		}
		group, err := resolver.stores.Groups.Get(ctx, orgID, groupID)
		if err != nil {
			if err == storage.ErrNotFound {
				continue
			}
			return fmt.Errorf("recipients: group %s: %w", groupID, err)
		}
		if err := resolver.streamSubscriberIDs(ctx, orgID, group.Subscribers, visitor); err != nil {
			return err
		}
	}
	return nil
}

func (resolver *Resolver) streamBroadcast(ctx context.Context, orgID string, visitor func(domain.Subscriber) error) error {
	page := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		list, err := resolver.stores.Subscribers.List(ctx, domain.ListOptions{
			Pagination: domain.PaginationParams{
				Page:     page,
				PageSize: resolver.pageSize,
			},
			Filter: map[string]string{
				"orgId": orgID,
			},
		})
		if err != nil {
			return fmt.Errorf("recipients: broadcast list: %w", err)
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

func (resolver *Resolver) getSubscriber(ctx context.Context, orgID, subscriberID string) (domain.Subscriber, bool, error) {
	if cached, ok := resolver.cache.Get(ctx, orgID, subscriberID); ok {
		return cached, true, nil
	}

	subscriber, err := resolver.stores.Subscribers.Get(ctx, orgID, subscriberID)
	if err != nil {
		if err == storage.ErrNotFound {
			return domain.Subscriber{}, false, nil
		}
		return domain.Subscriber{}, false, fmt.Errorf("recipients: subscriber %s: %w", subscriberID, err)
	}

	_ = resolver.cache.Set(ctx, subscriber)
	return subscriber, true, nil
}

func (resolver *Resolver) streamDirectEmails(ctx context.Context, orgID string, emails []string, visitor func(domain.Subscriber) error) error {
	for _, email := range emails {
		if err := ctx.Err(); err != nil {
			return err
		}
		if email == "" {
			continue
		}
		subscriber := createDirectEmailSubscriber(orgID, email)
		if err := visitor(subscriber); err != nil {
			return err
		}
	}
	return nil
}

func (resolver *Resolver) streamDirectPhones(ctx context.Context, orgID string, phones []string, visitor func(domain.Subscriber) error) error {
	for _, phone := range phones {
		if err := ctx.Err(); err != nil {
			return err
		}
		if phone == "" {
			continue
		}
		subscriber := createDirectPhoneSubscriber(orgID, phone)
		if err := visitor(subscriber); err != nil {
			return err
		}
	}
	return nil
}

func createDirectEmailSubscriber(orgID, email string) domain.Subscriber {
	return domain.Subscriber{
		ID:    "", // Empty ID for direct recipients, deduplication by email
		OrgID: orgID,
		Email: email,
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

func createDirectPhoneSubscriber(orgID, phone string) domain.Subscriber {
	return domain.Subscriber{
		ID:    "", // Empty ID for direct recipients, deduplication by phone
		OrgID: orgID,
		Phone: phone,
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
