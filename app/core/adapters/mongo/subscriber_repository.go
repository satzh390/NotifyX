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
	filter := BuildBsonFilter(opts)
	page, pageSize := PageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Subscriber]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(BuildBsonSort(opts, map[string]int{"createdAt": -1}))
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

