package subscriber

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
	subscribers.Put("/:id", handler.Update)
	subscribers.Delete("/:id", handler.Delete)
	subscribers.Get("", handler.List)

	return app, store
}

func TestSubscriberHandler_Create(t *testing.T) {
	app, store := setupSubscriberTestApp()

	store.On("Put", mock.Anything, mock.MatchedBy(func(s domain.Subscriber) bool {
		return s.Email == "test@example.com" && s.OrgID == "test-org"
	})).Return(nil).Once()

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

func TestSubscriberHandler_Update(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodPut, "/api/v1/subscribers/"+subID, bytes.NewReader(patchJSON))
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

func TestSubscriberHandler_Delete(t *testing.T) {
	app, store := setupSubscriberTestApp()

	subID := uuid.NewString()
	existingSubscriber := domain.Subscriber{
		ID:    subID,
		OrgID: "test-org",
		Email: "test@example.com",
	}

	store.On("Get", mock.Anything, "test-org", subID).Return(existingSubscriber, nil).Once()
	store.On("Delete", mock.Anything, "test-org", subID).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscribers/"+subID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestSubscriberHandler_Delete_NotFound(t *testing.T) {
	app, store := setupSubscriberTestApp()

	store.On("Get", mock.Anything, "test-org", "non-existent").Return(domain.Subscriber{}, storage.ErrNotFound).Once()

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
			Page:       1,
			PageSize:   20,
			TotalCount: 2,
			TotalPages: 1,
		},
	}

	store.On("List", mock.Anything, mock.MatchedBy(func(opts domain.ListOptions) bool {
		return opts.Pagination.Page == 1 && opts.Pagination.PageSize == 20
	})).Return(expectedResult, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscribers?page=1&pageSize=20", nil)
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
			Page:       1,
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
			Page:       1,
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
