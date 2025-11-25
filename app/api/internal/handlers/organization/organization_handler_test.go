package organization

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

func setupOrganizationTestApp() (*fiber.App, *MockOrganizationStore) {
	app := fiber.New()
	store := new(MockOrganizationStore)
	handler := NewOrganizationHandler(store)

	api := app.Group("/api/v1")
	groups := api.Group("/organizations")
	groups.Post("", handler.Create)
	groups.Get("/:id", handler.Get)
	groups.Put("/:id", handler.Put)
	groups.Patch("/:id", handler.Patch)
	groups.Delete("/:id", handler.Delete)
	groups.Get("", handler.List)

	return app, store
}

func TestOrganizationHandler_Create(t *testing.T) {
	app, store := setupOrganizationTestApp()

	orgID := uuid.NewString()
	expectedOrg := domain.Organization{
		ID:        orgID,
		Name:      "Test Organization",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	store.On("Put", mock.Anything, mock.MatchedBy(func(o domain.Organization) bool {
		return o.Name == "Test Organization" && o.Type == domain.OrganizationTypeCompany
	})).Return(nil).Once()
	store.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(expectedOrg, nil).Once()

	body := map[string]interface{}{
		"name": "Test Organization",
		"type": "company",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var org domain.Organization
	err = json.NewDecoder(resp.Body).Decode(&org)
	assert.NoError(t, err)
	assert.Equal(t, "Test Organization", org.Name)
	assert.Equal(t, domain.OrganizationTypeCompany, org.Type)

	store.AssertExpectations(t)
}

func TestOrganizationHandler_Create_ValidationError(t *testing.T) {
	app, _ := setupOrganizationTestApp()

	// Test missing name
	body := map[string]interface{}{
		"type": "company",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestOrganizationHandler_Get(t *testing.T) {
	app, store := setupOrganizationTestApp()

	orgID := uuid.NewString()
	expectedOrg := domain.Organization{
		ID:        orgID,
		Name:      "Test Organization",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	store.On("Get", mock.Anything, orgID).Return(expectedOrg, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+orgID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var org domain.Organization
	err = json.NewDecoder(resp.Body).Decode(&org)
	assert.NoError(t, err)
	assert.Equal(t, orgID, org.ID)
	assert.Equal(t, "Test Organization", org.Name)

	store.AssertExpectations(t)
}

func TestOrganizationHandler_Get_NotFound(t *testing.T) {
	app, store := setupOrganizationTestApp()

	store.On("Get", mock.Anything, "non-existent").Return(domain.Organization{}, storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestOrganizationHandler_Put(t *testing.T) {
	app, store := setupOrganizationTestApp()

	orgID := uuid.NewString()
	existingOrg := domain.Organization{
		ID:        orgID,
		Name:      "Old Name",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	updatedOrg := domain.Organization{
		ID:        orgID,
		Name:      "New Name",
		Type:      domain.OrganizationTypeSaaSProvider,
		CreatedAt: existingOrg.CreatedAt,
		UpdatedAt: time.Now(),
	}

	store.On("Get", mock.Anything, orgID).Return(existingOrg, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(o domain.Organization) bool {
		return o.ID == orgID && o.Name == "New Name" && o.Type == domain.OrganizationTypeSaaSProvider
	})).Return(nil).Once()
	store.On("Get", mock.Anything, orgID).Return(updatedOrg, nil).Once()

	body := map[string]interface{}{
		"id":   orgID,
		"name": "New Name",
		"type": "saasProvider",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/organizations/"+orgID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var org domain.Organization
	err = json.NewDecoder(resp.Body).Decode(&org)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", org.Name)
	assert.Equal(t, domain.OrganizationTypeSaaSProvider, org.Type)

	store.AssertExpectations(t)
}

func TestOrganizationHandler_Put_NotFound(t *testing.T) {
	app, store := setupOrganizationTestApp()

	orgID := uuid.NewString()
	store.On("Get", mock.Anything, orgID).Return(domain.Organization{}, storage.ErrNotFound).Once()

	body := map[string]interface{}{
		"id":   orgID,
		"name": "New Name",
		"type": "company",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/organizations/"+orgID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestOrganizationHandler_Patch(t *testing.T) {
	app, store := setupOrganizationTestApp()

	orgID := uuid.NewString()
	existingOrg := domain.Organization{
		ID:        orgID,
		Name:      "Old Name",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	updatedOrg := domain.Organization{
		ID:        orgID,
		Name:      "New Name",
		Type:      domain.OrganizationTypeCompany,
		CreatedAt: existingOrg.CreatedAt,
		UpdatedAt: time.Now(),
	}

	store.On("Get", mock.Anything, orgID).Return(existingOrg, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(o domain.Organization) bool {
		return o.ID == orgID && o.Name == "New Name"
	})).Return(nil).Once()
	store.On("Get", mock.Anything, orgID).Return(updatedOrg, nil).Once()

	patch := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/organizations/"+orgID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var org domain.Organization
	err = json.NewDecoder(resp.Body).Decode(&org)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", org.Name)

	store.AssertExpectations(t)
}

func TestOrganizationHandler_Patch_NotFound(t *testing.T) {
	app, store := setupOrganizationTestApp()

	orgID := uuid.NewString()
	store.On("Get", mock.Anything, orgID).Return(domain.Organization{}, storage.ErrNotFound).Once()

	patch := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/organizations/"+orgID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestOrganizationHandler_Delete(t *testing.T) {
	app, store := setupOrganizationTestApp()

	orgID := uuid.NewString()
	store.On("Delete", mock.Anything, orgID).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/"+orgID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestOrganizationHandler_Delete_NotFound(t *testing.T) {
	app, store := setupOrganizationTestApp()

	store.On("Delete", mock.Anything, "non-existent").Return(storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestOrganizationHandler_List(t *testing.T) {
	app, store := setupOrganizationTestApp()

	expectedResult := domain.ListResult[domain.Organization]{
		Items: []domain.Organization{
			{ID: "org-1", Name: "Organization 1", Type: domain.OrganizationTypeCompany},
			{ID: "org-2", Name: "Organization 2", Type: domain.OrganizationTypeSaaSProvider},
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations?page=0&pageSize=20", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result domain.ListResult[domain.Organization]
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Pagination.TotalCount)
	assert.Len(t, result.Items, 2)

	store.AssertExpectations(t)
}
