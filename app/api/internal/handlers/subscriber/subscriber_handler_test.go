package subscriber

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
	"github.com/notifyx/core/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupSubscriberTestApp() (*fiber.App, *MockSubscriberStore) {
	app := fiber.New()
	store := new(MockSubscriberStore)
	handler := NewSubscriberHandler(store)

	api := app.Group("/api/v1")
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("orgId", "test-org")
		return c.Next()
	})

	subscribers := api.Group("/subscribers")
	subscribers.Post("", handler.Create)
	subscribers.Get("/:id", handler.Get)
	subscribers.Put("/:id", handler.Put)
	subscribers.Patch("/:id", handler.Patch)
	subscribers.Delete("/:id", handler.Delete)
	subscribers.Get("", handler.List)

	return app, store
}

func TestSubscriberHandler_Create(t *testing.T) {
	app, store := setupSubscriberTestApp()

	store.On("Put", mock.Anything, mock.MatchedBy(func(s domain.Subscriber) bool {
		return s.Email == "test@example.com" && s.OrgID == "test-org"
	})).Return(nil).Once()

	// Mock Get to return the created subscriber (handler calls Get after Put)
	store.On("Get", mock.Anything, "test-org", mock.AnythingOfType("string")).Return(func(ctx context.Context, orgID string, id string) domain.Subscriber {
		return domain.Subscriber{
			ID:    id,
			OrgID: "test-org",
			Email: "test@example.com",
		}
	}, nil).Once()

	body := map[string]interface{}{
		"email": "test@example.com",
		"preferences": map[string]interface{}{
			"channels": map[string]bool{"email": true},
			"language": "en",
		},
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscribers", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var subscriber domain.Subscriber
	err = json.NewDecoder(resp.Body).Decode(&subscriber)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", subscriber.Email)
	assert.Equal(t, "test-org", subscriber.OrgID)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_Create_ValidationError(t *testing.T) {
	app, _ := setupSubscriberTestApp()

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "invalid email format",
			body: map[string]interface{}{
				"email": "invalid-email",
			},
		},
		{
			name: "invalid webhook URL",
			body: map[string]interface{}{
				"email":      "test@example.com",
				"webhookUrl": "not-a-valid-url",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/subscribers", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestSubscriberHandler_Get(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()
	expectedSubscriber := domain.Subscriber{
		ID:    subID,
		OrgID: "test-org",
		Email: "test@example.com",
	}

	store.On("Get", mock.Anything, "test-org", subID).Return(expectedSubscriber, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers/"+subID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var subscriber domain.Subscriber
	err = json.NewDecoder(resp.Body).Decode(&subscriber)
	assert.NoError(t, err)
	assert.Equal(t, subID, subscriber.ID)
	assert.Equal(t, "test@example.com", subscriber.Email)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_Get_NotFound(t *testing.T) {
	app, store := setupSubscriberTestApp()

	store.On("Get", mock.Anything, "test-org", "non-existent").Return(domain.Subscriber{}, storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_Put(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()
	existingSubscriber := domain.Subscriber{
		ID:        subID,
		OrgID:     "test-org",
		Email:     "old@example.com",
		CreatedAt: time.Now(),
	}

	// Test PUT with existing subscriber (update)
	store.On("Get", mock.Anything, "test-org", subID).Return(existingSubscriber, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(s domain.Subscriber) bool {
		return s.ID == subID && s.Email == "new@example.com"
	})).Return(nil).Once()
	store.On("Get", mock.Anything, "test-org", subID).Return(domain.Subscriber{
		ID:        subID,
		OrgID:     "test-org",
		Email:     "new@example.com",
		CreatedAt: existingSubscriber.CreatedAt,
	}, nil).Once()

	fullBody := map[string]interface{}{
		"subscriberId": subID,
		"email":        "new@example.com",
		"preferences": map[string]interface{}{
			"channels": map[string]bool{"email": true},
			"language": "en",
		},
	}
	bodyJSON, _ := json.Marshal(fullBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/subscribers/"+subID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated domain.Subscriber
	err = json.NewDecoder(resp.Body).Decode(&updated)
	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", updated.Email)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_Put_Create(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()

	// Test PUT with non-existing subscriber (create)
	store.On("Get", mock.Anything, "test-org", subID).Return(domain.Subscriber{}, storage.ErrNotFound).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(s domain.Subscriber) bool {
		return s.ID == subID && s.Email == "new@example.com"
	})).Return(nil).Once()
	store.On("Get", mock.Anything, "test-org", subID).Return(domain.Subscriber{
		ID:    subID,
		OrgID: "test-org",
		Email: "new@example.com",
	}, nil).Once()

	fullBody := map[string]interface{}{
		"subscriberId": subID,
		"email":        "new@example.com",
		"preferences": map[string]interface{}{
			"channels": map[string]bool{"email": true},
			"language": "en",
		},
	}
	bodyJSON, _ := json.Marshal(fullBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/subscribers/"+subID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_Patch(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()
	existingSubscriber := domain.Subscriber{
		ID:    subID,
		OrgID: "test-org",
		Email: "old@example.com",
	}

	store.On("Get", mock.Anything, "test-org", subID).Return(existingSubscriber, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(s domain.Subscriber) bool {
		return s.ID == subID && s.Email == "new@example.com"
	})).Return(nil).Once()

	patch := map[string]interface{}{
		"email": "new@example.com",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/subscribers/"+subID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated domain.Subscriber
	err = json.NewDecoder(resp.Body).Decode(&updated)
	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", updated.Email)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_Patch_NotFound(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()

	store.On("Get", mock.Anything, "test-org", subID).Return(domain.Subscriber{}, storage.ErrNotFound).Once()

	patch := map[string]interface{}{
		"email": "new@example.com",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/subscribers/"+subID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_Patch_ValidationError(t *testing.T) {
	app, _ := setupSubscriberTestApp()

	subID := uuid.NewString()

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "invalid email format in patch",
			body: map[string]interface{}{
				"email": "invalid-email",
			},
		},
		{
			name: "invalid webhook URL in patch",
			body: map[string]interface{}{
				"webhookUrl": "not-a-valid-url",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/subscribers/"+subID, bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestSubscriberHandler_Delete(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()

	// Mock Delete to succeed
	store.On("Delete", mock.Anything, "test-org", subID).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscribers/"+subID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_Delete_NotFound(t *testing.T) {
	app, store := setupSubscriberTestApp()

	store.On("Delete", mock.Anything, "test-org", "non-existent").Return(storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscribers/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_List(t *testing.T) {
	app, store := setupSubscriberTestApp()

	expectedResult := domain.ListResult[domain.Subscriber]{
		Items: []domain.Subscriber{
			{ID: "sub-1", OrgID: "test-org", Email: "test1@example.com", CreatedAt: time.Now()},
			{ID: "sub-2", OrgID: "test-org", Email: "test2@example.com", CreatedAt: time.Now()},
		},
		Pagination: domain.PaginationResult{
			Page:       0,
			PageSize:   20,
			TotalCount: 2,
			TotalPages: 1,
		},
	}

	store.On("List", mock.Anything, mock.MatchedBy(func(opts domain.ListOptions) bool {
		return opts.Pagination.Page == 0 && opts.Pagination.PageSize == 20
	})).Return(expectedResult, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?page=0&pageSize=20", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result domain.ListResult[domain.Subscriber]
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Items))
	assert.Equal(t, int64(2), result.Pagination.TotalCount)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_List_WithGroupFilter(t *testing.T) {
	app, store := setupSubscriberTestApp()

	expectedResult := domain.ListResult[domain.Subscriber]{
		Items: []domain.Subscriber{
			{ID: "sub-1", OrgID: "test-org", Groups: []string{"group-123"}},
		},
		Pagination: domain.PaginationResult{
			Page:       0,
			PageSize:   20,
			TotalCount: 1,
			TotalPages: 1,
		},
	}

	store.On("List", mock.Anything, mock.Anything).Return(expectedResult, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?groupId=group-123", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_List_WithSorting(t *testing.T) {
	app, store := setupSubscriberTestApp()

	expectedResult := domain.ListResult[domain.Subscriber]{
		Items: []domain.Subscriber{
			{ID: "sub-1", OrgID: "test-org", CreatedAt: time.Now()},
		},
		Pagination: domain.PaginationResult{
			Page:       0,
			PageSize:   20,
			TotalCount: 1,
			TotalPages: 1,
		},
	}

	store.On("List", mock.Anything, mock.Anything).Return(expectedResult, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?sortBy=createdAt:desc", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_List_DefaultPagination(t *testing.T) {
	app, store := setupSubscriberTestApp()

	expectedResult := domain.ListResult[domain.Subscriber]{
		Items: []domain.Subscriber{
			{ID: "sub-1", OrgID: "test-org", Email: "test1@example.com", CreatedAt: time.Now()},
		},
		Pagination: domain.PaginationResult{
			Page:       0,
			PageSize:   20,
			TotalCount: 1,
			TotalPages: 1,
		},
	}

	store.On("List", mock.Anything, mock.MatchedBy(func(opts domain.ListOptions) bool {
		return opts.Pagination.Page == 0 && opts.Pagination.PageSize == 20
	})).Return(expectedResult, nil).Once()

	// Test without pagination parameters (should default to page 0)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result domain.ListResult[domain.Subscriber]
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Pagination.Page, "Default page should be 0")

	store.AssertExpectations(t)
}

func TestSubscriberHandler_List_InvalidPagination(t *testing.T) {
	app, store := setupSubscriberTestApp()

	expectedResult := domain.ListResult[domain.Subscriber]{
		Items: []domain.Subscriber{},
		Pagination: domain.PaginationResult{
			Page:       0,
			PageSize:   20,
			TotalCount: 0,
			TotalPages: 0,
		},
	}

	// Test with negative page (should default to 0)
	store.On("List", mock.Anything, mock.MatchedBy(func(opts domain.ListOptions) bool {
		return opts.Pagination.Page == 0
	})).Return(expectedResult, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?page=-1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	store.AssertExpectations(t)
}
