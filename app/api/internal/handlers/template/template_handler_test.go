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

func setupTemplateTestApp() (*fiber.App, *mockTemplateStore) {
	app := fiber.New()
	store := &mockTemplateStore{templates: make(map[string]domain.Template)}
	handler := NewTemplateHandler(store)

	api := app.Group("/api/v1")
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("orgId", "test-org")
		return c.Next()
	})

	templates := api.Group("/templates")
	templates.Post("", handler.Create)
	templates.Get("/:id", handler.Get)
	templates.Put("/:id", handler.Update)
	templates.Delete("/:id", handler.Delete)

	return app, store
}

func TestTemplateHandler_Create(t *testing.T) {
	app, _ := setupTemplateTestApp()

	body := map[string]interface{}{
		"name":    "Order Confirmation",
		"channel": "email",
		"body":    "Your order has been confirmed",
		"subject": "Order Confirmation",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var template domain.Template
	if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if template.Name != "Order Confirmation" {
		t.Errorf("Expected name Order Confirmation, got %s", template.Name)
	}
	if template.OrgID != "test-org" {
		t.Errorf("Expected orgID test-org, got %s", template.OrgID)
	}
}

func TestTemplateHandler_Get(t *testing.T) {
	app, store := setupTemplateTestApp()

	templateID := uuid.NewString()
	store.Put(context.Background(), domain.Template{
		ID:        templateID,
		OrgID:     "test-org",
		Name:      "Test Template",
		Channel:   domain.ChannelEmail,
		Body:      "Test body",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID, nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var template domain.Template
	if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if template.ID != templateID {
		t.Errorf("Expected ID %s, got %s", templateID, template.ID)
	}
}

func TestTemplateHandler_Get_NotFound(t *testing.T) {
	app, _ := setupTemplateTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/non-existent", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestTemplateHandler_Update(t *testing.T) {
	app, store := setupTemplateTestApp()

	templateID := uuid.NewString()
	store.Put(context.Background(), domain.Template{
		ID:        templateID,
		OrgID:     "test-org",
		Name:      "Old Name",
		Channel:   domain.ChannelEmail,
		Body:      "Old body",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	patch := map[string]interface{}{
		"name": "New Name",
		"body": "New body",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+templateID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var updated domain.Template
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name New Name, got %s", updated.Name)
	}
}

func TestTemplateHandler_Delete(t *testing.T) {
	app, store := setupTemplateTestApp()

	templateID := uuid.NewString()
	store.Put(context.Background(), domain.Template{
		ID:        templateID,
		OrgID:     "test-org",
		Name:      "Test Template",
		Channel:   domain.ChannelEmail,
		Body:      "Test body",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/"+templateID, nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify deleted
	_, err = store.Get(context.Background(), "test-org", templateID)
	if err == nil {
		t.Error("Expected template to be deleted")
	}
}

