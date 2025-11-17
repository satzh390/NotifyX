//go:build integration
// +build integration

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
	s3adapter "github.com/notifyx/core/adapters/s3"
)

func setupTemplateIntegrationApp(t *testing.T) (*fiber.App, func()) {
	// For templates, we need S3. In integration tests, we'll use LocalStack
	ctx := context.Background()
	
	// Note: This requires LocalStack to be running with S3
	// For now, we'll skip if S3 is not available
	templateRepo, err := s3adapter.NewTemplateRepository(ctx, s3adapter.Options{
		Bucket:      "notifyx-templates",
		Region:      "us-east-1",
		Endpoint:    "http://localhost:4566", // LocalStack endpoint
		KeyPrefix:   "",
		AccessKeyID: "test",
		SecretKey:   "test",
	})
	if err != nil {
		t.Skipf("S3 not available for integration test: %v", err)
	}

	app := fiber.New()
	api := app.Group("/api/v1")
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("orgId", "test-org-integration")
		return c.Next()
	})

	templateHandler := NewTemplateHandler(templateRepo)
	templates := api.Group("/templates")
	templates.Post("", templateHandler.Create)
	templates.Get("/:id", templateHandler.Get)
	templates.Put("/:id", templateHandler.Update)
	templates.Delete("/:id", templateHandler.Delete)

	return app, func() {}
}

func TestIntegration_Template_Create(t *testing.T) {
	app, cleanup := setupTemplateIntegrationApp(t)
	defer cleanup()

	createBody := map[string]interface{}{
		"name":    "Order Confirmation",
		"channel": "email",
		"body":    "Your order has been confirmed",
		"subject": "Order Confirmation",
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
	ctx := context.Background()
	templateRepo, err := s3adapter.NewTemplateRepository(ctx, s3adapter.Options{
		Bucket:      "notifyx-templates",
		Region:      "us-east-1",
		Endpoint:    "http://localhost:4566",
		KeyPrefix:   "",
		AccessKeyID: "test",
		SecretKey:   "test",
	})
	if err != nil {
		t.Skipf("S3 not available: %v", err)
	}

	templateID := uuid.NewString()
	template := domain.Template{
		ID:        templateID,
		OrgID:     "test-org-integration",
		Name:      "Test Template",
		Channel:   domain.ChannelEmail,
		Body:      "Test body",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := templateRepo.Put(ctx, template); err != nil {
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
	ctx := context.Background()
	templateRepo, err := s3adapter.NewTemplateRepository(ctx, s3adapter.Options{
		Bucket:      "notifyx-templates",
		Region:      "us-east-1",
		Endpoint:    "http://localhost:4566",
		KeyPrefix:   "",
		AccessKeyID: "test",
		SecretKey:   "test",
	})
	if err != nil {
		t.Skipf("S3 not available: %v", err)
	}

	templateID := uuid.NewString()
	template := domain.Template{
		ID:        templateID,
		OrgID:     "test-org-integration",
		Name:      "Old Name",
		Channel:   domain.ChannelEmail,
		Body:      "Old body",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := templateRepo.Put(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Update the template
	updateBody := map[string]interface{}{
		"name": "New Name",
		"body": "New body",
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

func TestIntegration_Template_Delete(t *testing.T) {
	app, cleanup := setupTemplateIntegrationApp(t)
	defer cleanup()

	// Create a template first
	ctx := context.Background()
	templateRepo, err := s3adapter.NewTemplateRepository(ctx, s3adapter.Options{
		Bucket:      "notifyx-templates",
		Region:      "us-east-1",
		Endpoint:    "http://localhost:4566",
		KeyPrefix:   "",
		AccessKeyID: "test",
		SecretKey:   "test",
	})
	if err != nil {
		t.Skipf("S3 not available: %v", err)
	}

	templateID := uuid.NewString()
	template := domain.Template{
		ID:        templateID,
		OrgID:     "test-org-integration",
		Name:      "Test Template",
		Channel:   domain.ChannelEmail,
		Body:      "Test body",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := templateRepo.Put(ctx, template); err != nil {
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
	_, err = templateRepo.Get(ctx, "test-org-integration", templateID)
	if err == nil {
		t.Error("Expected template to be deleted")
	}
}

