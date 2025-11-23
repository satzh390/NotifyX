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

	mongoadapter "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/core/domain"
)

func TestIntegration_Rule_Create(t *testing.T) {
	app, cleanup := setupRuleIntegrationApp(t)
	defer cleanup()

	createBody := map[string]interface{}{
		"eventType": "order.created",
		"channels":  []string{"email", "sms"},
		"defaultRecipients": map[string]interface{}{
			"subscriberIds": []string{"sub-1"},
		},
	}
	bodyJSON, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rules", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Create request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var created domain.Rule
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	if created.EventType != "order.created" {
		t.Errorf("Expected eventType order.created, got %s", created.EventType)
	}
	if created.OrgID != "test-org-integration" {
		t.Errorf("Expected orgID test-org-integration, got %s", created.OrgID)
	}
}

func TestIntegration_Rule_Get(t *testing.T) {
	app, cleanup := setupRuleIntegrationApp(t)
	defer cleanup()

	// Create a rule first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	eventType := "order.created"
	rule := domain.Rule{
		EventType: eventType,
		OrgID:     "test-org-integration",
		Channels:  []domain.ChannelType{domain.ChannelEmail},
		CreatedAt: time.Now(),
	}
	if err := stores.Rules.Put(ctx, rule); err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Get the rule
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules/"+eventType, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var retrieved domain.Rule
	if err := json.NewDecoder(resp.Body).Decode(&retrieved); err != nil {
		t.Fatalf("Failed to decode get response: %v", err)
	}

	if retrieved.EventType != eventType {
		t.Errorf("Expected eventType %s, got %s", eventType, retrieved.EventType)
	}
}

func TestIntegration_Rule_Get_NotFound(t *testing.T) {
	app, cleanup := setupRuleIntegrationApp(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules/non-existent", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestIntegration_Rule_Update(t *testing.T) {
	app, cleanup := setupRuleIntegrationApp(t)
	defer cleanup()

	// Create a rule first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	eventType := "order.created"
	rule := domain.Rule{
		EventType: eventType,
		OrgID:     "test-org-integration",
		Channels:  []domain.ChannelType{domain.ChannelEmail},
		CreatedAt: time.Now(),
	}
	if err := stores.Rules.Put(ctx, rule); err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Update the rule
	updateBody := map[string]interface{}{
		"eventType": eventType,
		"channels":  []string{"email", "sms"},
	}
	updateJSON, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/rules/"+eventType, bytes.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Update request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_Rule_Patch(t *testing.T) {
	app, cleanup := setupRuleIntegrationApp(t)
	defer cleanup()

	// Create a rule first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	eventType := "order.created"
	rule := domain.Rule{
		EventType: eventType,
		OrgID:     "test-org-integration",
		Channels:  []domain.ChannelType{domain.ChannelEmail},
		CreatedAt: time.Now(),
	}
	if err := stores.Rules.Put(ctx, rule); err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Patch the rule
	patchBody := map[string]interface{}{
		"channels": []string{"email", "sms"},
	}
	patchJSON, _ := json.Marshal(patchBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/rules/"+eventType, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Patch request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var patched domain.Rule
	if err := json.NewDecoder(resp.Body).Decode(&patched); err != nil {
		t.Fatalf("Failed to decode patch response: %v", err)
	}

	if len(patched.Channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(patched.Channels))
	}
}

func TestIntegration_Rule_Patch_NotFound(t *testing.T) {
	app, cleanup := setupRuleIntegrationApp(t)
	defer cleanup()

	patchBody := map[string]interface{}{
		"channels": []string{"email", "sms"},
	}
	patchJSON, _ := json.Marshal(patchBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/rules/non-existent", bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Patch request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestIntegration_Rule_Delete(t *testing.T) {
	app, cleanup := setupRuleIntegrationApp(t)
	defer cleanup()

	// Create a rule first
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	eventType := "order.created"
	rule := domain.Rule{
		EventType: eventType,
		OrgID:     "test-org-integration",
		CreatedAt: time.Now(),
	}
	if err := stores.Rules.Put(ctx, rule); err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Delete the rule
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/rules/"+eventType, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Delete request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify deleted
	req = httptest.NewRequest(http.MethodGet, "/api/v1/rules/"+eventType, nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Get after delete request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 after delete, got %d", resp.StatusCode)
	}
}

func TestIntegration_Rule_List(t *testing.T) {
	app, cleanup := setupRuleIntegrationApp(t)
	defer cleanup()

	// Create multiple rules
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	store := stores.Rules
	store.Put(ctx, domain.Rule{
		EventType: "order.created",
		OrgID:     "test-org-integration",
		CreatedAt: time.Now(),
	})
	store.Put(ctx, domain.Rule{
		EventType: "order.updated",
		OrgID:     "test-org-integration",
		CreatedAt: time.Now(),
	})

	// List rules
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules?page=1&pageSize=20", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var listResult domain.ListResult[domain.Rule]
	if err := json.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}

	if listResult.Pagination.TotalCount < 2 {
		t.Errorf("Expected at least 2 rules, got %d", listResult.Pagination.TotalCount)
	}
}

func TestIntegration_Rule_List_WithSorting(t *testing.T) {
	app, cleanup := setupRuleIntegrationApp(t)
	defer cleanup()

	// Create rules
	ctx := context.Background()
	stores, _, _ := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})

	store := stores.Rules
	store.Put(ctx, domain.Rule{
		EventType: "order.created",
		OrgID:     "test-org-integration",
		CreatedAt: time.Now(),
	})
	store.Put(ctx, domain.Rule{
		EventType: "order.updated",
		OrgID:     "test-org-integration",
		CreatedAt: time.Now().Add(time.Second),
	})

	// List with sorting
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules?sortBy=createdAt:desc", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
