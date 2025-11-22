package rule

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/api/internal/middlewares"
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

type ruleRequest struct {
	EventType         string                        `json:"eventType"`
	Channels          []domain.ChannelType          `json:"channels"`
	DefaultRecipients domain.Recipients             `json:"defaultRecipients"`
	TemplateRefs      map[domain.ChannelType]string `json:"templateRefs"`
}

func (handler *RuleHandler) Create(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	var body ruleRequest
	if err := fiberCtx.BodyParser(&body); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid body: "+err.Error())
	}

	if body.EventType == "" {
		return fiber.NewError(http.StatusBadRequest, "eventType is required")
	}

	rule := domain.Rule{
		EventType:         body.EventType,
		OrgID:             orgID,
		Channels:          body.Channels,
		DefaultRecipients: body.DefaultRecipients,
		TemplateRefs:      body.TemplateRefs,
	}

	if err := handler.store.Put(context.Background(), rule); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the created rule to get CreatedAt set by the store
	created, err := handler.store.Get(context.Background(), orgID, rule.EventType)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve created rule: "+err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(created)
}

func (handler *RuleHandler) Get(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	eventType := fiberCtx.Params("eventType")
	if eventType == "" {
		return fiber.NewError(http.StatusBadRequest, "missing event type")
	}

	rule, err := handler.store.Get(context.Background(), orgID, eventType)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "rule not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(rule)
}

func (handler *RuleHandler) Update(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	eventType := fiberCtx.Params("eventType")
	if eventType == "" {
		return fiber.NewError(http.StatusBadRequest, "missing event type")
	}

	existing, err := handler.store.Get(context.Background(), orgID, eventType)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "rule not found")
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

func (handler *RuleHandler) Delete(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	eventType := fiberCtx.Params("eventType")
	if eventType == "" {
		return fiber.NewError(http.StatusBadRequest, "missing event type")
	}

	if err := handler.store.Delete(context.Background(), orgID, eventType); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "rule not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}

func (handler *RuleHandler) List(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	opts := httpx.ParseListOptions(fiberCtx, orgID)

	result, err := handler.store.List(context.Background(), opts)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(result)
}
