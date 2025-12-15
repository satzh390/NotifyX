package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"go.mongodb.org/mongo-driver/bson"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppConfigRepository struct {
	collection *mongoDriver.Collection
}

func (repo *AppConfigRepository) Put(ctx context.Context, appConfig domain.AppConfig) error {
	filter := bson.M{"orgId": appConfig.OrgID, "id": appConfig.ID}
	updateMap, err := BuildUpdateMap(appConfig)
	if err != nil {
		return err
	}

	updateMap["updatedAt"] = time.Now()
	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
			"orgId":     appConfig.OrgID,
			"id":        appConfig.ID,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (repo *AppConfigRepository) Get(ctx context.Context, orgID, appID string) (domain.AppConfig, error) {
	filter := bson.M{"orgId": orgID, "id": appID}
	var appConfig domain.AppConfig
	err := repo.collection.FindOne(ctx, filter).Decode(&appConfig)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.AppConfig{}, storage.ErrNotFound
	}
	return appConfig, err
}

func (repo *AppConfigRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.AppConfig], error) {
	filter := bson.M{}
	
	// Apply filters
	for k, v := range opts.Filter {
		filter[k] = v
	}

	// Build sort options
	sortOptions := bson.D{}
	for _, sort := range opts.SortBy {
		order := 1
		if sort.Order == domain.SortDesc {
			order = -1
		}
		sortOptions = append(sortOptions, bson.E{Key: sort.Field, Value: order})
	}
	if len(sortOptions) == 0 {
		sortOptions = append(sortOptions, bson.E{Key: "createdAt", Value: -1})
	}

	// Count total
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.AppConfig]{}, fmt.Errorf("mongo: count documents: %w", err)
	}

	// Calculate pagination
	skip := opts.Pagination.Page * opts.Pagination.PageSize
	limit := opts.Pagination.PageSize
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))

	// Find documents
	findOpts := options.Find().
		SetSort(sortOptions).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := repo.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return domain.ListResult[domain.AppConfig]{}, fmt.Errorf("mongo: find: %w", err)
	}
	defer cursor.Close(ctx)

	var items []domain.AppConfig
	if err := cursor.All(ctx, &items); err != nil {
		return domain.ListResult[domain.AppConfig]{}, fmt.Errorf("mongo: decode: %w", err)
	}

	return domain.ListResult[domain.AppConfig]{
		Items: items,
		Pagination: domain.PaginationResult{
			Page:       opts.Pagination.Page,
			PageSize:   opts.Pagination.PageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

func (repo *AppConfigRepository) Delete(ctx context.Context, orgID, appID string) error {
	filter := bson.M{"orgId": orgID, "id": appID}
	result, err := repo.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

