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

type OrganizationRepository struct {
	collection *mongoDriver.Collection
}

func (repo *OrganizationRepository) Put(ctx context.Context, org domain.Organization) error {
	filter := bson.M{"id": org.ID}
	updateMap, err := BuildUpdateMap(org)
	if err != nil {
		return err
	}

	updateMap["updatedAt"] = time.Now()
	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
			"id":        org.ID,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (repo *OrganizationRepository) Get(ctx context.Context, orgID string) (domain.Organization, error) {
	filter := bson.M{"id": orgID}
	var org domain.Organization
	err := repo.collection.FindOne(ctx, filter).Decode(&org)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.Organization{}, storage.ErrNotFound
	}
	return org, err
}

func (repo *OrganizationRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Organization], error) {
	filter := BuildBsonFilter(opts)
	page, pageSize := PageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Organization]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(BuildBsonSort(opts, map[string]int{"createdAt": -1}))
	cursor, err := repo.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return domain.ListResult[domain.Organization]{}, fmt.Errorf("mongo: find: %w", err)
	}

	defer cursor.Close(ctx)
	var orgs []domain.Organization
	if err := cursor.All(ctx, &orgs); err != nil {
		return domain.ListResult[domain.Organization]{}, fmt.Errorf("mongo: decode: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	return domain.ListResult[domain.Organization]{
		Items: orgs,
		Pagination: domain.PaginationResult{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

func (repo *OrganizationRepository) Delete(ctx context.Context, orgID string) error {
	filter := bson.M{"id": orgID}
	result, err := repo.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}
