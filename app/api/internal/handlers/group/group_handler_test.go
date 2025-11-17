package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notifyx/core/domain"
)

func setupGroupTestApp() (*fiber.App, *mockGroupStore) {
	app := fiber.New()
	store := &mockGroupStore{groups: make(map[string]domain.Group)}
	handler := NewGroupHandler(store)

	api := app.Group("/api/v1")
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("orgId", "test-org")
		return c.Next()
	})

	groups := api.Group("/groups")
	groups.Post("", handler.Create)
	groups.Get("/:id", handler.Get)
	groups.Put("/:id", handler.Update)
	groups.Delete("/:id", handler.Delete)
	groups.Get("", handler.List)

	return app, store
}

func TestGroupHandler_Create(t *testing.T) {
	app, _ := setupGroupTestApp()

	body := map[string]interface{}{
		"name":        "Test Group",
		"description": "Test Description",
		"subscribers": []string{"sub-1", "sub-2"},
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var group domain.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if group.Name != "Test Group" {
		t.Errorf("Expected name Test Group, got %s", group.Name)
	}
	if group.OrgID != "test-org" {
		t.Errorf("Expected orgID test-org, got %s", group.OrgID)
	}
}

func TestGroupHandler_Get(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()
	store.Put(context.Background(), domain.Group{
		ID:          groupID,
		OrgID:       "test-org",
		Name:        "Test Group",
		Description: "Test Description",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/"+groupID, nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var group domain.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if group.ID != groupID {
		t.Errorf("Expected ID %s, got %s", groupID, group.ID)
	}
	if group.Name != "Test Group" {
		t.Errorf("Expected name Test Group, got %s", group.Name)
	}
}

func TestGroupHandler_Get_NotFound(t *testing.T) {
	app, _ := setupGroupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/non-existent", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_Update(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()
	store.Put(context.Background(), domain.Group{
		ID:    groupID,
		OrgID: "test-org",
		Name:  "Old Name",
	})

	patch := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/groups/"+groupID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Group
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name New Name, got %s", updated.Name)
	}
}

func TestGroupHandler_Delete(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()
	store.Put(context.Background(), domain.Group{
		ID:    groupID,
		OrgID: "test-org",
		Name:  "Test Group",
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/"+groupID, nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify deleted
	_, err = store.Get(context.Background(), "test-org", groupID)
	if err == nil {
		t.Error("Expected group to be deleted")
	}
}

func TestGroupHandler_Delete_NotFound(t *testing.T) {
	app, _ := setupGroupTestApp()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/non-existent", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_List(t *testing.T) {
	app, store := setupGroupTestApp()

	// Add some groups
	store.Put(context.Background(), domain.Group{
		ID:    "group-1",
		OrgID: "test-org",
		Name:  "Group 1",
	})
	store.Put(context.Background(), domain.Group{
		ID:    "group-2",
		OrgID: "test-org",
		Name:  "Group 2",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups?page=1&pageSize=20", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result domain.ListResult[domain.Group]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Pagination.TotalCount < 2 {
		t.Errorf("Expected at least 2 groups, got %d", result.Pagination.TotalCount)
	}
}

func TestGroupHandler_List_WithSorting(t *testing.T) {
	app, store := setupGroupTestApp()

	store.Put(context.Background(), domain.Group{
		ID:   "group-1",
		OrgID: "test-org",
		Name: "A Group",
	})
	store.Put(context.Background(), domain.Group{
		ID:   "group-2",
		OrgID: "test-org",
		Name: "B Group",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups?sortBy=name:asc", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

