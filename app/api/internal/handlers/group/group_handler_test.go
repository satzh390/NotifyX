package group

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupGroupTestApp() (*fiber.App, *MockGroupStore) {
	app := fiber.New()
	store := new(MockGroupStore)
	handler := NewGroupHandler(store)

	api := app.Group("/api/v1")
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("customerId", "test-customer")
		return c.Next()
	})

	groups := api.Group("/groups")
	groups.Post("", handler.Create)
	groups.Get("/:id", handler.Get)
	groups.Put("/:id", handler.Put)
	groups.Patch("/:id", handler.Patch)
	groups.Delete("/:id", handler.Delete)
	groups.Get("", handler.List)

	return app, store
}

func TestGroupHandler_Create(t *testing.T) {
	app, store := setupGroupTestApp()

	// Mock Put to succeed
	store.On("Put", mock.Anything, mock.MatchedBy(func(g domain.Group) bool {
		return g.Name == "Test Group" && g.CustomerID == "test-customer"
	})).Return(nil).Once()

	body := map[string]interface{}{
		"name":        "Test Group",
		"description": "Test Description",
		"subscribers": []string{"sub-1", "sub-2"},
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var group domain.Group
	err = json.NewDecoder(resp.Body).Decode(&group)
	assert.NoError(t, err)
	assert.Equal(t, "Test Group", group.Name)
	assert.Equal(t, "test-customer", group.CustomerID)

	store.AssertExpectations(t)
}

func TestGroupHandler_Create_ValidationError(t *testing.T) {
	app, _ := setupGroupTestApp()

	// Test missing name
	body := map[string]interface{}{
		"description": "Test Description",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGroupHandler_Get(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()
	expectedGroup := domain.Group{
		ID:          groupID,
		CustomerID:  "test-customer",
		Name:        "Test Group",
		Description: "Test Description",
	}

	store.On("Get", mock.Anything, "test-customer", groupID).Return(expectedGroup, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/"+groupID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var group domain.Group
	err = json.NewDecoder(resp.Body).Decode(&group)
	assert.NoError(t, err)
	assert.Equal(t, groupID, group.ID)
	assert.Equal(t, "Test Group", group.Name)

	store.AssertExpectations(t)
}

func TestGroupHandler_Get_NotFound(t *testing.T) {
	app, store := setupGroupTestApp()

	store.On("Get", mock.Anything, "test-customer", "non-existent").Return(domain.Group{}, storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestGroupHandler_Put(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()
	existingGroup := domain.Group{
		ID:         groupID,
		CustomerID: "test-customer",
		Name:       "Old Name",
	}

	// Test PUT with existing group (update)
	store.On("Get", mock.Anything, "test-customer", groupID).Return(existingGroup, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(g domain.Group) bool {
		return g.ID == groupID && g.Name == "New Name"
	})).Return(nil).Once()

	fullBody := map[string]interface{}{
		"groupId":     groupID,
		"name":        "New Name",
		"description": "New Description",
	}
	bodyJSON, _ := json.Marshal(fullBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/groups/"+groupID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated domain.Group
	err = json.NewDecoder(resp.Body).Decode(&updated)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)

	store.AssertExpectations(t)
}

func TestGroupHandler_Put_Create(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()

	// Test PUT with non-existing group (create)
	store.On("Get", mock.Anything, "test-customer", groupID).Return(domain.Group{}, storage.ErrNotFound).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(g domain.Group) bool {
		return g.ID == groupID && g.Name == "New Group"
	})).Return(nil).Once()

	fullBody := map[string]interface{}{
		"groupId": groupID,
		"name":    "New Group",
	}
	bodyJSON, _ := json.Marshal(fullBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/groups/"+groupID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestGroupHandler_Patch(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()
	existingGroup := domain.Group{
		ID:         groupID,
		CustomerID: "test-customer",
		Name:       "Old Name",
	}

	store.On("Get", mock.Anything, "test-customer", groupID).Return(existingGroup, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(g domain.Group) bool {
		return g.ID == groupID && g.Name == "New Name"
	})).Return(nil).Once()

	patch := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/groups/"+groupID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated domain.Group
	err = json.NewDecoder(resp.Body).Decode(&updated)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)

	store.AssertExpectations(t)
}

func TestGroupHandler_Patch_NotFound(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()

	store.On("Get", mock.Anything, "test-customer", groupID).Return(domain.Group{}, storage.ErrNotFound).Once()

	patch := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/groups/"+groupID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestGroupHandler_Delete(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()

	// Mock Delete to succeed
	store.On("Delete", mock.Anything, "test-customer", groupID).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/"+groupID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestGroupHandler_Delete_NotFound(t *testing.T) {
	app, store := setupGroupTestApp()

	store.On("Delete", mock.Anything, "test-customer", "non-existent").Return(storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestGroupHandler_List(t *testing.T) {
	app, store := setupGroupTestApp()

	expectedResult := domain.ListResult[domain.Group]{
		Items: []domain.Group{
			{ID: "group-1", CustomerID: "test-customer", Name: "Group 1"},
			{ID: "group-2", CustomerID: "test-customer", Name: "Group 2"},
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups?page=0&pageSize=20", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result domain.ListResult[domain.Group]
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Pagination.TotalCount, int64(2))

	store.AssertExpectations(t)
}

func TestGroupHandler_List_WithSorting(t *testing.T) {
	app, store := setupGroupTestApp()

	expectedResult := domain.ListResult[domain.Group]{
		Items: []domain.Group{
			{ID: "group-1", CustomerID: "test-customer", Name: "A Group"},
			{ID: "group-2", CustomerID: "test-customer", Name: "B Group"},
		},
		Pagination: domain.PaginationResult{
			Page:       0,
			PageSize:   20,
			TotalCount: 2,
			TotalPages: 1,
		},
	}

	store.On("List", mock.Anything, mock.Anything).Return(expectedResult, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups?sortBy=name:asc", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	store.AssertExpectations(t)
}
