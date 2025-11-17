package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notifyx/api/internal/middlewares"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/core/utils"
)

type TemplateHandler struct {
	store storage.TemplateStore
}

func NewTemplateHandler(store storage.TemplateStore) *TemplateHandler {
	return &TemplateHandler{store: store}
}

type templateRequest struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Channel   domain.ChannelType `json:"channel"`
	Version   int               `json:"version"`
	Body      string            `json:"body"`
	Subject   string            `json:"subject"`
	Metadata  map[string]string `json:"metadata"`
}

func (handler *TemplateHandler) Create(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	var body templateRequest
	if err := fiberCtx.BodyParser(&body); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid body: "+err.Error())
	}

	if body.Name == "" {
		return fiber.NewError(http.StatusBadRequest, "name is required")
	}
	if body.Channel == "" {
		return fiber.NewError(http.StatusBadRequest, "channel is required")
	}
	if body.Body == "" {
		return fiber.NewError(http.StatusBadRequest, "body is required")
	}

	templateID := body.ID
	if templateID == "" {
		templateID = uuid.NewString()
	}

	now := time.Now()
	template := domain.Template{
		ID:        templateID,
		OrgID:     orgID,
		Name:      body.Name,
		Channel:   body.Channel,
		Version:   body.Version,
		Body:      body.Body,
		Subject:   body.Subject,
		Metadata:  body.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := handler.store.Put(context.Background(), template); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(template)
}

func (handler *TemplateHandler) Get(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	templateID := fiberCtx.Params("id")
	if templateID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing template id")
	}

	template, err := handler.store.Get(context.Background(), orgID, templateID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "template not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(template)
}

func (handler *TemplateHandler) Update(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
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

	// Apply merge patch
	patchData := fiberCtx.Body()
	if err := utils.MergePatch(&existing, patchData); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid patch: "+err.Error())
	}

	existing.UpdatedAt = time.Now()

	if err := handler.store.Put(context.Background(), existing); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(existing)
}

func (handler *TemplateHandler) Delete(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
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

