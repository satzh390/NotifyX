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

type CustomerRepository struct {
	collection *mongoDriver.Collection
}

func (repo *CustomerRepository) Put(ctx context.Context, customer domain.Customer) error {
	filter := bson.M{"id": customer.ID}
	updateMap, err := BuildUpdateMap(customer)
	if err != nil {
		return err
	}

	updateMap["updatedAt"] = time.Now()
	update := bson.M{
		"$set": updateMap,
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
			"id":        customer.ID,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = repo.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (repo *CustomerRepository) Get(ctx context.Context, customerID string) (domain.Customer, error) {
	filter := bson.M{"id": customerID}
	var customer domain.Customer
	err := repo.collection.FindOne(ctx, filter).Decode(&customer)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return domain.Customer{}, storage.ErrNotFound
	}
	return customer, err
}

func (repo *CustomerRepository) List(ctx context.Context, opts domain.ListOptions) (domain.ListResult[domain.Customer], error) {
	filter := BuildBsonFilter(opts)
	page, pageSize := PageOrDefaultParam(opts)
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return domain.ListResult[domain.Customer]{}, fmt.Errorf("mongo: count: %w", err)
	}

	skip := int64(page * pageSize)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize))
	findOpts.SetSort(BuildBsonSort(opts, map[string]int{"createdAt": -1}))
	cursor, err := repo.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return domain.ListResult[domain.Customer]{}, fmt.Errorf("mongo: find: %w", err)
	}

	defer cursor.Close(ctx)
	var customers []domain.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return domain.ListResult[domain.Customer]{}, fmt.Errorf("mongo: decode: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	return domain.ListResult[domain.Customer]{
		Items: customers,
		Pagination: domain.PaginationResult{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

func (repo *CustomerRepository) Delete(ctx context.Context, customerID string) error {
	filter := bson.M{"id": customerID}
	result, err := repo.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

