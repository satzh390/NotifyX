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

	"github.com/google/uuid"
	mongoadapter "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/core/domain"
)

func TestIntegration_Group_Create(t *testing.T) {
	app, cleanup := setupGroupIntegrationApp(t)
	defer cleanup()

	createBody := map[string]interface{}{
		"name":        "Integration Test Group",
		"description": "Test Description",
		"subscribers": []string{"sub-1", "sub-2"},
	}
	bodyJSON, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Create request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var created domain.Group
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	if created.Name != "Integration Test Group" {
		t.Errorf("Expected name Integration Test Group, got %s", created.Name)
	}
	if created.CustomerID != "test-customer-integration" {
		t.Errorf("Expected customerID test-customer-integration, got %s", created.CustomerID)
	}
}

func TestIntegration_Group_Get(t *testing.T) {
	app, cleanup := setupGroupIntegrationApp(t)
	defer cleanup()

	// Create a group first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	groupID := uuid.NewString()
	group := domain.Group{
		ID:    groupID,
		CustomerID: "test-customer-integration",
		Name:  "Test Group",
	}
	if err := stores.Groups.Put(ctx, group); err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Get the group
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/"+groupID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var retrieved domain.Group
	if err := json.NewDecoder(resp.Body).Decode(&retrieved); err != nil {
		t.Fatalf("Failed to decode get response: %v", err)
	}

	if retrieved.ID != groupID {
		t.Errorf("Expected ID %s, got %s", groupID, retrieved.ID)
	}
}

func TestIntegration_Group_Get_NotFound(t *testing.T) {
	app, cleanup := setupGroupIntegrationApp(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/non-existent", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestIntegration_Group_Update(t *testing.T) {
	app, cleanup := setupGroupIntegrationApp(t)
	defer cleanup()

	// Create a group first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	groupID := uuid.NewString()
	group := domain.Group{
		ID:    groupID,
		CustomerID: "test-customer-integration",
		Name:  "Old Name",
	}
	if err := stores.Groups.Put(ctx, group); err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Update the group
	updateBody := map[string]interface{}{
		"groupId": groupID,
		"name":    "Updated Group Name",
	}
	updateJSON, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/groups/"+groupID, bytes.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Update request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Group
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode update response: %v", err)
	}

	if updated.Name != "Updated Group Name" {
		t.Errorf("Expected name Updated Group Name, got %s", updated.Name)
	}
}

func TestIntegration_Group_Patch(t *testing.T) {
	app, cleanup := setupGroupIntegrationApp(t)
	defer cleanup()

	// Create a group first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	groupID := uuid.NewString()
	group := domain.Group{
		ID:          groupID,
		OrgID:       "test-org-integration",
		Name:        "Old Name",
		Description: "Old Description",
	}
	if err := stores.Groups.Put(ctx, group); err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Patch the group
	patchBody := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patchBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/groups/"+groupID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Patch request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var patched domain.Group
	if err := json.NewDecoder(resp.Body).Decode(&patched); err != nil {
		t.Fatalf("Failed to decode patch response: %v", err)
	}

	if patched.Name != "New Name" {
		t.Errorf("Expected name New Name, got %s", patched.Name)
	}
	// Description should remain unchanged
	if patched.Description != "Old Description" {
		t.Errorf("Expected description Old Description, got %s", patched.Description)
	}
}

func TestIntegration_Group_Patch_NotFound(t *testing.T) {
	app, cleanup := setupGroupIntegrationApp(t)
	defer cleanup()

	patchBody := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patchBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/groups/non-existent", bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Patch request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestIntegration_Group_Delete(t *testing.T) {
	app, cleanup := setupGroupIntegrationApp(t)
	defer cleanup()

	// Create a group first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	groupID := uuid.NewString()
	group := domain.Group{
		ID:    groupID,
		CustomerID: "test-customer-integration",
		Name:  "Test Group",
	}
	if err := stores.Groups.Put(ctx, group); err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Delete the group
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/"+groupID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Delete request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify deleted
	req = httptest.NewRequest(http.MethodGet, "/api/v1/groups/"+groupID, nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Get after delete request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 after delete, got %d", resp.StatusCode)
	}
}

func TestIntegration_Group_List(t *testing.T) {
	app, cleanup := setupGroupIntegrationApp(t)
	defer cleanup()

	// Create multiple groups
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	for i := 0; i < 5; i++ {
		group := domain.Group{
			ID:    uuid.NewString(),
			CustomerID: "test-customer-integration",
			Name:  fmt.Sprintf("Group %d", i),
		}
		_ = stores.Groups.Put(ctx, group)
	}

	// List groups
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups?page=1&pageSize=20", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var listResult domain.ListResult[domain.Group]
	if err := json.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}

	if listResult.Pagination.TotalCount < 5 {
		t.Errorf("Expected at least 5 groups, got %d", listResult.Pagination.TotalCount)
	}
}

func TestIntegration_Group_List_WithSorting(t *testing.T) {
	app, cleanup := setupGroupIntegrationApp(t)
	defer cleanup()

	// Create groups
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	store := stores.Groups
	store.Put(ctx, domain.Group{
		ID:    uuid.NewString(),
		CustomerID: "test-customer-integration",
		Name:  "B Group",
	})
	store.Put(ctx, domain.Group{
		ID:    uuid.NewString(),
		CustomerID: "test-customer-integration",
		Name:  "A Group",
	})

	// List with sorting
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups?sortBy=name:asc", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
