package template

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTemplateTestApp() (*fiber.App, *MockTemplateStore) {
	app := fiber.New()
	store := new(MockTemplateStore)
	handler := NewTemplateHandler(store)

	api := app.Group("/api/v1")
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("customerId", "test-customer")
		return c.Next()
	})

	templates := api.Group("/templates")
	templates.Post("", handler.Create)
	templates.Get("/:id", handler.Get)
	templates.Put("/:id", handler.Put)
	templates.Patch("/:id", handler.Patch)
	templates.Delete("/:id", handler.Delete)

	return app, store
}

func TestTemplateHandler_Create(t *testing.T) {
	app, store := setupTemplateTestApp()

	store.On("Put", mock.Anything, mock.MatchedBy(func(tpl domain.Template) bool {
		return tpl.Name == "Order Confirmation" && tpl.CustomerID == "test-customer"
	})).Return(nil).Once()

	body := map[string]interface{}{
		"name":    "Order Confirmation",
		"channel": "email",
		"content": map[string]interface{}{
			"subject": "Order Confirmation",
			"body":    "Your order has been confirmed",
		},
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var template domain.Template
	err = json.NewDecoder(resp.Body).Decode(&template)
	assert.NoError(t, err)
	assert.Equal(t, "Order Confirmation", template.Name)
	assert.Equal(t, "test-customer", template.CustomerID)
	assert.Equal(t, "Your order has been confirmed", template.Content.Body)

	store.AssertExpectations(t)
}

func TestTemplateHandler_Create_ValidationError(t *testing.T) {
	app, _ := setupTemplateTestApp()

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "missing name",
			body: map[string]interface{}{
				"channel": "email",
				"content": map[string]interface{}{
					"body": "Test body",
				},
			},
		},
		{
			name: "missing channel",
			body: map[string]interface{}{
				"name": "Test Template",
				"content": map[string]interface{}{
					"body": "Test body",
				},
			},
		},
		{
			name: "missing content",
			body: map[string]interface{}{
				"name":    "Test Template",
				"channel": "email",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestTemplateHandler_Get(t *testing.T) {
	app, store := setupTemplateTestApp()

	templateID := uuid.NewString()
	expectedTemplate := domain.Template{
		ID:      templateID,
		CustomerID: "test-customer",
		Name:    "Test Template",
		Channel: domain.ChannelEmail,
		Content: domain.TemplateContent{
			Body:    "Test body",
			Subject: "Test subject",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	store.On("Get", mock.Anything, "test-org", templateID).Return(expectedTemplate, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var template domain.Template
	err = json.NewDecoder(resp.Body).Decode(&template)
	assert.NoError(t, err)
	assert.Equal(t, templateID, template.ID)
	assert.Equal(t, "Test body", template.Content.Body)

	store.AssertExpectations(t)
}

func TestTemplateHandler_Get_NotFound(t *testing.T) {
	app, store := setupTemplateTestApp()

	store.On("Get", mock.Anything, "test-org", "non-existent").Return(domain.Template{}, storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestTemplateHandler_Put(t *testing.T) {
	app, store := setupTemplateTestApp()

	templateID := uuid.NewString()
	existingTemplate := domain.Template{
		ID:      templateID,
		CustomerID: "test-customer",
		Name:    "Old Name",
		Channel: domain.ChannelEmail,
		Content: domain.TemplateContent{
			Body: "Old body",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test PUT with existing template (update)
	store.On("Get", mock.Anything, "test-org", templateID).Return(existingTemplate, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(tpl domain.Template) bool {
		return tpl.ID == templateID && tpl.Name == "New Name"
	})).Return(nil).Once()
	store.On("Get", mock.Anything, "test-org", templateID).Return(domain.Template{
		ID:      templateID,
		CustomerID: "test-customer",
		Name:    "New Name",
		Channel: domain.ChannelEmail,
		Content: domain.TemplateContent{
			Body: "New body",
		},
		CreatedAt: existingTemplate.CreatedAt,
		UpdatedAt: time.Now(),
	}, nil).Once()

	fullBody := map[string]interface{}{
		"id":      templateID,
		"name":    "New Name",
		"channel": "email",
		"content": map[string]interface{}{
			"body": "New body",
		},
	}
	bodyJSON, _ := json.Marshal(fullBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+templateID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated domain.Template
	err = json.NewDecoder(resp.Body).Decode(&updated)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)

	store.AssertExpectations(t)
}

func TestTemplateHandler_Put_Create(t *testing.T) {
	app, store := setupTemplateTestApp()

	templateID := uuid.NewString()

	// Test PUT with non-existing template (create)
	store.On("Get", mock.Anything, "test-org", templateID).Return(domain.Template{}, storage.ErrNotFound).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(tpl domain.Template) bool {
		return tpl.ID == templateID && tpl.Name == "New Template"
	})).Return(nil).Once()
	store.On("Get", mock.Anything, "test-org", templateID).Return(domain.Template{
		ID:      templateID,
		CustomerID: "test-customer",
		Name:    "New Template",
		Channel: domain.ChannelEmail,
		Content: domain.TemplateContent{
			Body: "New body",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil).Once()

	fullBody := map[string]interface{}{
		"id":      templateID,
		"name":    "New Template",
		"channel": "email",
		"content": map[string]interface{}{
			"body": "New body",
		},
	}
	bodyJSON, _ := json.Marshal(fullBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+templateID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestTemplateHandler_Patch(t *testing.T) {
	app, store := setupTemplateTestApp()

	templateID := uuid.NewString()
	existingTemplate := domain.Template{
		ID:      templateID,
		CustomerID: "test-customer",
		Name:    "Old Name",
		Channel: domain.ChannelEmail,
		Content: domain.TemplateContent{
			Body: "Old body",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	store.On("Get", mock.Anything, "test-org", templateID).Return(existingTemplate, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(tpl domain.Template) bool {
		return tpl.ID == templateID && tpl.Name == "New Name"
	})).Return(nil).Once()

	patch := map[string]interface{}{
		"name": "New Name",
		"content": map[string]interface{}{
			"body": "New body",
		},
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/templates/"+templateID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated domain.Template
	err = json.NewDecoder(resp.Body).Decode(&updated)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)

	store.AssertExpectations(t)
}

func TestTemplateHandler_Patch_NotFound(t *testing.T) {
	app, store := setupTemplateTestApp()

	templateID := uuid.NewString()

	store.On("Get", mock.Anything, "test-org", templateID).Return(domain.Template{}, storage.ErrNotFound).Once()

	patch := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/templates/"+templateID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestTemplateHandler_Delete(t *testing.T) {
	app, store := setupTemplateTestApp()

	templateID := uuid.NewString()

	store.On("Delete", mock.Anything, "test-org", templateID).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/"+templateID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	store.AssertExpectations(t)
}
