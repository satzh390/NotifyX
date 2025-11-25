package group

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/httpx"
)

type GroupHandler struct {
	store storage.GroupStore
}

func NewGroupHandler(store storage.GroupStore) *GroupHandler {
	return &GroupHandler{store: store}
}

type GroupRequest struct {
	// GroupID - required for PUT, optional for POST (will be auto-generated if not provided)
	GroupID string `json:"groupId" example:"group-123"`
	// Name is required - the name of the group
	Name string `json:"name" validate:"required" example:"VIP Customers"`
	// Description - optional description of the group
	Description string `json:"description"`
	// Subscribers - list of subscriber IDs in this group
	Subscribers []string `json:"subscribers"`
	// SubscribedEventTypes - list of event types this group opted into
	SubscribedEventTypes []string `json:"subscribedEventTypes"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata"`
}

// GroupPatchRequest is used for PATCH requests - only mutable fields, no required validation
type GroupPatchRequest struct {
	// Name - optional, the name of the group
	Name string `json:"name" example:"VIP Customers"`
	// Description - optional description of the group
	Description string `json:"description"`
	// Subscribers - list of subscriber IDs in this group
	Subscribers []string `json:"subscribers"`
	// SubscribedEventTypes - list of event types this group opted into
	SubscribedEventTypes []string `json:"subscribedEventTypes"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata"`
}

// CreateGroup godoc
// @Summary Create a new group
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group body GroupRequest true "Group data (name is required)"
// @Success 201 {object} domain.Group
// @Failure 400 {object} map[string]string "Bad request - validation error (e.g., name is required)"
// @Failure 500 {object} map[string]string
// @Router /groups [post]
func (handler *GroupHandler) Create(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	body, err := httpx.ParseAndValidateBody[GroupRequest](fiberCtx)
	if err != nil {
		return err
	}

	groupID := body.GroupID
	if groupID == "" {
		groupID = uuid.NewString()
	}

	group := domain.Group{
		ID:                   groupID,
		CustomerID:           customerID,
		Name:                 body.Name,
		Description:          body.Description,
		Subscribers:          body.Subscribers,
		SubscribedEventTypes: body.SubscribedEventTypes,
		Metadata:             body.Metadata,
	}

	if err := handler.store.Put(context.Background(), group); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(group)
}

// GetGroup godoc
// @Summary Get a group by ID
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {object} domain.Group
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id} [get]
func (handler *GroupHandler) Get(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	groupID := fiberCtx.Params("id")
	if groupID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing group id")
	}

	group, err := handler.store.Get(context.Background(), customerID, groupID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "group not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(group)
}

// PutGroup godoc
// @Summary Create or update a group (full object)
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param group body GroupRequest true "Group data (all fields including groupId and name required)"
// @Success 200 {object} domain.Group "Updated group"
// @Success 201 {object} domain.Group "Created group"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id} [put]
func (handler *GroupHandler) Put(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	groupID := fiberCtx.Params("id")
	if groupID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing group id")
	}

	body, err := httpx.ParseAndValidateBody[GroupRequest](fiberCtx)
	if err != nil {
		return err
	}

	// For PUT, groupId is required in the body
	if body.GroupID == "" {
		return fiber.NewError(http.StatusBadRequest, "groupId is required in request body")
	}

	// Ensure the ID in the path matches the ID in the body
	if body.GroupID != groupID {
		return fiber.NewError(http.StatusBadRequest, "group id in path must match group id in body")
	}

	// Check if group exists
	_, err = handler.store.Get(context.Background(), customerID, groupID)
	exists := err == nil

	group := domain.Group{
		ID:                   groupID,
		CustomerID:           customerID,
		Name:                 body.Name,
		Description:          body.Description,
		Subscribers:          body.Subscribers,
		SubscribedEventTypes: body.SubscribedEventTypes,
		Metadata:             body.Metadata,
	}

	if err := handler.store.Put(context.Background(), group); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	statusCode := http.StatusOK
	if !exists {
		statusCode = http.StatusCreated
	}

	return fiberCtx.Status(statusCode).JSON(group)
}

// PatchGroup godoc
// @Summary Partially update a group
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param group body GroupPatchRequest true "Group patch data (only fields to update)"
// @Success 200 {object} domain.Group
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id} [patch]
func (handler *GroupHandler) Patch(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	groupID := fiberCtx.Params("id")
	if groupID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing group id")
	}

	// Validate patch body and get raw bytes for merge patch
	patchData, err := httpx.ValidatePatchBody[GroupPatchRequest](fiberCtx)
	if err != nil {
		return err
	}

	existing, err := handler.store.Get(context.Background(), customerID, groupID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "group not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Apply merge patch
	if err := httpx.MergePatch(&existing, patchData); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid patch: "+err.Error())
	}

	if err := handler.store.Put(context.Background(), existing); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(existing)
}

// DeleteGroup godoc
// @Summary Delete a group
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id} [delete]
func (handler *GroupHandler) Delete(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	groupID := fiberCtx.Params("id")
	if groupID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing group id")
	}

	if err := handler.store.Delete(context.Background(), customerID, groupID); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "group not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}

// ListGroups godoc
// @Summary List groups with pagination
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (0-based)" default(0)
// @Param pageSize query int false "Page size" default(10)
// @Param sortBy query string false "Sort field"
// @Param sortOrder query string false "Sort order (asc/desc)" default(asc)
// @Success 200 {object} domain.ListResult[domain.Group]
// @Failure 500 {object} map[string]string
// @Router /groups [get]
func (handler *GroupHandler) List(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	opts := httpx.ParseListOptions(fiberCtx, customerID)
	if eventType := httpx.EventTypeFilterFromQuery(fiberCtx); eventType != "" {
		opts.Filter["subscribedEventTypes"] = eventType
	}

	result, err := handler.store.List(context.Background(), opts)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(result)
}
