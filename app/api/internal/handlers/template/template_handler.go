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
	// ID - optional template ID, will be auto-generated if not provided
	ID string `json:"id"`
	// Name is required - the name of the template
	Name string `json:"name" validate:"required" example:"Order Confirmation"`
	// Channel is required - the notification channel type (email, sms, push, webhook)
	Channel domain.ChannelType `json:"channel" validate:"required" example:"email"`
	// Version - template version number
	Version int `json:"version"`
	// Content - template content (validation depends on channel type)
	// For email: body is required
	// For SMS: body is required
	// For push: body or title is required
	// For webhook: payload or body is required
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
	orgID := httpx.OrgIDFromCtx(fiberCtx)
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
		OrgID:        orgID,
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
	case domain.ChannelWebhook:
		if content.Payload == nil && content.Body == "" {
			return errors.New("payload or body is required for webhook templates")
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
	orgID := httpx.OrgIDFromCtx(fiberCtx)
	templateID := fiberCtx.Params("id")
	if templateID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing template id")
	}

	// Check if language query parameter is provided
	language := fiberCtx.Query("language")

	var template domain.Template
	var err error
	if language != "" {
		template, err = handler.store.GetByLanguage(context.Background(), orgID, templateID, language)
	} else {
		template, err = handler.store.Get(context.Background(), orgID, templateID)
	}

	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "template not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(template)
}

// UpdateTemplate godoc
// @Summary Update a template (merge patch)
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Param template body object true "Template patch data"
// @Success 200 {object} domain.Template
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id} [put]
func (handler *TemplateHandler) Update(fiberCtx *fiber.Ctx) error {
	orgID := httpx.OrgIDFromCtx(fiberCtx)
	templateID := fiberCtx.Params("id")
	if templateID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing template id")
	}

	existing, err := handler.store.Get(context.Background(), orgID, templateID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "template not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Validate patch body and get raw bytes for merge patch
	patchData, err := httpx.ValidatePatchBody[TemplateRequest](fiberCtx)
	if err != nil {
		return err
	}

	// Apply merge patch
	if err := httpx.MergePatch(&existing, patchData); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid patch: "+err.Error())
	}

	// Validate content after merge (channel or content might have been updated)
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
	orgID := httpx.OrgIDFromCtx(fiberCtx)
	templateID := fiberCtx.Params("id")
	if templateID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing template id")
	}

	if err := handler.store.Delete(context.Background(), orgID, templateID); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "template not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}
