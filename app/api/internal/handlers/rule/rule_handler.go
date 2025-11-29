package rule

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/httpx"
)

type RuleHandler struct {
	store storage.RuleStore
}

func NewRuleHandler(store storage.RuleStore) *RuleHandler {
	return &RuleHandler{store: store}
}

type RuleRequest struct {
	// EventType is required - the event type identifier (e.g., "order.created")
	EventType string `json:"eventType" validate:"required" example:"order.created"`
	// Channels - list of notification channels for this rule
	Channels []domain.ChannelType `json:"channels"`
	// TemplateRefs - template references by channel type
	TemplateRefs map[domain.ChannelType]string `json:"templateRefs"`
	// CustomFilter - optional custom filter configuration. The filter type must match a registered custom filter name.
	// Custom filters are NOT applied to direct emails and phone numbers.
	CustomFilter *domain.CustomFilterConfig `json:"customFilter,omitempty"`
}

// RulePatchRequest is used for PATCH requests - only mutable fields, no required validation
type RulePatchRequest struct {
	// Channels - list of notification channels for this rule
	Channels []domain.ChannelType `json:"channels"`
	// TemplateRefs - template references by channel type
	TemplateRefs map[domain.ChannelType]string `json:"templateRefs"`
	// CustomFilter - optional custom filter configuration. The filter type must match a registered custom filter name.
	// Custom filters are NOT applied to direct emails and phone numbers.
	CustomFilter *domain.CustomFilterConfig `json:"customFilter,omitempty"`
}

// CreateRule godoc
// @Summary Create a new rule
// @Tags rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rule body RuleRequest true "Rule data (eventType is required)"
// @Success 201 {object} domain.Rule
// @Failure 400 {object} map[string]string "Bad request - validation error (e.g., eventType is required)"
// @Failure 500 {object} map[string]string
// @Router /rules [post]
func (handler *RuleHandler) Create(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	body, err := httpx.ParseAndValidateBody[RuleRequest](fiberCtx)
	if err != nil {
		return err
	}

	rule := domain.Rule{
		EventType:    body.EventType,
		CustomerID:   customerID,
		Channels:     body.Channels,
		TemplateRefs: body.TemplateRefs,
		CustomFilter: body.CustomFilter,
	}

	if err := handler.store.Put(context.Background(), rule); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the created rule to get CreatedAt set by the store
	created, err := handler.store.Get(context.Background(), customerID, rule.EventType)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve created rule: "+err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(created)
}

// GetRule godoc
// @Summary Get a rule by event type
// @Tags rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param eventType path string true "Event Type"
// @Success 200 {object} domain.Rule
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rules/{eventType} [get]
func (handler *RuleHandler) Get(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	eventType := fiberCtx.Params("eventType")
	if eventType == "" {
		return fiber.NewError(http.StatusBadRequest, "missing event type")
	}

	rule, err := handler.store.Get(context.Background(), customerID, eventType)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "rule not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(rule)
}

// PutRule godoc
// @Summary Create or update a rule (full object)
// @Tags rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param eventType path string true "Event Type"
// @Param rule body RuleRequest true "Rule data (all fields including eventType required)"
// @Success 200 {object} domain.Rule "Updated rule"
// @Success 201 {object} domain.Rule "Created rule"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rules/{eventType} [put]
func (handler *RuleHandler) Put(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	eventType := fiberCtx.Params("eventType")
	if eventType == "" {
		return fiber.NewError(http.StatusBadRequest, "missing event type")
	}

	body, err := httpx.ParseAndValidateBody[RuleRequest](fiberCtx)
	if err != nil {
		return err
	}

	// Ensure the eventType in the path matches the eventType in the body
	if body.EventType != eventType {
		return fiber.NewError(http.StatusBadRequest, "event type in path must match event type in body")
	}

	// Check if rule exists
	_, err = handler.store.Get(context.Background(), customerID, eventType)
	exists := err == nil

	rule := domain.Rule{
		EventType:    eventType,
		CustomerID:   customerID,
		Channels:     body.Channels,
		TemplateRefs: body.TemplateRefs,
		CustomFilter: body.CustomFilter,
	}

	if err := handler.store.Put(context.Background(), rule); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the rule to get CreatedAt/UpdatedAt set by the store
	updated, err := handler.store.Get(context.Background(), customerID, eventType)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve rule: "+err.Error())
	}

	statusCode := http.StatusOK
	if !exists {
		statusCode = http.StatusCreated
	}

	return fiberCtx.Status(statusCode).JSON(updated)
}

// PatchRule godoc
// @Summary Partially update a rule
// @Tags rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param eventType path string true "Event Type"
// @Param rule body RulePatchRequest true "Rule patch data (only fields to update)"
// @Success 200 {object} domain.Rule
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rules/{eventType} [patch]
func (handler *RuleHandler) Patch(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	eventType := fiberCtx.Params("eventType")
	if eventType == "" {
		return fiber.NewError(http.StatusBadRequest, "missing event type")
	}

	// Validate patch body and get raw bytes for merge patch
	patchData, err := httpx.ValidatePatchBody[RulePatchRequest](fiberCtx)
	if err != nil {
		return err
	}

	existing, err := handler.store.Get(context.Background(), customerID, eventType)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "rule not found")
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

// DeleteRule godoc
// @Summary Delete a rule
// @Tags rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param eventType path string true "Event Type"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rules/{eventType} [delete]
func (handler *RuleHandler) Delete(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	eventType := fiberCtx.Params("eventType")
	if eventType == "" {
		return fiber.NewError(http.StatusBadRequest, "missing event type")
	}

	if err := handler.store.Delete(context.Background(), customerID, eventType); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "rule not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}

// ListRules godoc
// @Summary List rules with pagination
// @Tags rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (0-based)" default(0)
// @Param pageSize query int false "Page size" default(10)
// @Param sortBy query string false "Sort field"
// @Param sortOrder query string false "Sort order (asc/desc)" default(asc)
// @Success 200 {object} domain.ListResult[domain.Rule]
// @Failure 500 {object} map[string]string
// @Router /rules [get]
func (handler *RuleHandler) List(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	opts := httpx.ParseListOptions(fiberCtx, customerID)

	result, err := handler.store.List(context.Background(), opts)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(result)
}
