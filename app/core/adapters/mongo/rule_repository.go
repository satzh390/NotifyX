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
	filter := BuildBsonFilter(opts)
	page, pageSize := PageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Rule]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(BuildBsonSort(opts, map[string]int{"createdAt": -1}))
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

