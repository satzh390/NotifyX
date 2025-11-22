package group

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notifyx/api/internal/middlewares"
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

type groupRequest struct {
	GroupID     string            `json:"groupId"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Subscribers []string          `json:"subscribers"`
	Metadata    map[string]string `json:"metadata"`
}

func (handler *GroupHandler) Create(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	var body groupRequest
	if err := fiberCtx.BodyParser(&body); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid body: "+err.Error())
	}

	groupID := body.GroupID
	if groupID == "" {
		groupID = uuid.NewString()
	}

	group := domain.Group{
		ID:          groupID,
		OrgID:       orgID,
		Name:        body.Name,
		Description: body.Description,
		Subscribers: body.Subscribers,
		Metadata:    body.Metadata,
	}

	if err := handler.store.Put(context.Background(), group); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(group)
}

func (handler *GroupHandler) Get(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	groupID := fiberCtx.Params("id")
	if groupID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing group id")
	}

	group, err := handler.store.Get(context.Background(), orgID, groupID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "group not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(group)
}

func (handler *GroupHandler) Update(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	groupID := fiberCtx.Params("id")
	if groupID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing group id")
	}

	existing, err := handler.store.Get(context.Background(), orgID, groupID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "group not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Apply merge patch
	patchData := fiberCtx.Body()
	if err := httpx.MergePatch(&existing, patchData); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid patch: "+err.Error())
	}

	if err := handler.store.Put(context.Background(), existing); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(existing)
}

func (handler *GroupHandler) Delete(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	groupID := fiberCtx.Params("id")
	if groupID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing group id")
	}

	if err := handler.store.Delete(context.Background(), orgID, groupID); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "group not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}

func (handler *GroupHandler) List(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	opts := httpx.ParseListOptions(fiberCtx, orgID)

	result, err := handler.store.List(context.Background(), opts)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(result)
}
