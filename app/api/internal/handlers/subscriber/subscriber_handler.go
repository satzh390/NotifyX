package subscriber

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

type SubscriberHandler struct {
	store storage.SubscriberStore
}

func NewSubscriberHandler(store storage.SubscriberStore) *SubscriberHandler {
	return &SubscriberHandler{store: store}
}

type subscriberRequest struct {
	SubscriberID string            `json:"subscriberId"`
	Email        string            `json:"email"`
	Phone        string            `json:"phone"`
	PushToken    string            `json:"pushToken"`
	WebhookURL   string            `json:"webhookUrl"`
	Groups       []string          `json:"groups"`
	Metadata     map[string]string `json:"metadata"`
	Preferences  struct {
		Channels           map[domain.ChannelType]bool `json:"channels"`
		Language           string                      `json:"language"`
		AllowedDays        []string                    `json:"allowedDays"`
		NotificationWindow struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"notificationWindow"`
	} `json:"preferences"`
}

func (handler *SubscriberHandler) Create(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	var body subscriberRequest
	if err := fiberCtx.BodyParser(&body); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid body: "+err.Error())
	}

	subscriberID := body.SubscriberID
	if subscriberID == "" {
		subscriberID = uuid.NewString()
	}

	subscriber := domain.Subscriber{
		ID:         subscriberID,
		OrgID:      orgID,
		Email:      body.Email,
		Phone:      body.Phone,
		PushToken:  body.PushToken,
		WebhookURL: body.WebhookURL,
		Groups:     body.Groups,
		Metadata:   body.Metadata,
		Preferences: domain.SubscriberPrefs{
			Channels:    body.Preferences.Channels,
			Language:    body.Preferences.Language,
			AllowedDays: body.Preferences.AllowedDays,
			NotificationWindow: domain.TimeWindow{
				Start: body.Preferences.NotificationWindow.Start,
				End:   body.Preferences.NotificationWindow.End,
			},
		},
	}

	if err := handler.store.Put(context.Background(), subscriber); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the created subscriber to get CreatedAt set by the store
	created, err := handler.store.Get(context.Background(), orgID, subscriberID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve created subscriber: "+err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(created)
}

func (handler *SubscriberHandler) Get(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	subscriberID := fiberCtx.Params("id")
	if subscriberID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing subscriber id")
	}

	subscriber, err := handler.store.Get(context.Background(), orgID, subscriberID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "subscriber not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(subscriber)
}

func (handler *SubscriberHandler) Update(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	subscriberID := fiberCtx.Params("id")
	if subscriberID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing subscriber id")
	}

	existing, err := handler.store.Get(context.Background(), orgID, subscriberID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "subscriber not found")
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

func (handler *SubscriberHandler) Delete(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	subscriberID := fiberCtx.Params("id")
	if subscriberID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing subscriber id")
	}

	if err := handler.store.Delete(context.Background(), orgID, subscriberID); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "subscriber not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}

func (handler *SubscriberHandler) List(fiberCtx *fiber.Ctx) error {
	orgID := middlewares.OrgIDFromCtx(fiberCtx)
	opts := httpx.ParseListOptions(fiberCtx, orgID)

	result, err := handler.store.List(context.Background(), opts)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(result)
}
