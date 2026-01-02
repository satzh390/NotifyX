package mongo

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"go.mongodb.org/mongo-driver/bson"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeliveryLogRepository struct {
	collection *mongoDriver.Collection
}

func (repo *DeliveryLogRepository) Put(ctx context.Context, log domain.DeliveryLog) error {
	filter := bson.M{"taskId": log.TaskID, "timestamp": log.Timestamp}
	updateMap, err := BuildUpdateMap(log)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"taskId":    log.TaskID,
			"timestamp": log.Timestamp,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (repo *DeliveryLogRepository) Get(ctx context.Context, taskID string) (domain.DeliveryLog, error) {
	filter := bson.M{"taskId": taskID}
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1}) // Get latest log
	var log domain.DeliveryLog
	err := repo.collection.FindOne(ctx, filter, opts).Decode(&log)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.DeliveryLog{}, storage.ErrNotFound
	}
	return log, err
}

func (repo *DeliveryLogRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.DeliveryLog], error) {
	filter := BuildBsonFilter(opts)
	page, pageSize := PageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.DeliveryLog]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(BuildBsonSort(opts, map[string]int{"timestamp": -1}))
	cursor, err := repo.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return domain.ListResult[domain.DeliveryLog]{}, fmt.Errorf("mongo: find: %w", err)
	}

	defer cursor.Close(ctx)
	var logs []domain.DeliveryLog
	if err := cursor.All(ctx, &logs); err != nil {
		return domain.ListResult[domain.DeliveryLog]{}, fmt.Errorf("mongo: decode: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	return domain.ListResult[domain.DeliveryLog]{
		Items: logs,
		Pagination: domain.PaginationResult{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

