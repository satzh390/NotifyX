package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notifyx/core/domain"
)

func setupSubscriberTestApp() (*fiber.App, *mockSubscriberStore) {
	app := fiber.New()
	store := &mockSubscriberStore{subscribers: make(map[string]domain.Subscriber)}
	handler := NewSubscriberHandler(store)

	api := app.Group("/api/v1")
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("orgId", "test-org")
		return c.Next()
	})

	subscribers := api.Group("/subscribers")
	subscribers.Post("", handler.Create)
	subscribers.Get("/:id", handler.Get)
	subscribers.Put("/:id", handler.Update)
	subscribers.Delete("/:id", handler.Delete)
	subscribers.Get("", handler.List)

	return app, store
}

func TestSubscriberHandler_Create(t *testing.T) {
	app, _ := setupSubscriberTestApp()

	body := map[string]interface{}{
		"email": "test@example.com",
		"preferences": map[string]interface{}{
			"channels": map[string]bool{"email": true},
			"language": "en",
		},
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscribers", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var subscriber domain.Subscriber
	if err := json.NewDecoder(resp.Body).Decode(&subscriber); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if subscriber.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", subscriber.Email)
	}
	if subscriber.OrgID != "test-org" {
		t.Errorf("Expected orgID test-org, got %s", subscriber.OrgID)
	}
}

func TestSubscriberHandler_Get(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()
	store.Put(context.Background(), domain.Subscriber{
		ID:    subID,
		OrgID: "test-org",
		Email: "test@example.com",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers/"+subID, nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var subscriber domain.Subscriber
	if err := json.NewDecoder(resp.Body).Decode(&subscriber); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if subscriber.ID != subID {
		t.Errorf("Expected ID %s, got %s", subID, subscriber.ID)
	}
	if subscriber.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", subscriber.Email)
	}
}

func TestSubscriberHandler_Get_NotFound(t *testing.T) {
	app, _ := setupSubscriberTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers/non-existent", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestSubscriberHandler_Update(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()
	store.Put(context.Background(), domain.Subscriber{
		ID:    subID,
		OrgID: "test-org",
		Email: "old@example.com",
	})

	patch := map[string]interface{}{
		"email": "new@example.com",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/subscribers/"+subID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Subscriber
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Email != "new@example.com" {
		t.Errorf("Expected email new@example.com, got %s", updated.Email)
	}
}

func TestSubscriberHandler_Delete(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()
	store.Put(context.Background(), domain.Subscriber{
		ID:    subID,
		OrgID: "test-org",
		Email: "test@example.com",
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscribers/"+subID, nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify deleted
	_, err = store.Get(context.Background(), "test-org", subID)
	if err == nil {
		t.Error("Expected subscriber to be deleted")
	}
}

func TestSubscriberHandler_Delete_NotFound(t *testing.T) {
	app, _ := setupSubscriberTestApp()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscribers/non-existent", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestSubscriberHandler_List(t *testing.T) {
	app, store := setupSubscriberTestApp()

	store.lists = []domain.ListResult[domain.Subscriber]{
		{
			Items: []domain.Subscriber{
				{ID: "sub-1", OrgID: "test-org", Email: "test1@example.com", CreatedAt: time.Now()},
				{ID: "sub-2", OrgID: "test-org", Email: "test2@example.com", CreatedAt: time.Now()},
			},
			Pagination: domain.PaginationResult{
				Page:       1,
				PageSize:   20,
				TotalCount: 2,
				TotalPages: 1,
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?page=1&pageSize=20", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result domain.ListResult[domain.Subscriber]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}
	if result.Pagination.TotalCount != 2 {
		t.Errorf("Expected totalCount 2, got %d", result.Pagination.TotalCount)
	}
}

func TestSubscriberHandler_List_WithGroupFilter(t *testing.T) {
	app, store := setupSubscriberTestApp()

	store.lists = []domain.ListResult[domain.Subscriber]{
		{
			Items: []domain.Subscriber{
				{ID: "sub-1", OrgID: "test-org", Groups: []string{"group-123"}},
			},
			Pagination: domain.PaginationResult{
				Page:       1,
				PageSize:   20,
				TotalCount: 1,
				TotalPages: 1,
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?groupId=group-123", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestSubscriberHandler_List_WithSorting(t *testing.T) {
	app, store := setupSubscriberTestApp()

	store.lists = []domain.ListResult[domain.Subscriber]{
		{
			Items: []domain.Subscriber{
				{ID: "sub-1", OrgID: "test-org", CreatedAt: time.Now()},
			},
			Pagination: domain.PaginationResult{
				Page:       1,
				PageSize:   20,
				TotalCount: 1,
				TotalPages: 1,
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?sortBy=createdAt:desc", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

