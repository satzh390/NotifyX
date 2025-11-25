package organization

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

type OrganizationHandler struct {
	store storage.OrganizationStore
}

func NewOrganizationHandler(store storage.OrganizationStore) *OrganizationHandler {
	return &OrganizationHandler{store: store}
}

type OrganizationRequest struct {
	// ID - required for PUT, optional for POST (will be auto-generated if not provided)
	ID string `json:"id" example:"org-123"`
	// Name is required - the name of the organization
	Name string `json:"name" validate:"required" example:"Acme Corp"`
	// Type is required - the type of organization (company | saasProvider)
	Type domain.OrganizationType `json:"type" validate:"required" example:"company"`
}

// OrganizationPatchRequest is used for PATCH requests - only mutable fields, no required validation
type OrganizationPatchRequest struct {
	// Name - optional, the name of the organization
	Name string `json:"name" example:"Acme Corp"`
	// Type - optional, the type of organization (company | saasProvider)
	Type domain.OrganizationType `json:"type" example:"company"`
}

// CreateOrganization godoc
// @Summary Create a new organization
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization body OrganizationRequest true "Organization data (name and type are required)"
// @Success 201 {object} domain.Organization
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 500 {object} map[string]string
// @Router /organizations [post]
func (handler *OrganizationHandler) Create(fiberCtx *fiber.Ctx) error {
	body, err := httpx.ParseAndValidateBody[OrganizationRequest](fiberCtx)
	if err != nil {
		return err
	}

	orgID := body.ID
	if orgID == "" {
		orgID = uuid.NewString()
	}

	org := domain.Organization{
		ID:        orgID,
		Name:      body.Name,
		Type:      body.Type,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := handler.store.Put(context.Background(), org); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the created organization to get CreatedAt set by the store
	created, err := handler.store.Get(context.Background(), orgID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve created organization: "+err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(created)
}

// GetOrganization godoc
// @Summary Get an organization by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Success 200 {object} domain.Organization
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /organizations/{id} [get]
func (handler *OrganizationHandler) Get(fiberCtx *fiber.Ctx) error {
	orgID := fiberCtx.Params("id")
	if orgID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing organization id")
	}

	org, err := handler.store.Get(context.Background(), orgID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "organization not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(org)
}

// PutOrganization godoc
// @Summary Create or update an organization (full object)
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Param organization body OrganizationRequest true "Organization data (name and type are required)"
// @Success 200 {object} domain.Organization "Updated organization"
// @Success 201 {object} domain.Organization "Created organization"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /organizations/{id} [put]
func (handler *OrganizationHandler) Put(fiberCtx *fiber.Ctx) error {
	orgID := fiberCtx.Params("id")
	if orgID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing organization id")
	}

	body, err := httpx.ParseAndValidateBody[OrganizationRequest](fiberCtx)
	if err != nil {
		return err
	}

	// For PUT, id is required in the body
	if body.ID == "" {
		return fiber.NewError(http.StatusBadRequest, "id is required in request body")
	}

	// Ensure the ID in the path matches the ID in the body
	if body.ID != orgID {
		return fiber.NewError(http.StatusBadRequest, "organization id in path must match organization id in body")
	}

	// Check if organization exists
	_, err = handler.store.Get(context.Background(), orgID)
	exists := err == nil

	org := domain.Organization{
		ID:        orgID,
		Name:      body.Name,
		Type:      body.Type,
		UpdatedAt: time.Now(),
	}

	if err := handler.store.Put(context.Background(), org); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the organization to get CreatedAt/UpdatedAt set by the store
	updated, err := handler.store.Get(context.Background(), orgID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve organization: "+err.Error())
	}

	statusCode := http.StatusOK
	if !exists {
		statusCode = http.StatusCreated
	}

	return fiberCtx.Status(statusCode).JSON(updated)
}

// PatchOrganization godoc
// @Summary Partially update an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Param organization body OrganizationPatchRequest true "Organization data (all fields optional)"
// @Success 200 {object} domain.Organization
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /organizations/{id} [patch]
func (handler *OrganizationHandler) Patch(fiberCtx *fiber.Ctx) error {
	orgID := fiberCtx.Params("id")
	if orgID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing organization id")
	}

	existing, err := handler.store.Get(context.Background(), orgID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "organization not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Validate patch body and get raw bytes for merge patch
	patchBytes, err := httpx.ValidatePatchBody[OrganizationPatchRequest](fiberCtx)
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

	updated, err := handler.store.Get(context.Background(), orgID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve updated organization: "+err.Error())
	}

	return fiberCtx.JSON(updated)
}

// DeleteOrganization godoc
// @Summary Delete an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /organizations/{id} [delete]
func (handler *OrganizationHandler) Delete(fiberCtx *fiber.Ctx) error {
	orgID := fiberCtx.Params("id")
	if orgID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing organization id")
	}

	if err := handler.store.Delete(context.Background(), orgID); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "organization not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}

// ListOrganizations godoc
// @Summary List organizations
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (0-based)" default(0)
// @Param pageSize query int false "Page size" default(20)
// @Param sortBy query string false "Sort fields (e.g., name:asc,createdAt:desc)"
// @Success 200 {object} domain.ListResult[domain.Organization]
// @Failure 500 {object} map[string]string
// @Router /organizations [get]
func (handler *OrganizationHandler) List(fiberCtx *fiber.Ctx) error {
	opts := httpx.ParseListOptions(fiberCtx, "")
	result, err := handler.store.List(context.Background(), opts)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(result)
}
