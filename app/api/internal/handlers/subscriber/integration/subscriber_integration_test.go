//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	mongoadapter "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/core/domain"
)

func TestIntegration_Subscriber_Create(t *testing.T) {
	app, cleanup := setupSubscriberIntegrationApp(t)
	defer cleanup()

	createBody := map[string]interface{}{
		"email": "integration@example.com",
		"preferences": map[string]interface{}{
			"channels":    map[string]bool{"email": true},
			"language":    "en",
			"allowedDays": []string{"monday", "tuesday"},
			"notificationWindow": map[string]string{
				"start": "09:00",
				"end":   "17:00",
			},
		},
	}
	bodyJSON, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscribers", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Create request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var created domain.Subscriber
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	if created.OrgID != "test-org-integration" {
		t.Errorf("Expected orgID test-org-integration, got %s", created.OrgID)
	}
	if created.Email != "integration@example.com" {
		t.Errorf("Expected email integration@example.com, got %s", created.Email)
	}
}

func TestIntegration_Subscriber_Get(t *testing.T) {
	app, cleanup := setupSubscriberIntegrationApp(t)
	defer cleanup()

	// Create a subscriber first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	subID := uuid.NewString()
	sub := domain.Subscriber{
		ID:        subID,
		OrgID:     "test-org-integration",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}
	if err := stores.Subscribers.Put(ctx, sub); err != nil {
		t.Fatalf("Failed to create subscriber: %v", err)
	}

	// Get the subscriber
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers/"+subID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var retrieved domain.Subscriber
	if err := json.NewDecoder(resp.Body).Decode(&retrieved); err != nil {
		t.Fatalf("Failed to decode get response: %v", err)
	}

	if retrieved.ID != subID {
		t.Errorf("Expected ID %s, got %s", subID, retrieved.ID)
	}
}

func TestIntegration_Subscriber_Get_NotFound(t *testing.T) {
	app, cleanup := setupSubscriberIntegrationApp(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers/non-existent", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestIntegration_Subscriber_Update(t *testing.T) {
	app, cleanup := setupSubscriberIntegrationApp(t)
	defer cleanup()

	// Create a subscriber first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	subID := uuid.NewString()
	sub := domain.Subscriber{
		ID:        subID,
		OrgID:     "test-org-integration",
		Email:     "old@example.com",
		CreatedAt: time.Now(),
	}
	if err := stores.Subscribers.Put(ctx, sub); err != nil {
		t.Fatalf("Failed to create subscriber: %v", err)
	}

	// Update the subscriber
	updateBody := map[string]interface{}{
		"email": "updated@example.com",
	}
	updateJSON, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/subscribers/"+subID, bytes.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Update request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Subscriber
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode update response: %v", err)
	}

	if updated.Email != "updated@example.com" {
		t.Errorf("Expected email updated@example.com, got %s", updated.Email)
	}
}

func TestIntegration_Subscriber_Delete(t *testing.T) {
	app, cleanup := setupSubscriberIntegrationApp(t)
	defer cleanup()

	// Create a subscriber first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	subID := uuid.NewString()
	sub := domain.Subscriber{
		ID:        subID,
		OrgID:     "test-org-integration",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}
	if err := stores.Subscribers.Put(ctx, sub); err != nil {
		t.Fatalf("Failed to create subscriber: %v", err)
	}

	// Delete the subscriber
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscribers/"+subID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Delete request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify deleted
	req = httptest.NewRequest(http.MethodGet, "/api/v1/subscribers/"+subID, nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Get after delete request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 after delete, got %d", resp.StatusCode)
	}
}

func TestIntegration_Subscriber_List(t *testing.T) {
	app, cleanup := setupSubscriberIntegrationApp(t)
	defer cleanup()

	// Create multiple subscribers
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	for i := 0; i < 5; i++ {
		sub := domain.Subscriber{
			ID:        uuid.NewString(),
			OrgID:     "test-org-integration",
			Email:     fmt.Sprintf("test%d@example.com", i),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		_ = stores.Subscribers.Put(ctx, sub)
	}

	// List subscribers
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?page=0&pageSize=20", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var listResult domain.ListResult[domain.Subscriber]
	if err := json.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}

	if listResult.Pagination.TotalCount < 5 {
		t.Errorf("Expected at least 5 subscribers, got %d", listResult.Pagination.TotalCount)
	}
}

func TestIntegration_Subscriber_List_WithPagination(t *testing.T) {
	app, cleanup := setupSubscriberIntegrationApp(t)
	defer cleanup()

	// Create multiple subscribers
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	for i := 0; i < 25; i++ {
		sub := domain.Subscriber{
			ID:        uuid.NewString(),
			OrgID:     "test-org-integration",
			Email:     fmt.Sprintf("test%d@example.com", i),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		_ = stores.Subscribers.Put(ctx, sub)
	}

	// Test pagination - page 0 (first page)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?page=0&pageSize=10", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result domain.ListResult[domain.Subscriber]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result.Items) != 10 {
		t.Errorf("Expected 10 items, got %d", len(result.Items))
	}
	if result.Pagination.Page != 0 {
		t.Errorf("Expected page 0, got %d", result.Pagination.Page)
	}
	if result.Pagination.PageSize != 10 {
		t.Errorf("Expected pageSize 10, got %d", result.Pagination.PageSize)
	}
	if result.Pagination.TotalCount < 25 {
		t.Errorf("Expected totalCount >= 25, got %d", result.Pagination.TotalCount)
	}

	// Test pagination - page 1 (second page)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?page=1&pageSize=10", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result2 domain.ListResult[domain.Subscriber]
	if err := json.NewDecoder(resp.Body).Decode(&result2); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result2.Items) != 10 {
		t.Errorf("Expected 10 items on page 1, got %d", len(result2.Items))
	}
	if result2.Pagination.Page != 1 {
		t.Errorf("Expected page 1, got %d", result2.Pagination.Page)
	}
}

func TestIntegration_Subscriber_List_WithGroupFilter(t *testing.T) {
	app, cleanup := setupSubscriberIntegrationApp(t)
	defer cleanup()

	// Create subscribers with different groups
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	groupID := "group-123"
	sub1 := domain.Subscriber{
		ID:        uuid.NewString(),
		OrgID:     "test-org-integration",
		Groups:    []string{groupID},
		CreatedAt: time.Now(),
	}
	sub2 := domain.Subscriber{
		ID:        uuid.NewString(),
		OrgID:     "test-org-integration",
		Groups:    []string{"other-group"},
		CreatedAt: time.Now(),
	}
	_ = stores.Subscribers.Put(ctx, sub1)
	_ = stores.Subscribers.Put(ctx, sub2)

	// List with group filter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?groupId="+groupID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result domain.ListResult[domain.Subscriber]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Pagination.TotalCount < 1 {
		t.Errorf("Expected at least 1 subscriber in group, got %d", result.Pagination.TotalCount)
	}
}

func TestIntegration_Subscriber_List_WithSorting(t *testing.T) {
	app, cleanup := setupSubscriberIntegrationApp(t)
	defer cleanup()

	// Create subscribers
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	for i := 0; i < 5; i++ {
		sub := domain.Subscriber{
			ID:        uuid.NewString(),
			OrgID:     "test-org-integration",
			Email:     fmt.Sprintf("test%d@example.com", i),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		_ = stores.Subscribers.Put(ctx, sub)
	}

	// List with sorting
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?sortBy=createdAt:desc", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
