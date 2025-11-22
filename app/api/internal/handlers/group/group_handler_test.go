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
		c.Locals("orgId", "test-org")
		return c.Next()
	})

	groups := api.Group("/groups")
	groups.Post("", handler.Create)
	groups.Get("/:id", handler.Get)
	groups.Put("/:id", handler.Update)
	groups.Delete("/:id", handler.Delete)
	groups.Get("", handler.List)

	return app, store
}

func TestGroupHandler_Create(t *testing.T) {
	app, store := setupGroupTestApp()

	// Mock Put to succeed
	store.On("Put", mock.Anything, mock.MatchedBy(func(g domain.Group) bool {
		return g.Name == "Test Group" && g.OrgID == "test-org"
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
	assert.Equal(t, "test-org", group.OrgID)

	store.AssertExpectations(t)
}

func TestGroupHandler_Get(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()
	expectedGroup := domain.Group{
		ID:          groupID,
		OrgID:       "test-org",
		Name:        "Test Group",
		Description: "Test Description",
	}

	store.On("Get", mock.Anything, "test-org", groupID).Return(expectedGroup, nil).Once()

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

	store.On("Get", mock.Anything, "test-org", "non-existent").Return(domain.Group{}, storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestGroupHandler_Update(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()
	existingGroup := domain.Group{
		ID:    groupID,
		OrgID: "test-org",
		Name:  "Old Name",
	}

	// Mock Get to return existing group
	store.On("Get", mock.Anything, "test-org", groupID).Return(existingGroup, nil).Once()
	// Mock Put to update the group
	store.On("Put", mock.Anything, mock.MatchedBy(func(g domain.Group) bool {
		return g.ID == groupID && g.Name == "New Name"
	})).Return(nil).Once()

	patch := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/groups/"+groupID, bytes.NewReader(patchJSON))
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

func TestGroupHandler_Delete(t *testing.T) {
	app, store := setupGroupTestApp()

	groupID := uuid.NewString()
	existingGroup := domain.Group{
		ID:    groupID,
		OrgID: "test-org",
		Name:  "Test Group",
	}

	// Mock Get to return existing group
	store.On("Get", mock.Anything, "test-org", groupID).Return(existingGroup, nil).Once()
	// Mock Delete to succeed
	store.On("Delete", mock.Anything, "test-org", groupID).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/"+groupID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestGroupHandler_Delete_NotFound(t *testing.T) {
	app, store := setupGroupTestApp()

	store.On("Get", mock.Anything, "test-org", "non-existent").Return(domain.Group{}, storage.ErrNotFound).Once()

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
			{ID: "group-1", OrgID: "test-org", Name: "Group 1"},
			{ID: "group-2", OrgID: "test-org", Name: "Group 2"},
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
			{ID: "group-1", OrgID: "test-org", Name: "A Group"},
			{ID: "group-2", OrgID: "test-org", Name: "B Group"},
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
