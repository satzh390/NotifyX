package customer

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/httpx"
)

type CustomerHandler struct {
	store storage.CustomerStore
}

func NewCustomerHandler(store storage.CustomerStore) *CustomerHandler {
	return &CustomerHandler{store: store}
}

type CustomerRequest struct {
	// ID - required for PUT, optional for POST (will be auto-generated if not provided)
	ID string `json:"id" example:"customer-123"`
	// OrgID is required - the organization ID this customer belongs to
	OrgID string `json:"orgId" validate:"required" example:"org-123"`
	// Name is required - the name of the customer
	Name string `json:"name" validate:"required" example:"Acme Customer"`
	// Logo - optional logo URL
	Logo string `json:"logo,omitempty" example:"https://example.com/logo.png"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CustomerPatchRequest is used for PATCH requests - only mutable fields, no required validation
type CustomerPatchRequest struct {
	// Name - optional, the name of the customer
	Name string `json:"name" example:"Acme Customer"`
	// Logo - optional logo URL
	Logo string `json:"logo,omitempty" example:"https://example.com/logo.png"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CreateCustomer godoc
// @Summary Create a new customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param customer body CustomerRequest true "Customer data (orgId and name are required)"
// @Success 201 {object} domain.Customer
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 500 {object} map[string]string
// @Router /customers [post]
func (handler *CustomerHandler) Create(fiberCtx *fiber.Ctx) error {
	body, err := httpx.ParseAndValidateBody[CustomerRequest](fiberCtx)
	if err != nil {
		return err
	}

	customerID := body.ID
	if customerID == "" {
		customerID = uuid.NewString()
	}

	customer := domain.Customer{
		ID:        customerID,
		OrgID:     body.OrgID,
		Name:      body.Name,
		Logo:      body.Logo,
		Metadata:  body.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := handler.store.Put(context.Background(), customer); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the created customer to get CreatedAt set by the store
	created, err := handler.store.Get(context.Background(), customerID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve created customer: "+err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(created)
}

// GetCustomer godoc
// @Summary Get a customer by ID
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID"
// @Success 200 {object} domain.Customer
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customers/{id} [get]
func (handler *CustomerHandler) Get(fiberCtx *fiber.Ctx) error {
	customerID := fiberCtx.Params("id")
	if customerID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing customer id")
	}

	customer, err := handler.store.Get(context.Background(), customerID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "customer not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(customer)
}

// PutCustomer godoc
// @Summary Create or update a customer (full object)
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID"
// @Param customer body CustomerRequest true "Customer data (orgId and name are required)"
// @Success 200 {object} domain.Customer "Updated customer"
// @Success 201 {object} domain.Customer "Created customer"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customers/{id} [put]
func (handler *CustomerHandler) Put(fiberCtx *fiber.Ctx) error {
	customerID := fiberCtx.Params("id")
	if customerID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing customer id")
	}

	body, err := httpx.ParseAndValidateBody[CustomerRequest](fiberCtx)
	if err != nil {
		return err
	}

	// For PUT, id is required in the body
	if body.ID == "" {
		return fiber.NewError(http.StatusBadRequest, "id is required in request body")
	}

	// Ensure the ID in the path matches the ID in the body
	if body.ID != customerID {
		return fiber.NewError(http.StatusBadRequest, "customer id in path must match customer id in body")
	}

	// Check if customer exists
	_, err = handler.store.Get(context.Background(), customerID)
	exists := err == nil

	customer := domain.Customer{
		ID:        customerID,
		OrgID:     body.OrgID,
		Name:      body.Name,
		Logo:      body.Logo,
		Metadata:  body.Metadata,
		UpdatedAt: time.Now(),
	}

	if err := handler.store.Put(context.Background(), customer); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the customer to get CreatedAt/UpdatedAt set by the store
	updated, err := handler.store.Get(context.Background(), customerID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve customer: "+err.Error())
	}

	statusCode := http.StatusOK
	if !exists {
		statusCode = http.StatusCreated
	}

	return fiberCtx.Status(statusCode).JSON(updated)
}

// PatchCustomer godoc
// @Summary Partially update a customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID"
// @Param customer body CustomerPatchRequest true "Customer data (all fields optional)"
// @Success 200 {object} domain.Customer
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customers/{id} [patch]
func (handler *CustomerHandler) Patch(fiberCtx *fiber.Ctx) error {
	customerID := fiberCtx.Params("id")
	if customerID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing customer id")
	}

	existing, err := handler.store.Get(context.Background(), customerID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "customer not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Validate patch body and get raw bytes for merge patch
	patchBytes, err := httpx.ValidatePatchBody[CustomerPatchRequest](fiberCtx)
	if err != nil {
		return err
	}

	// Apply merge patch (automatically protects immutable fields)
	if err := httpx.MergePatch(&existing, patchBytes); err != nil {
		return fiber.NewError(http.StatusBadRequest, "failed to apply patch: "+err.Error())
	}

	existing.UpdatedAt = time.Now()

	if err := handler.store.Put(context.Background(), existing); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	updated, err := handler.store.Get(context.Background(), customerID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve updated customer: "+err.Error())
	}

	return fiberCtx.JSON(updated)
}

// DeleteCustomer godoc
// @Summary Delete a customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Customer ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customers/{id} [delete]
func (handler *CustomerHandler) Delete(fiberCtx *fiber.Ctx) error {
	customerID := fiberCtx.Params("id")
	if customerID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing customer id")
	}

	if err := handler.store.Delete(context.Background(), customerID); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "customer not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}

// ListCustomers godoc
// @Summary List customers
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (0-based)" default(0)
// @Param pageSize query int false "Page size" default(20)
// @Param sortBy query string false "Sort fields (e.g., name:asc,createdAt:desc)"
// @Success 200 {object} domain.ListResult[domain.Customer]
// @Failure 500 {object} map[string]string
// @Router /customers [get]
func (handler *CustomerHandler) List(fiberCtx *fiber.Ctx) error {
	opts := httpx.ParseListOptions(fiberCtx, "")
	result, err := handler.store.List(context.Background(), opts)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(result)
}
