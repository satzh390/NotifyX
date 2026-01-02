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

type DeliveryTaskRepository struct {
	collection *mongoDriver.Collection
}

func (repo *DeliveryTaskRepository) Put(ctx context.Context, task domain.DeliveryTask) error {
	filter := bson.M{"taskId": task.TaskID}
	updateMap, err := BuildUpdateMap(task)
	if err != nil {
		return err
	}

	updateMap["createdAt"] = time.Now()
	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"taskId": task.TaskID,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (repo *DeliveryTaskRepository) Get(ctx context.Context, taskID string) (domain.DeliveryTask, error) {
	filter := bson.M{"taskId": taskID}
	var task domain.DeliveryTask
	err := repo.collection.FindOne(ctx, filter).Decode(&task)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.DeliveryTask{}, storage.ErrNotFound
	}
	return task, err
}

func (repo *DeliveryTaskRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.DeliveryTask], error) {
	filter := BuildBsonFilter(opts)
	page, pageSize := PageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.DeliveryTask]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(BuildBsonSort(opts, map[string]int{"createdAt": -1}))
	cursor, err := repo.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return domain.ListResult[domain.DeliveryTask]{}, fmt.Errorf("mongo: find: %w", err)
	}

	defer cursor.Close(ctx)
	var tasks []domain.DeliveryTask
	if err := cursor.All(ctx, &tasks); err != nil {
		return domain.ListResult[domain.DeliveryTask]{}, fmt.Errorf("mongo: decode: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	return domain.ListResult[domain.DeliveryTask]{
		Items: tasks,
		Pagination: domain.PaginationResult{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

