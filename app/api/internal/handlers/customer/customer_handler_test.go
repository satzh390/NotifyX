package customer

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

func setupCustomerTestApp() (*fiber.App, *MockCustomerStore) {
	app := fiber.New()
	store := new(MockCustomerStore)
	handler := NewCustomerHandler(store)

	api := app.Group("/api/v1")
	customers := api.Group("/customers")
	customers.Post("", handler.Create)
	customers.Get("/:id", handler.Get)
	customers.Put("/:id", handler.Put)
	customers.Patch("/:id", handler.Patch)
	customers.Delete("/:id", handler.Delete)
	customers.Get("", handler.List)

	return app, store
}

func TestCustomerHandler_Create(t *testing.T) {
	app, store := setupCustomerTestApp()

	customerID := uuid.NewString()
	orgID := uuid.NewString()
	expectedCustomer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "Test Customer",
		Logo:      "https://example.com/logo.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	store.On("Put", mock.Anything, mock.MatchedBy(func(c domain.Customer) bool {
		return c.Name == "Test Customer" && c.OrgID == orgID && c.Logo == "https://example.com/logo.png"
	})).Return(nil).Once()
	store.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(expectedCustomer, nil).Once()

	body := map[string]interface{}{
		"orgId": orgID,
		"name":  "Test Customer",
		"logo":  "https://example.com/logo.png",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var customer domain.Customer
	err = json.NewDecoder(resp.Body).Decode(&customer)
	assert.NoError(t, err)
	assert.Equal(t, "Test Customer", customer.Name)
	assert.Equal(t, orgID, customer.OrgID)
	assert.Equal(t, "https://example.com/logo.png", customer.Logo)

	store.AssertExpectations(t)
}

func TestCustomerHandler_Create_ValidationError(t *testing.T) {
	app, _ := setupCustomerTestApp()

	// Test missing orgId
	body := map[string]interface{}{
		"name": "Test Customer",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCustomerHandler_Get(t *testing.T) {
	app, store := setupCustomerTestApp()

	customerID := uuid.NewString()
	orgID := uuid.NewString()
	expectedCustomer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "Test Customer",
		Logo:      "https://example.com/logo.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	store.On("Get", mock.Anything, customerID).Return(expectedCustomer, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers/"+customerID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var customer domain.Customer
	err = json.NewDecoder(resp.Body).Decode(&customer)
	assert.NoError(t, err)
	assert.Equal(t, customerID, customer.ID)
	assert.Equal(t, "Test Customer", customer.Name)

	store.AssertExpectations(t)
}

func TestCustomerHandler_Get_NotFound(t *testing.T) {
	app, store := setupCustomerTestApp()

	store.On("Get", mock.Anything, "non-existent").Return(domain.Customer{}, storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestCustomerHandler_Put(t *testing.T) {
	app, store := setupCustomerTestApp()

	customerID := uuid.NewString()
	orgID := uuid.NewString()
	existingCustomer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "Old Name",
		Logo:      "https://example.com/old-logo.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	updatedCustomer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "New Name",
		Logo:      "https://example.com/new-logo.png",
		CreatedAt: existingCustomer.CreatedAt,
		UpdatedAt: time.Now(),
	}

	store.On("Get", mock.Anything, customerID).Return(existingCustomer, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(c domain.Customer) bool {
		return c.ID == customerID && c.Name == "New Name" && c.Logo == "https://example.com/new-logo.png"
	})).Return(nil).Once()
	store.On("Get", mock.Anything, customerID).Return(updatedCustomer, nil).Once()

	body := map[string]interface{}{
		"id":    customerID,
		"orgId": orgID,
		"name":  "New Name",
		"logo":  "https://example.com/new-logo.png",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/customers/"+customerID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var customer domain.Customer
	err = json.NewDecoder(resp.Body).Decode(&customer)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", customer.Name)
	assert.Equal(t, "https://example.com/new-logo.png", customer.Logo)

	store.AssertExpectations(t)
}

func TestCustomerHandler_Put_NotFound(t *testing.T) {
	app, store := setupCustomerTestApp()

	customerID := uuid.NewString()
	orgID := uuid.NewString()
	createdCustomer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "New Name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// First Get returns NotFound (entity doesn't exist)
	store.On("Get", mock.Anything, customerID).Return(domain.Customer{}, storage.ErrNotFound).Once()
	// Put creates the entity
	store.On("Put", mock.Anything, mock.MatchedBy(func(c domain.Customer) bool {
		return c.ID == customerID && c.Name == "New Name" && c.OrgID == orgID
	})).Return(nil).Once()
	// Second Get retrieves the created entity
	store.On("Get", mock.Anything, customerID).Return(createdCustomer, nil).Once()

	body := map[string]interface{}{
		"id":    customerID,
		"orgId": orgID,
		"name":  "New Name",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/customers/"+customerID, bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	// PUT with non-existent entity should create it and return 201
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestCustomerHandler_Patch(t *testing.T) {
	app, store := setupCustomerTestApp()

	customerID := uuid.NewString()
	orgID := uuid.NewString()
	existingCustomer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "Old Name",
		Logo:      "https://example.com/old-logo.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	updatedCustomer := domain.Customer{
		ID:        customerID,
		OrgID:     orgID,
		Name:      "New Name",
		Logo:      "https://example.com/old-logo.png",
		CreatedAt: existingCustomer.CreatedAt,
		UpdatedAt: time.Now(),
	}

	store.On("Get", mock.Anything, customerID).Return(existingCustomer, nil).Once()
	store.On("Put", mock.Anything, mock.MatchedBy(func(c domain.Customer) bool {
		return c.ID == customerID && c.Name == "New Name"
	})).Return(nil).Once()
	store.On("Get", mock.Anything, customerID).Return(updatedCustomer, nil).Once()

	patch := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/customers/"+customerID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var customer domain.Customer
	err = json.NewDecoder(resp.Body).Decode(&customer)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", customer.Name)

	store.AssertExpectations(t)
}

func TestCustomerHandler_Patch_NotFound(t *testing.T) {
	app, store := setupCustomerTestApp()

	customerID := uuid.NewString()
	store.On("Get", mock.Anything, customerID).Return(domain.Customer{}, storage.ErrNotFound).Once()

	patch := map[string]interface{}{
		"name": "New Name",
	}
	patchJSON, _ := json.Marshal(patch)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/customers/"+customerID, bytes.NewReader(patchJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestCustomerHandler_Delete(t *testing.T) {
	app, store := setupCustomerTestApp()

	customerID := uuid.NewString()
	store.On("Delete", mock.Anything, customerID).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/customers/"+customerID, nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestCustomerHandler_Delete_NotFound(t *testing.T) {
	app, store := setupCustomerTestApp()

	store.On("Delete", mock.Anything, "non-existent").Return(storage.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/customers/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	store.AssertExpectations(t)
}

func TestCustomerHandler_List(t *testing.T) {
	app, store := setupCustomerTestApp()

	expectedResult := domain.ListResult[domain.Customer]{
		Items: []domain.Customer{
			{ID: "customer-1", OrgID: "org-1", Name: "Customer 1"},
			{ID: "customer-2", OrgID: "org-1", Name: "Customer 2"},
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers?page=0&pageSize=20", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result domain.ListResult[domain.Customer]
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Pagination.TotalCount)
	assert.Len(t, result.Items, 2)

	store.AssertExpectations(t)
}

