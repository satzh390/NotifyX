package subscriber

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

type SubscriberRequest struct {
	// SubscriberID - required for PUT, optional for POST (will be auto-generated if not provided)
	SubscriberID string `json:"subscriberId" example:"sub-123"`
	// Email - optional, must be valid email format if provided
	Email string `json:"email" validate:"omitempty,email" example:"user@example.com"`
	// Phone - optional phone number
	Phone string `json:"phone"`
	// PushToken - optional push notification token
	PushToken string `json:"pushToken"`
	// WebhookURL - optional, must be valid URL format if provided
	WebhookURL string `json:"webhookUrl" validate:"omitempty,url" example:"https://example.com/webhook"`
	// Groups - list of group IDs this subscriber belongs to
	Groups []string `json:"groups"`
	// SubscribedEventTypes - list of event types this subscriber opted into
	SubscribedEventTypes []string `json:"subscribedEventTypes"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata"`
	// Preferences - subscriber notification preferences
	Preferences struct {
		Channels           map[domain.ChannelType]bool `json:"channels"`
		Language           string                      `json:"language"`
		AllowedDays        []string                    `json:"allowedDays"`
		NotificationWindow struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"notificationWindow"`
	} `json:"preferences"`
}

// SubscriberPatchRequest is used for PATCH requests - only mutable fields, no required validation
type SubscriberPatchRequest struct {
	// Email - optional, must be valid email format if provided
	Email string `json:"email" validate:"omitempty,email" example:"user@example.com"`
	// Phone - optional phone number
	Phone string `json:"phone"`
	// PushToken - optional push notification token
	PushToken string `json:"pushToken"`
	// WebhookURL - optional, must be valid URL format if provided
	WebhookURL string `json:"webhookUrl" validate:"omitempty,url" example:"https://example.com/webhook"`
	// Groups - list of group IDs this subscriber belongs to
	Groups []string `json:"groups"`
	// SubscribedEventTypes - list of event types this subscriber opted into
	SubscribedEventTypes []string `json:"subscribedEventTypes"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata"`
	// Preferences - subscriber notification preferences
	Preferences struct {
		Channels           map[domain.ChannelType]bool `json:"channels"`
		Language           string                      `json:"language"`
		AllowedDays        []string                    `json:"allowedDays"`
		NotificationWindow struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"notificationWindow"`
	} `json:"preferences"`
}

// CreateSubscriber godoc
// @Summary Create a new subscriber
// @Tags subscribers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscriber body SubscriberRequest true "Subscriber data (email must be valid email format if provided, webhookUrl must be valid URL if provided)"
// @Success 201 {object} domain.Subscriber
// @Failure 400 {object} map[string]string "Bad request - validation error (e.g., invalid email format, invalid URL format)"
// @Failure 500 {object} map[string]string
// @Router /subscribers [post]
func (handler *SubscriberHandler) Create(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	body, err := httpx.ParseAndValidateBody[SubscriberRequest](fiberCtx)
	if err != nil {
		return err
	}

	subscriberID := body.SubscriberID
	if subscriberID == "" {
		subscriberID = uuid.NewString()
	}

	subscriber := domain.Subscriber{
		ID:                   subscriberID,
		CustomerID:           customerID,
		Email:                body.Email,
		Phone:                body.Phone,
		PushToken:            body.PushToken,
		WebhookURL:           body.WebhookURL,
		Groups:               body.Groups,
		SubscribedEventTypes: body.SubscribedEventTypes,
		Metadata:             body.Metadata,
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
	created, err := handler.store.Get(context.Background(), customerID, subscriberID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve created subscriber: "+err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(created)
}

// GetSubscriber godoc
// @Summary Get a subscriber by ID
// @Tags subscribers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscriber ID"
// @Success 200 {object} domain.Subscriber
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscribers/{id} [get]
func (handler *SubscriberHandler) Get(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	subscriberID := fiberCtx.Params("id")
	if subscriberID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing subscriber id")
	}

	subscriber, err := handler.store.Get(context.Background(), customerID, subscriberID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "subscriber not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(subscriber)
}

// PutSubscriber godoc
// @Summary Create or update a subscriber (full object)
// @Tags subscribers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscriber ID"
// @Param subscriber body SubscriberRequest true "Subscriber data (all fields including subscriberId required)"
// @Success 200 {object} domain.Subscriber "Updated subscriber"
// @Success 201 {object} domain.Subscriber "Created subscriber"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscribers/{id} [put]
func (handler *SubscriberHandler) Put(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	subscriberID := fiberCtx.Params("id")
	if subscriberID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing subscriber id")
	}

	body, err := httpx.ParseAndValidateBody[SubscriberRequest](fiberCtx)
	if err != nil {
		return err
	}

	// For PUT, subscriberId is required in the body
	if body.SubscriberID == "" {
		return fiber.NewError(http.StatusBadRequest, "subscriberId is required in request body")
	}

	// Ensure the ID in the path matches the ID in the body
	if body.SubscriberID != subscriberID {
		return fiber.NewError(http.StatusBadRequest, "subscriber id in path must match subscriber id in body")
	}

	// Check if subscriber exists
	_, err = handler.store.Get(context.Background(), customerID, subscriberID)
	exists := err == nil

	subscriber := domain.Subscriber{
		ID:                   subscriberID,
		CustomerID:           customerID,
		Email:                body.Email,
		Phone:                body.Phone,
		PushToken:            body.PushToken,
		WebhookURL:           body.WebhookURL,
		Groups:               body.Groups,
		SubscribedEventTypes: body.SubscribedEventTypes,
		Metadata:             body.Metadata,
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

	// Fetch the subscriber to get CreatedAt/UpdatedAt set by the store
	updated, err := handler.store.Get(context.Background(), customerID, subscriberID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve subscriber: "+err.Error())
	}

	statusCode := http.StatusOK
	if !exists {
		statusCode = http.StatusCreated
	}

	return fiberCtx.Status(statusCode).JSON(updated)
}

// PatchSubscriber godoc
// @Summary Partially update a subscriber
// @Tags subscribers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscriber ID"
// @Param subscriber body SubscriberPatchRequest true "Subscriber patch data (only fields to update)"
// @Success 200 {object} domain.Subscriber
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscribers/{id} [patch]
func (handler *SubscriberHandler) Patch(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	subscriberID := fiberCtx.Params("id")
	if subscriberID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing subscriber id")
	}

	// Validate patch body and get raw bytes for merge patch
	patchData, err := httpx.ValidatePatchBody[SubscriberPatchRequest](fiberCtx)
	if err != nil {
		return err
	}

	existing, err := handler.store.Get(context.Background(), customerID, subscriberID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "subscriber not found")
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

// DeleteSubscriber godoc
// @Summary Delete a subscriber
// @Tags subscribers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscriber ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscribers/{id} [delete]
func (handler *SubscriberHandler) Delete(fiberCtx *fiber.Ctx) error {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	subscriberID := fiberCtx.Params("id")
	if subscriberID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing subscriber id")
	}

	if err := handler.store.Delete(context.Background(), customerID, subscriberID); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "subscriber not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}

// ListSubscribers godoc
// @Summary List subscribers with pagination
// @Tags subscribers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (0-based)" default(0)
// @Param pageSize query int false "Page size" default(10)
// @Param sortBy query string false "Sort field"
// @Param sortOrder query string false "Sort order (asc/desc)" default(asc)
// @Success 200 {object} domain.ListResult[domain.Subscriber]
// @Failure 500 {object} map[string]string
// @Router /subscribers [get]
func (handler *SubscriberHandler) List(fiberCtx *fiber.Ctx) error {
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
