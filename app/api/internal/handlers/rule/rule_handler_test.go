package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/core/domain"
)

func setupRuleTestApp() (*fiber.App, *mockRuleStore) {
	app := fiber.New()
	store := &mockRuleStore{rules: make(map[string]domain.Rule)}
	handler := NewRuleHandler(store)

	api := app.Group("/api/v1")
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("orgId", "test-org")
		return c.Next()
	})

	rules := api.Group("/rules")
	rules.Post("", handler.Create)
	rules.Get("/:eventType", handler.Get)
	rules.Put("/:eventType", handler.Update)
	rules.Delete("/:eventType", handler.Delete)
	rules.Get("", handler.List)

	return app, store
}

func TestRuleHandler_Create(t *testing.T) {
	app, _ := setupRuleTestApp()

	body := map[string]interface{}{
		"eventType": "order.created",
		"channels":  []string{"email", "sms"},
		"defaultRecipients": map[string]interface{}{
			"subscriberIds": []string{"sub-1"},
		},
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rules", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var rule domain.Rule
	if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if rule.EventType != "order.created" {
		t.Errorf("Expected eventType order.created, got %s", rule.EventType)
	}
	if rule.OrgID != "test-org" {
		t.Errorf("Expected orgID test-org, got %s", rule.OrgID)
	}
}

func TestRuleHandler_Get(t *testing.T) {
	app, store := setupRuleTestApp()

	eventType := "order.created"
	store.Put(context.Background(), domain.Rule{
		EventType: eventType,
		OrgID:     "test-org",
		Channels:  []domain.ChannelType{domain.ChannelEmail},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules/"+eventType, nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var rule domain.Rule
	if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if rule.EventType != eventType {
		t.Errorf("Expected eventType %s, got %s", eventType, rule.EventType)
	}
}

func TestRuleHandler_Get_NotFound(t *testing.T) {
	app, _ := setupRuleTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules/non-existent", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestRuleHandler_Update(t *testing.T) {
	app, store := setupRuleTestApp()

	eventType := "order.created"
	store.Put(context.Background(), domain.Rule{
		EventType: eventType,
		OrgID:     "test-org",
		Channels:  []domain.ChannelType{domain.ChannelEmail},
	})

	patch := map[string]interface{}{
		"channels": []string{"email", "sms"},
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/rules/"+eventType, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRuleHandler_Delete(t *testing.T) {
	app, store := setupRuleTestApp()

	eventType := "order.created"
	store.Put(context.Background(), domain.Rule{
		EventType: eventType,
		OrgID:     "test-org",
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/rules/"+eventType, nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify deleted
	_, err = store.Get(context.Background(), "test-org", eventType)
	if err == nil {
		t.Error("Expected rule to be deleted")
	}
}

func TestRuleHandler_List(t *testing.T) {
	app, store := setupRuleTestApp()

	store.Put(context.Background(), domain.Rule{
		EventType: "order.created",
		OrgID:     "test-org",
	})
	store.Put(context.Background(), domain.Rule{
		EventType: "order.updated",
		OrgID:     "test-org",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules?page=1&pageSize=20", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result domain.ListResult[domain.Rule]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Pagination.TotalCount < 2 {
		t.Errorf("Expected at least 2 rules, got %d", result.Pagination.TotalCount)
	}
}

