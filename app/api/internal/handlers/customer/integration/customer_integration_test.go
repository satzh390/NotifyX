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

func TestIntegration_Customer_Create(t *testing.T) {
	app, cleanup := setupCustomerIntegrationApp(t)
	defer cleanup()

	orgID := uuid.NewString()
	createBody := map[string]interface{}{
		"orgId": orgID,
		"name":  "Integration Test Customer",
		"logo":  "https://example.com/logo.png",
	}
	bodyJSON, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Create request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var created domain.Customer
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	if created.Name != "Integration Test Customer" {
		t.Errorf("Expected name Integration Test Customer, got %s", created.Name)
	}
	if created.OrgID != orgID {
		t.Errorf("Expected orgID %s, got %s", orgID, created.OrgID)
	}
	if created.Logo != "https://example.com/logo.png" {
		t.Errorf("Expected logo https://example.com/logo.png, got %s", created.Logo)
	}
	if created.ID == "" {
		t.Error("Expected ID to be set")
	}
}

func TestIntegration_Customer_Get(t *testing.T) {
	app, cleanup := setupCustomerIntegrationApp(t)
	defer cleanup()

	// Create a customer first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	customerID := uuid.NewString()
	orgID := uuid.NewString()
	customer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "Test Customer",
		Logo:      "https://example.com/logo.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Customers.Put(ctx, customer); err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	// Get the customer
	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers/"+customerID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var retrieved domain.Customer
	if err := json.NewDecoder(resp.Body).Decode(&retrieved); err != nil {
		t.Fatalf("Failed to decode get response: %v", err)
	}

	if retrieved.ID != customerID {
		t.Errorf("Expected ID %s, got %s", customerID, retrieved.ID)
	}
	if retrieved.Name != "Test Customer" {
		t.Errorf("Expected name Test Customer, got %s", retrieved.Name)
	}
}

func TestIntegration_Customer_Get_NotFound(t *testing.T) {
	app, cleanup := setupCustomerIntegrationApp(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers/non-existent", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestIntegration_Customer_Put(t *testing.T) {
	app, cleanup := setupCustomerIntegrationApp(t)
	defer cleanup()

	// Create a customer first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	customerID := uuid.NewString()
	orgID := uuid.NewString()
	customer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "Old Name",
		Logo:      "https://example.com/old-logo.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Customers.Put(ctx, customer); err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	// Update the customer
	updateBody := map[string]interface{}{
		"id":    customerID,
		"orgId": orgID,
		"name":  "New Name",
		"logo":  "https://example.com/new-logo.png",
	}
	bodyJSON, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/customers/"+customerID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Put request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Customer
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode put response: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name New Name, got %s", updated.Name)
	}
	if updated.Logo != "https://example.com/new-logo.png" {
		t.Errorf("Expected logo https://example.com/new-logo.png, got %s", updated.Logo)
	}
}

func TestIntegration_Customer_Patch(t *testing.T) {
	app, cleanup := setupCustomerIntegrationApp(t)
	defer cleanup()

	// Create a customer first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	customerID := uuid.NewString()
	orgID := uuid.NewString()
	customer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "Old Name",
		Logo:      "https://example.com/old-logo.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Customers.Put(ctx, customer); err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	// Partially update the customer
	patchBody := map[string]interface{}{
		"name": "Patched Name",
	}
	bodyJSON, _ := json.Marshal(patchBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/customers/"+customerID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Patch request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Customer
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode patch response: %v", err)
	}

	if updated.Name != "Patched Name" {
		t.Errorf("Expected name Patched Name, got %s", updated.Name)
	}
	if updated.Logo != "https://example.com/old-logo.png" {
		t.Errorf("Expected logo to remain unchanged, got %s", updated.Logo)
	}
}

func TestIntegration_Customer_Delete(t *testing.T) {
	app, cleanup := setupCustomerIntegrationApp(t)
	defer cleanup()

	// Create a customer first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	customerID := uuid.NewString()
	orgID := uuid.NewString()
	customer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "To Be Deleted",
		Logo:      "https://example.com/logo.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Customers.Put(ctx, customer); err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	// Delete the customer
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/customers/"+customerID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Delete request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify it's deleted
	_, err = stores.Customers.Get(ctx, customerID)
	if err == nil {
		t.Error("Expected customer to be deleted, but it still exists")
	}
}

func TestIntegration_Customer_List(t *testing.T) {
	app, cleanup := setupCustomerIntegrationApp(t)
	defer cleanup()

	// Create some customers
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	orgID := uuid.NewString()
	customer1 := domain.Customer{
		ID:        uuid.NewString(),
		OrgID:     orgID,
		Name:      "Customer 1",
		Logo:      "https://example.com/logo1.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	customer2 := domain.Customer{
		ID:        uuid.NewString(),
		OrgID:     orgID,
		Name:      "Customer 2",
		Logo:      "https://example.com/logo2.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := stores.Customers.Put(ctx, customer1); err != nil {
		t.Fatalf("Failed to create customer 1: %v", err)
	}
	if err := stores.Customers.Put(ctx, customer2); err != nil {
		t.Fatalf("Failed to create customer 2: %v", err)
	}

	// List customers
	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers?page=0&pageSize=10", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result domain.ListResult[domain.Customer]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}

	if result.Pagination.TotalCount < 2 {
		t.Errorf("Expected at least 2 customers, got %d", result.Pagination.TotalCount)
	}
}

