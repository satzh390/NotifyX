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
	filter := BuildBsonFilter(opts)
	page, pageSize := PageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Group]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(BuildBsonSort(opts, map[string]int{"createdAt": -1}))
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

