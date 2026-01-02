package template

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/httpx"
)

type TemplateHandler struct {
	store storage.TemplateStore
}

func NewTemplateHandler(store storage.TemplateStore) *TemplateHandler {
	return &TemplateHandler{store: store}
}

type TemplateRequest struct {
	// ID - required for PUT, optional for POST (will be auto-generated if not provided)
	ID string `json:"id" example:"template-123"`
	// Name is required - the name of the template
	Name string `json:"name" validate:"required" example:"Order Confirmation"`
	// Channel is required - the notification channel type (email, sms, push)
	Channel domain.ChannelType `json:"channel" validate:"required" example:"email"`
	// Version - template version number
	Version int `json:"version"`
	// Content - template content (validation depends on channel type)
	// For email: body is required
	// For SMS: body is required
	// For push: body or title is required
	Content domain.TemplateContent `json:"content"`
	// Translations - optional translations by language code
	Translations map[string]domain.TemplateContent `json:"translations,omitempty"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata,omitempty"`
}

// TemplatePatchRequest is used for PATCH requests - only mutable fields, no required validation
type TemplatePatchRequest struct {
	// Name - optional, the name of the template
	Name string `json:"name" example:"Order Confirmation"`
	// Version - template version number
	Version int `json:"version"`
	// Content - template content (validation depends on channel type)
	Content domain.TemplateContent `json:"content"`
	// Translations - optional translations by language code
	Translations map[string]domain.TemplateContent `json:"translations,omitempty"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CreateTemplate godoc
// @Summary Create a new template
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param template body TemplateRequest true "Template data (name and channel are required; content validation depends on channel type)"
// @Success 201 {object} domain.Template
// @Failure 400 {object} map[string]string "Bad request - validation error (e.g., name is required, channel is required, or content validation failed based on channel type)"
// @Failure 500 {object} map[string]string
// @Router /templates [post]
func (handler *TemplateHandler) Create(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	body, err := httpx.ParseAndValidateBody[TemplateRequest](fiberCtx)
	if err != nil {
		return err
	}

	// Validate content based on channel
	if err := validateTemplateContent(body.Channel, body.Content); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	templateID := body.ID
	if templateID == "" {
		templateID = uuid.NewString()
	}

	now := time.Now()
	template := domain.Template{
		ID:           templateID,
		CustomerID:   customerID,
		Name:         body.Name,
		Channel:      body.Channel,
		Version:      body.Version,
		Content:      body.Content,
		Translations: body.Translations,
		Metadata:     body.Metadata,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := handler.store.Put(context.Background(), template); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(template)
}

// validateTemplateContent validates content based on channel type
func validateTemplateContent(channel domain.ChannelType, content domain.TemplateContent) error {
	switch channel {
	case domain.ChannelEmail:
		if content.Body == "" {
			return errors.New("body is required for email templates")
		}
		// Subject is optional for email
	case domain.ChannelSMS:
		if content.Body == "" {
			return errors.New("body is required for SMS templates")
		}
	case domain.ChannelPush:
		if content.Body == "" && content.Title == "" {
			return errors.New("body or title is required for push templates")
		}
	default:
		return errors.New("unsupported channel type")
	}
	return nil
}

// GetTemplate godoc
// @Summary Get a template by ID
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Param language query string false "Language code for translation"
// @Success 200 {object} domain.Template
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id} [get]
func (handler *TemplateHandler) Get(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	templateID := fiberCtx.Params("id")
	if templateID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing template id")
	}

	// Check if language query parameter is provided
	language := fiberCtx.Query("language")

	var template domain.Template
	var err error
	if language != "" {
		template, err = handler.store.GetByLanguage(context.Background(), customerID, templateID, language)
	} else {
		template, err = handler.store.Get(context.Background(), customerID, templateID)
	}

	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "template not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(template)
}

// PutTemplate godoc
// @Summary Create or update a template (full object)
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Param template body TemplateRequest true "Template data (all fields including id, name, and channel required)"
// @Success 200 {object} domain.Template "Updated template"
// @Success 201 {object} domain.Template "Created template"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id} [put]
func (handler *TemplateHandler) Put(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	templateID := fiberCtx.Params("id")
	if templateID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing template id")
	}

	body, err := httpx.ParseAndValidateBody[TemplateRequest](fiberCtx)
	if err != nil {
		return err
	}

	// For PUT, id is required in the body
	if body.ID == "" {
		return fiber.NewError(http.StatusBadRequest, "id is required in request body")
	}

	// Ensure the ID in the path matches the ID in the body
	if body.ID != templateID {
		return fiber.NewError(http.StatusBadRequest, "template id in path must match template id in body")
	}

	// Validate content based on channel
	if err := validateTemplateContent(body.Channel, body.Content); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	// Check if template exists
	_, err = handler.store.Get(context.Background(), customerID, templateID)
	exists := err == nil

	template := domain.Template{
		ID:           templateID,
		CustomerID:   customerID,
		Name:         body.Name,
		Channel:      body.Channel,
		Version:      body.Version,
		Content:      body.Content,
		Translations: body.Translations,
		Metadata:     body.Metadata,
	}

	if err := handler.store.Put(context.Background(), template); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the template to get CreatedAt/UpdatedAt set by the store
	updated, err := handler.store.Get(context.Background(), customerID, templateID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve template: "+err.Error())
	}

	statusCode := http.StatusOK
	if !exists {
		statusCode = http.StatusCreated
	}

	return fiberCtx.Status(statusCode).JSON(updated)
}

// PatchTemplate godoc
// @Summary Partially update a template
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Param template body TemplatePatchRequest true "Template patch data (only fields to update)"
// @Success 200 {object} domain.Template
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id} [patch]
func (handler *TemplateHandler) Patch(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	templateID := fiberCtx.Params("id")
	if templateID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing template id")
	}

	// Validate patch body and get raw bytes for merge patch
	patchData, err := httpx.ValidatePatchBody[TemplatePatchRequest](fiberCtx)
	if err != nil {
		return err
	}

	existing, err := handler.store.Get(context.Background(), customerID, templateID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "template not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Apply merge patch
	if err := httpx.MergePatch(&existing, patchData); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid patch: "+err.Error())
	}

	// Validate content after merge (content might have been updated)
	if err := validateTemplateContent(existing.Channel, existing.Content); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	existing.UpdatedAt = time.Now()

	if err := handler.store.Put(context.Background(), existing); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(existing)
}

// DeleteTemplate godoc
// @Summary Delete a template
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id} [delete]
func (handler *TemplateHandler) Delete(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	templateID := fiberCtx.Params("id")
	if templateID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing template id")
	}

	if err := handler.store.Delete(context.Background(), customerID, templateID); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "template not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}
