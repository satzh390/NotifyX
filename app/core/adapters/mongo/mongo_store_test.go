package mongo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
)

// Integration test - requires MongoDB running
func TestSubscriberRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	stores, cleanup, err := NewStoreSet(ctx, Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer cleanup(ctx)

	repo := stores.Subscribers
	orgID := "test-org"
	subID := "test-sub-1"

	// Create
	sub := domain.Subscriber{
		ID:        subID,
		OrgID:     orgID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}
	if err := repo.Put(ctx, sub); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get
	retrieved, err := repo.Get(ctx, orgID, subID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", retrieved.Email)
	}

	// Update
	sub.Email = "updated@example.com"
	if err := repo.Put(ctx, sub); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, err := repo.Get(ctx, orgID, subID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if updated.Email != "updated@example.com" {
		t.Errorf("Expected email updated@example.com, got %s", updated.Email)
	}

	// List
	opts := domain.ListOptions{
		OrgID: orgID,
		Pagination: domain.PaginationParams{
			Page:     1,
			PageSize: 20,
		},
	}
	result, err := repo.List(ctx, opts)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if result.Pagination.TotalCount < 1 {
		t.Errorf("Expected at least 1 subscriber, got %d", result.Pagination.TotalCount)
	}

	// Delete
	if err := repo.Delete(ctx, orgID, subID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = repo.Get(ctx, orgID, subID)
	if err != storage.ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}

func TestSubscriberRepository_ListWithPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	stores, cleanup, err := NewStoreSet(ctx, Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer cleanup(ctx)

	repo := stores.Subscribers
	orgID := "test-org-pagination"

	// Create multiple subscribers
	for i := 0; i < 25; i++ {
		sub := domain.Subscriber{
			ID:        fmt.Sprintf("sub-%d", i),
			OrgID:     orgID,
			Email:     fmt.Sprintf("test%d@example.com", i),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		_ = repo.Put(ctx, sub)
	}

	// Test pagination
	opts := domain.ListOptions{
		OrgID: orgID,
		Pagination: domain.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
		SortBy: []domain.SortOption{
			{Field: "createdAt", Order: domain.SortDesc},
		},
	}

	result, err := repo.List(ctx, opts)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(result.Items) != 10 {
		t.Errorf("Expected 10 items, got %d", len(result.Items))
	}
	if result.Pagination.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Pagination.Page)
	}
	if result.Pagination.TotalCount < 25 {
		t.Errorf("Expected totalCount >= 25, got %d", result.Pagination.TotalCount)
	}
}

func TestSubscriberRepository_ListWithGroupFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	stores, cleanup, err := NewStoreSet(ctx, Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer cleanup(ctx)

	repo := stores.Subscribers
	orgID := "test-org-group-filter"
	groupID := "group-123"

	// Create subscribers with different groups
	sub1 := domain.Subscriber{
		ID:      "sub-1",
		OrgID:   orgID,
		Groups:  []string{groupID},
		CreatedAt: time.Now(),
	}
	sub2 := domain.Subscriber{
		ID:      "sub-2",
		OrgID:   orgID,
		Groups:  []string{"other-group"},
		CreatedAt: time.Now(),
	}
	sub3 := domain.Subscriber{
		ID:      "sub-3",
		OrgID:   orgID,
		Groups:  []string{groupID, "other-group"},
		CreatedAt: time.Now(),
	}

	_ = repo.Put(ctx, sub1)
	_ = repo.Put(ctx, sub2)
	_ = repo.Put(ctx, sub3)

	// Filter by group
	opts := domain.ListOptions{
		OrgID:   orgID,
		GroupID: groupID,
		Pagination: domain.PaginationParams{
			Page:     1,
			PageSize: 20,
		},
	}

	result, err := repo.List(ctx, opts)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result.Pagination.TotalCount != 2 {
		t.Errorf("Expected 2 subscribers in group, got %d", result.Pagination.TotalCount)
	}
}

