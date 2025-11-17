package mongo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"go.mongodb.org/mongo-driver/bson"
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
		Subscribers: &SubscriberRepository{collection: db.Collection(prefix + "subscribers")},
		Groups:      &GroupRepository{collection: db.Collection(prefix + "groups")},
		Rules:       &RuleRepository{collection: db.Collection(prefix + "rules")},
		// Templates are stored in S3, not MongoDB
	}

	cleanup := func(ctx context.Context) error {
		return client.Disconnect(ctx)
	}

	return stores, cleanup, nil
}

type SubscriberRepository struct {
	collection *mongoDriver.Collection
}

func (repo *SubscriberRepository) Put(ctx context.Context, subscriber domain.Subscriber) error {
	// Set createdAt if it's zero (new record)
	if subscriber.CreatedAt.IsZero() {
		// Check if document exists
		existing, err := repo.Get(ctx, subscriber.OrgID, subscriber.ID)
		if err != nil && err != storage.ErrNotFound {
			return err
		}
		if err == storage.ErrNotFound {
			// New record, set createdAt
			subscriber.CreatedAt = time.Now()
		} else {
			// Existing record, preserve original createdAt
			subscriber.CreatedAt = existing.CreatedAt
		}
	}

	filter := bson.M{"orgId": subscriber.OrgID, "subscriberId": subscriber.ID}
	opts := options.Replace().SetUpsert(true)
	_, err := repo.collection.ReplaceOne(ctx, filter, subscriber, opts)
	return err
}

func (repo *SubscriberRepository) Get(ctx context.Context, orgID, subscriberID string) (domain.Subscriber, error) {
	filter := bson.M{"orgId": orgID, "subscriberId": subscriberID}
	var subscriber domain.Subscriber
	err := repo.collection.FindOne(ctx, filter).Decode(&subscriber)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.Subscriber{}, storage.ErrNotFound
	}
	return subscriber, err
}

func (repo *SubscriberRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Subscriber], error) {
	if opts.OrgID == "" {
		return domain.ListResult[domain.Subscriber]{}, errors.New("orgID is required")
	}

	// Build filter
	filter := bson.M{"orgId": opts.OrgID}
	if opts.GroupID != "" {
		// MongoDB matches if the value is in the array
		filter["groups"] = opts.GroupID
	}

	// Set default pagination if not provided
	page := opts.Pagination.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.Pagination.PageSize
	if pageSize < 1 {
		pageSize = 20 // default page size
	}
	if pageSize > 100 {
		pageSize = 100 // max page size
	}

	// Count total documents
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Subscriber]{}, fmt.Errorf("mongo: count: %w", err)
	}

	// Calculate skip
	skip := int64((page - 1) * pageSize)

	// Find with pagination
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))

	// Apply sorting if specified
	if len(opts.SortBy) > 0 {
		sortDoc := bson.D{}
		for _, sort := range opts.SortBy {
			order := 1 // ascending
			if sort.Order == domain.SortDesc {
				order = -1 // descending
			}
			sortDoc = append(sortDoc, bson.E{Key: sort.Field, Value: order})
		}
		findOpts.SetSort(sortDoc)
	} else {
		// Default sort by createdAt descending (newest first)
		findOpts.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	}

	cursor, err := repo.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return domain.ListResult[domain.Subscriber]{}, fmt.Errorf("mongo: find: %w", err)
	}
	defer cursor.Close(ctx)

	var subscribers []domain.Subscriber
	if err := cursor.All(ctx, &subscribers); err != nil {
		return domain.ListResult[domain.Subscriber]{}, fmt.Errorf("mongo: decode: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return domain.ListResult[domain.Subscriber]{
		Items: subscribers,
		Pagination: domain.PaginationResult{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

func (repo *SubscriberRepository) Delete(ctx context.Context, orgID, subscriberID string) error {
	filter := bson.M{"orgId": orgID, "subscriberId": subscriberID}
	result, err := repo.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

type GroupRepository struct {
	collection *mongoDriver.Collection
}

func (repo *GroupRepository) Put(ctx context.Context, group domain.Group) error {
	filter := bson.M{"orgId": group.OrgID, "groupId": group.ID}
	opts := options.Replace().SetUpsert(true)
	_, err := repo.collection.ReplaceOne(ctx, filter, group, opts)
	return err
}

func (repo *GroupRepository) Get(ctx context.Context, orgID, groupID string) (domain.Group, error) {
	filter := bson.M{"orgId": orgID, "groupId": groupID}
	var group domain.Group
	err := repo.collection.FindOne(ctx, filter).Decode(&group)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.Group{}, storage.ErrNotFound
	}
	return group, err
}

func (repo *GroupRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Group], error) {
	if opts.OrgID == "" {
		return domain.ListResult[domain.Group]{}, errors.New("orgID is required")
	}

	// Build filter
	filter := bson.M{"orgId": opts.OrgID}

	// Set default pagination if not provided
	page := opts.Pagination.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.Pagination.PageSize
	if pageSize < 1 {
		pageSize = 20 // default page size
	}
	if pageSize > 100 {
		pageSize = 100 // max page size
	}

	// Count total documents
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Group]{}, fmt.Errorf("mongo: count: %w", err)
	}

	// Calculate skip
	skip := int64((page - 1) * pageSize)

	// Find with pagination
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))

	// Apply sorting if specified
	if len(opts.SortBy) > 0 {
		sortDoc := bson.D{}
		for _, sort := range opts.SortBy {
			order := 1 // ascending
			if sort.Order == domain.SortDesc {
				order = -1 // descending
			}
			sortDoc = append(sortDoc, bson.E{Key: sort.Field, Value: order})
		}
		findOpts.SetSort(sortDoc)
	}

	cursor, err := repo.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return domain.ListResult[domain.Group]{}, fmt.Errorf("mongo: find: %w", err)
	}
	defer cursor.Close(ctx)

	var groups []domain.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return domain.ListResult[domain.Group]{}, fmt.Errorf("mongo: decode: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return domain.ListResult[domain.Group]{
		Items: groups,
		Pagination: domain.PaginationResult{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

func (repo *GroupRepository) Delete(ctx context.Context, orgID, groupID string) error {
	filter := bson.M{"orgId": orgID, "groupId": groupID}
	result, err := repo.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

type RuleRepository struct {
	collection *mongoDriver.Collection
}

func (repo *RuleRepository) Put(ctx context.Context, rule domain.Rule) error {
	// Set createdAt if it's zero (new record)
	if rule.CreatedAt.IsZero() {
		// Check if document exists
		existing, err := repo.Get(ctx, rule.OrgID, rule.EventType)
		if err != nil && err != storage.ErrNotFound {
			return err
		}
		if err == storage.ErrNotFound {
			// New record, set createdAt
			rule.CreatedAt = time.Now()
		} else {
			// Existing record, preserve original createdAt
			rule.CreatedAt = existing.CreatedAt
		}
	}

	filter := bson.M{"orgId": rule.OrgID, "eventType": rule.EventType}
	opts := options.Replace().SetUpsert(true)
	_, err := repo.collection.ReplaceOne(ctx, filter, rule, opts)
	return err
}

func (repo *RuleRepository) Get(ctx context.Context, orgID, eventType string) (domain.Rule, error) {
	filter := bson.M{"orgId": orgID, "eventType": eventType}
	var rule domain.Rule
	err := repo.collection.FindOne(ctx, filter).Decode(&rule)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.Rule{}, storage.ErrNotFound
	}
	return rule, err
}

func (repo *RuleRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Rule], error) {
	if opts.OrgID == "" {
		return domain.ListResult[domain.Rule]{}, errors.New("orgID is required")
	}

	// Build filter
	filter := bson.M{"orgId": opts.OrgID}

	// Set default pagination if not provided
	page := opts.Pagination.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.Pagination.PageSize
	if pageSize < 1 {
		pageSize = 20 // default page size
	}
	if pageSize > 100 {
		pageSize = 100 // max page size
	}

	// Count total documents
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Rule]{}, fmt.Errorf("mongo: count: %w", err)
	}

	// Calculate skip
	skip := int64((page - 1) * pageSize)

	// Find with pagination
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))

	// Apply sorting if specified
	if len(opts.SortBy) > 0 {
		sortDoc := bson.D{}
		for _, sort := range opts.SortBy {
			order := 1 // ascending
			if sort.Order == domain.SortDesc {
				order = -1 // descending
			}
			sortDoc = append(sortDoc, bson.E{Key: sort.Field, Value: order})
		}
		findOpts.SetSort(sortDoc)
	} else {
		// Default sort by createdAt descending (newest first)
		findOpts.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	}

	cursor, err := repo.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return domain.ListResult[domain.Rule]{}, fmt.Errorf("mongo: find: %w", err)
	}
	defer cursor.Close(ctx)

	var rules []domain.Rule
	if err := cursor.All(ctx, &rules); err != nil {
		return domain.ListResult[domain.Rule]{}, fmt.Errorf("mongo: decode: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return domain.ListResult[domain.Rule]{
		Items: rules,
		Pagination: domain.PaginationResult{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

func (repo *RuleRepository) Delete(ctx context.Context, orgID, eventType string) error {
	filter := bson.M{"orgId": orgID, "eventType": eventType}
	result, err := repo.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

