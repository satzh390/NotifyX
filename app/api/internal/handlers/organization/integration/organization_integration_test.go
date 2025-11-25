//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	mongoadapter "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/core/domain"
)

func TestIntegration_Organization_Create(t *testing.T) {
	app, cleanup := setupOrganizationIntegrationApp(t)
	defer cleanup()

	createBody := map[string]interface{}{
		"name": "Integration Test Organization",
		"type": "company",
	}
	bodyJSON, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Create request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var created domain.Organization
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	if created.Name != "Integration Test Organization" {
		t.Errorf("Expected name Integration Test Organization, got %s", created.Name)
	}
	if created.Type != domain.OrganizationTypeCompany {
		t.Errorf("Expected type company, got %s", created.Type)
	}
	if created.ID == "" {
		t.Error("Expected ID to be set")
	}
}

func TestIntegration_Organization_Get(t *testing.T) {
	app, cleanup := setupOrganizationIntegrationApp(t)
	defer cleanup()

	// Create an organization first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	orgID := uuid.NewString()
	org := domain.Organization{
		ID:        orgID,
		Name:      "Test Organization",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Organizations.Put(ctx, org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	// Get the organization
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+orgID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var retrieved domain.Organization
	if err := json.NewDecoder(resp.Body).Decode(&retrieved); err != nil {
		t.Fatalf("Failed to decode get response: %v", err)
	}

	if retrieved.ID != orgID {
		t.Errorf("Expected ID %s, got %s", orgID, retrieved.ID)
	}
	if retrieved.Name != "Test Organization" {
		t.Errorf("Expected name Test Organization, got %s", retrieved.Name)
	}
}

func TestIntegration_Organization_Get_NotFound(t *testing.T) {
	app, cleanup := setupOrganizationIntegrationApp(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/non-existent", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestIntegration_Organization_Put(t *testing.T) {
	app, cleanup := setupOrganizationIntegrationApp(t)
	defer cleanup()

	// Create an organization first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	orgID := uuid.NewString()
	org := domain.Organization{
		ID:        orgID,
		Name:      "Old Name",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Organizations.Put(ctx, org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	// Update the organization
	updateBody := map[string]interface{}{
		"id":   orgID,
		"name": "New Name",
		"type": "saasProvider",
	}
	bodyJSON, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/organizations/"+orgID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Put request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Organization
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode put response: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name New Name, got %s", updated.Name)
	}
	if updated.Type != domain.OrganizationTypeSaaSProvider {
		t.Errorf("Expected type saasProvider, got %s", updated.Type)
	}
}

func TestIntegration_Organization_Patch(t *testing.T) {
	app, cleanup := setupOrganizationIntegrationApp(t)
	defer cleanup()

	// Create an organization first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	orgID := uuid.NewString()
	org := domain.Organization{
		ID:        orgID,
		Name:      "Old Name",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Organizations.Put(ctx, org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	// Partially update the organization
	patchBody := map[string]interface{}{
		"name": "Patched Name",
	}
	bodyJSON, _ := json.Marshal(patchBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/organizations/"+orgID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Patch request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Organization
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode patch response: %v", err)
	}

	if updated.Name != "Patched Name" {
		t.Errorf("Expected name Patched Name, got %s", updated.Name)
	}
	if updated.Type != domain.OrganizationTypeCompany {
		t.Errorf("Expected type to remain company, got %s", updated.Type)
	}
}

func TestIntegration_Organization_Delete(t *testing.T) {
	app, cleanup := setupOrganizationIntegrationApp(t)
	defer cleanup()

	// Create an organization first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	orgID := uuid.NewString()
	org := domain.Organization{
		ID:        orgID,
		Name:      "To Be Deleted",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Organizations.Put(ctx, org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	// Delete the organization
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/"+orgID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Delete request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify it's deleted
	_, err = stores.Organizations.Get(ctx, orgID)
	if err == nil {
		t.Error("Expected organization to be deleted, but it still exists")
	}
}

func TestIntegration_Organization_List(t *testing.T) {
	app, cleanup := setupOrganizationIntegrationApp(t)
	defer cleanup()

	// Create some organizations
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	org1 := domain.Organization{
		ID:        uuid.NewString(),
		Name:      "Organization 1",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	org2 := domain.Organization{
		ID:        uuid.NewString(),
		Name:      "Organization 2",
		Type:      domain.OrganizationTypeSaaSProvider,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := stores.Organizations.Put(ctx, org1); err != nil {
		t.Fatalf("Failed to create organization 1: %v", err)
	}
	if err := stores.Organizations.Put(ctx, org2); err != nil {
		t.Fatalf("Failed to create organization 2: %v", err)
	}

	// List organizations
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations?page=0&pageSize=10", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result domain.ListResult[domain.Organization]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}

	if result.Pagination.TotalCount < 2 {
		t.Errorf("Expected at least 2 organizations, got %d", result.Pagination.TotalCount)
	}
}

