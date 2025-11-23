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
	"github.com/notifyx/core/domain"
)

func TestIntegration_Template_Create(t *testing.T) {
	app, cleanup := setupTemplateIntegrationApp(t)
	defer cleanup()

	createBody := map[string]interface{}{
		"name":    "Order Confirmation",
		"channel": "email",
		"content": map[string]interface{}{
			"body":    "Your order has been confirmed",
			"subject": "Order Confirmation",
		},
	}
	bodyJSON, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Create request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var created domain.Template
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	if created.Name != "Order Confirmation" {
		t.Errorf("Expected name Order Confirmation, got %s", created.Name)
	}
	if created.OrgID != "test-org-integration" {
		t.Errorf("Expected orgID test-org-integration, got %s", created.OrgID)
	}
}

func TestIntegration_Template_Get(t *testing.T) {
	app, cleanup := setupTemplateIntegrationApp(t)
	defer cleanup()

	// Create a template first
	stores, _, _ := setupIntegrationTest(t)
	ctx := context.Background()

	templateID := uuid.NewString()
	template := domain.Template{
		ID:      templateID,
		OrgID:   "test-org-integration",
		Name:    "Test Template",
		Channel: domain.ChannelEmail,
		Content: domain.TemplateContent{
			Body: "Test body",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Templates.Put(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Get the template
	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var retrieved domain.Template
	if err := json.NewDecoder(resp.Body).Decode(&retrieved); err != nil {
		t.Fatalf("Failed to decode get response: %v", err)
	}

	if retrieved.ID != templateID {
		t.Errorf("Expected ID %s, got %s", templateID, retrieved.ID)
	}
}

func TestIntegration_Template_Update(t *testing.T) {
	app, cleanup := setupTemplateIntegrationApp(t)
	defer cleanup()

	// Create a template first
	stores, _, _ := setupIntegrationTest(t)
	ctx := context.Background()

	templateID := uuid.NewString()
	template := domain.Template{
		ID:      templateID,
		OrgID:   "test-org-integration",
		Name:    "Old Name",
		Channel: domain.ChannelEmail,
		Content: domain.TemplateContent{
			Body: "Old body",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Templates.Put(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Update the template
	updateBody := map[string]interface{}{
		"name": "New Name",
		"content": map[string]interface{}{
			"body": "New body",
		},
	}
	updateJSON, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+templateID, bytes.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Update request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Template
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode update response: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name New Name, got %s", updated.Name)
	}
}

func TestIntegration_Template_Patch(t *testing.T) {
	app, cleanup := setupTemplateIntegrationApp(t)
	defer cleanup()

	// Create a template first
	stores, _, _ := setupIntegrationTest(t)
	ctx := context.Background()

	templateID := uuid.NewString()
	template := domain.Template{
		ID:      templateID,
		OrgID:   "test-org-integration",
		Name:    "Old Name",
		Channel: domain.ChannelEmail,
		Content: domain.TemplateContent{
			Body:    "Old body",
			Subject: "Old subject",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Templates.Put(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Patch the template
	patchBody := map[string]interface{}{
		"name": "New Name",
		"content": map[string]interface{}{
			"body": "New body",
		},
	}
	patchJSON, _ := json.Marshal(patchBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/templates/"+templateID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Patch request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var patched domain.Template
	if err := json.NewDecoder(resp.Body).Decode(&patched); err != nil {
		t.Fatalf("Failed to decode patch response: %v", err)
	}

	if patched.Name != "New Name" {
		t.Errorf("Expected name New Name, got %s", patched.Name)
	}
	if patched.Content.Body != "New body" {
		t.Errorf("Expected body New body, got %s", patched.Content.Body)
	}
}

func TestIntegration_Template_Patch_NotFound(t *testing.T) {
	app, cleanup := setupTemplateIntegrationApp(t)
	defer cleanup()

	patchBody := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patchBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/templates/non-existent", bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Patch request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestIntegration_Template_Delete(t *testing.T) {
	app, cleanup := setupTemplateIntegrationApp(t)
	defer cleanup()

	// Create a template first
	stores, _, _ := setupIntegrationTest(t)
	ctx := context.Background()

	templateID := uuid.NewString()
	template := domain.Template{
		ID:      templateID,
		OrgID:   "test-org-integration",
		Name:    "Test Template",
		Channel: domain.ChannelEmail,
		Content: domain.TemplateContent{
			Body: "Test body",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := stores.Templates.Put(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Delete the template
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/"+templateID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Delete request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify deleted
	_, err = stores.Templates.Get(ctx, "test-org-integration", templateID)
	if err == nil {
		t.Error("Expected template to be deleted")
	}
}
