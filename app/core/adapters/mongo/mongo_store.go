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
		Templates:   &TemplateRepository{collection: db.Collection(prefix + "templates")},
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
	filter := bson.M{"orgId": subscriber.OrgID, "subscriberId": subscriber.ID}
	updateMap, err := BuildUpdateMap(subscriber)
	if err != nil {
		return err
	}

	updateMap["updatedAt"] = time.Now()
	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"createdAt":    time.Now(),
			"orgId":        subscriber.OrgID,
			"subscriberId": subscriber.ID,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
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
	filter := buildBsonFilter(opts)
	page, pageSize := pageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Subscriber]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(buildBsonSort(opts, map[string]int{"createdAt": -1}))
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
	updateMap, err := BuildUpdateMap(group)
	if err != nil {
		return err
	}

	updateMap["updatedAt"] = time.Now()
	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
			"orgId":     group.OrgID,
			"groupId":   group.ID,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
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
	filter := buildBsonFilter(opts)
	page, pageSize := pageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Group]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(buildBsonSort(opts, map[string]int{"createdAt": -1}))
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
	filter := bson.M{"orgId": rule.OrgID, "eventType": rule.EventType}
	updateMap, err := BuildUpdateMap(rule)
	if err != nil {
		return err
	}

	updateMap["updatedAt"] = time.Now()
	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
			"orgId":     rule.OrgID,
			"eventType": rule.EventType,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
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
	filter := buildBsonFilter(opts)
	page, pageSize := pageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Rule]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(buildBsonSort(opts, map[string]int{"createdAt": -1}))
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

func buildBsonFilter(opts domain.ListOptions) bson.M {
	filter := bson.M{}
	for key, value := range opts.Filter {
		if value != "" {
			filter[key] = value
		}
	}

	return filter
}

func buildBsonSort(opts domain.ListOptions, defaultSort map[string]int) bson.D {
	sortMap := make(map[string]int)
	for k, v := range defaultSort {
		sortMap[k] = v
	}

	for _, s := range opts.SortBy {
		order := 1
		if s.Order == domain.SortDesc {
			order = -1
		}
		sortMap[s.Field] = order
	}

	// Build bson.D in correct order:
	// 1. User-specified fields (in order provided)
	finalSort := bson.D{}
	used := map[string]bool{}
	for _, s := range opts.SortBy {
		finalSort = append(finalSort, bson.E{
			Key:   s.Field,
			Value: sortMap[s.Field],
		})
		used[s.Field] = true
	}

	// 2. Default fields not specified by user
	for k, v := range defaultSort {
		if !used[k] {
			finalSort = append(finalSort, bson.E{Key: k, Value: v})
		}
	}

	return finalSort
}

func pageOrDefaultParam(opts domain.ListOptions) (page int, pageSize int) {
	// Set default pagination if not provided (0-based)
	page = max(0, opts.Pagination.Page)
	pageSize = opts.Pagination.PageSize
	if pageSize < 1 {
		pageSize = 20 // default page size
	}

	pageSize = min(100, pageSize)
	return
}

type TemplateRepository struct {
	collection *mongoDriver.Collection
}

func (repo *TemplateRepository) Put(ctx context.Context, template domain.Template) error {
	// Template ID is unique per org, channel is a property of the template
	filter := bson.M{
		"orgId": template.OrgID,
		"id":    template.ID,
	}
	updateMap, err := BuildUpdateMap(template)
	if err != nil {
		return err
	}

	updateMap["updatedAt"] = time.Now()
	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
			"orgId":     template.OrgID,
			"id":        template.ID,
			"channel":   template.Channel,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (repo *TemplateRepository) Get(ctx context.Context, orgID, templateID string) (domain.Template, error) {
	filter := bson.M{
		"orgId": orgID,
		"id":    templateID,
	}
	var template domain.Template
	err := repo.collection.FindOne(ctx, filter).Decode(&template)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.Template{}, storage.ErrNotFound
	}

	if err != nil {
		return domain.Template{}, err
	}

	return template, nil
}

func (repo *TemplateRepository) GetByLanguage(ctx context.Context, orgID, templateID, language string) (domain.Template, error) {
	template, err := repo.Get(ctx, orgID, templateID)
	if err != nil {
		return domain.Template{}, err
	}

	// If language is specified and translation exists, merge it into content
	if language != "" && template.Translations != nil {
		if translatedContent, ok := template.Translations[language]; ok {
			// Merge translated content into the main content
			// Translated content takes precedence
			if translatedContent.Body != "" {
				template.Content.Body = translatedContent.Body
			}
			if translatedContent.Subject != "" {
				template.Content.Subject = translatedContent.Subject
			}
			if translatedContent.Title != "" {
				template.Content.Title = translatedContent.Title
			}
			if translatedContent.Payload != nil {
				template.Content.Payload = translatedContent.Payload
			}
		}
	}

	return template, nil
}

func (repo *TemplateRepository) Delete(ctx context.Context, orgID, templateID string) error {
	// Delete all variants (all channels and languages)
	filter := bson.M{
		"orgId": orgID,
		"id":    templateID,
	}
	result, err := repo.collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}
