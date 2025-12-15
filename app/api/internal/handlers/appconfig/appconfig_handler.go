package appconfig

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

type AppConfigHandler struct {
	appStore      storage.AppConfigStore
	customerStore storage.CustomerStore
}

func NewAppConfigHandler(appStore storage.AppConfigStore, customerStore storage.CustomerStore) *AppConfigHandler {
	return &AppConfigHandler{
		appStore:      appStore,
		customerStore: customerStore,
	}
}

type APNSConfigRequest struct {
	KeyID      string `json:"keyId" validate:"required"`
	TeamID     string `json:"teamId" validate:"required"`
	BundleID   string `json:"bundleId" validate:"required"`
	KeyPath    string `json:"keyPath" validate:"required"`
	Production bool   `json:"production"`
}

type FirebaseConfigRequest struct {
	ProjectID  string `json:"projectId" validate:"required"`
	Credential string `json:"credential" validate:"required"`
}

type AppConfigRequest struct {
	// ID - required for PUT, optional for POST (will be auto-generated if not provided)
	ID string `json:"id" example:"ios-app-1"`
	// Name is required - the name of the app
	Name string `json:"name" validate:"required" example:"iOS App"`
	// Provider is required - the push provider type (apns, firebase, or mock)
	Provider domain.PushProviderType `json:"provider" validate:"required,oneof=apns firebase mock" example:"apns"`
	// APNS - APNS configuration (required if provider is "apns")
	APNS *APNSConfigRequest `json:"apns,omitempty"`
	// Firebase - Firebase configuration (required if provider is "firebase")
	Firebase *FirebaseConfigRequest `json:"firebase,omitempty"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata,omitempty"`
}

// AppConfigPatchRequest is used for PATCH requests - only mutable fields, no required validation
type AppConfigPatchRequest struct {
	// Name - optional, the name of the app
	Name string `json:"name" example:"iOS App"`
	// Provider - optional, the push provider type (apns, firebase, or mock)
	Provider domain.PushProviderType `json:"provider" example:"apns"`
	// APNS - optional APNS configuration
	APNS *APNSConfigRequest `json:"apns,omitempty"`
	// Firebase - optional Firebase configuration
	Firebase *FirebaseConfigRequest `json:"firebase,omitempty"`
	// Metadata - optional key-value pairs for additional data
	Metadata map[string]string `json:"metadata,omitempty"`
}

// getOrgID gets the organization ID from the customer context
func (handler *AppConfigHandler) getOrgID(fiberCtx *fiber.Ctx) (string, error) {
	customerID := httpx.CustomerIDFromCtx(fiberCtx)
	customer, err := handler.customerStore.Get(context.Background(), customerID)
	if err != nil {
		if err == storage.ErrNotFound {
			return "", fiber.NewError(http.StatusNotFound, "customer not found")
		}
		return "", fiber.NewError(http.StatusInternalServerError, "failed to get customer: "+err.Error())
	}
	return customer.OrgID, nil
}

// CreateAppConfig godoc
// @Summary Create a new app configuration
// @Tags app-configs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param appConfig body AppConfigRequest true "App configuration data (name and provider are required; apns/firebase config required based on provider)"
// @Success 201 {object} domain.AppConfig
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 500 {object} map[string]string
// @Router /app-configs [post]
func (handler *AppConfigHandler) Create(fiberCtx *fiber.Ctx) error {
	orgID, err := handler.getOrgID(fiberCtx)
	if err != nil {
		return err
	}

	body, err := httpx.ParseAndValidateBody[AppConfigRequest](fiberCtx)
	if err != nil {
		return err
	}

	// Validate provider-specific config
	if body.Provider == domain.PushProviderAPNS && body.APNS == nil {
		return fiber.NewError(http.StatusBadRequest, "APNS configuration is required when provider is 'apns'")
	}
	if body.Provider == domain.PushProviderFirebase && body.Firebase == nil {
		return fiber.NewError(http.StatusBadRequest, "Firebase configuration is required when provider is 'firebase'")
	}

	appID := body.ID
	if appID == "" {
		appID = uuid.NewString()
	}

	appConfig := domain.AppConfig{
		ID:        appID,
		OrgID:     orgID,
		Name:      body.Name,
		Provider:  body.Provider,
		Metadata:  body.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if body.APNS != nil {
		appConfig.APNS = &domain.APNSConfig{
			KeyID:      body.APNS.KeyID,
			TeamID:     body.APNS.TeamID,
			BundleID:   body.APNS.BundleID,
			KeyPath:    body.APNS.KeyPath,
			Production: body.APNS.Production,
		}
	}

	if body.Firebase != nil {
		appConfig.Firebase = &domain.FirebaseConfig{
			ProjectID:  body.Firebase.ProjectID,
			Credential: body.Firebase.Credential,
		}
	}

	if err := handler.appStore.Put(context.Background(), appConfig); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the created app config to get CreatedAt set by the store
	created, err := handler.appStore.Get(context.Background(), orgID, appID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve created app config: "+err.Error())
	}

	return fiberCtx.Status(http.StatusCreated).JSON(created)
}

// GetAppConfig godoc
// @Summary Get an app configuration by ID
// @Tags app-configs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "App Config ID"
// @Success 200 {object} domain.AppConfig
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /app-configs/{id} [get]
func (handler *AppConfigHandler) Get(fiberCtx *fiber.Ctx) error {
	orgID, err := handler.getOrgID(fiberCtx)
	if err != nil {
		return err
	}

	appID := fiberCtx.Params("id")
	if appID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing app config id")
	}

	appConfig, err := handler.appStore.Get(context.Background(), orgID, appID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "app config not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(appConfig)
}

// PutAppConfig godoc
// @Summary Create or update an app configuration (full object)
// @Tags app-configs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "App Config ID"
// @Param appConfig body AppConfigRequest true "App configuration data (name and provider are required)"
// @Success 200 {object} domain.AppConfig "Updated app config"
// @Success 201 {object} domain.AppConfig "Created app config"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /app-configs/{id} [put]
func (handler *AppConfigHandler) Put(fiberCtx *fiber.Ctx) error {
	orgID, err := handler.getOrgID(fiberCtx)
	if err != nil {
		return err
	}

	appID := fiberCtx.Params("id")
	if appID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing app config id")
	}

	body, err := httpx.ParseAndValidateBody[AppConfigRequest](fiberCtx)
	if err != nil {
		return err
	}

	// For PUT, id is required in the body
	if body.ID == "" {
		return fiber.NewError(http.StatusBadRequest, "id is required in request body")
	}

	// Ensure the ID in the path matches the ID in the body
	if body.ID != appID {
		return fiber.NewError(http.StatusBadRequest, "app config id in path must match app config id in body")
	}

	// Validate provider-specific config
	if body.Provider == domain.PushProviderAPNS && body.APNS == nil {
		return fiber.NewError(http.StatusBadRequest, "APNS configuration is required when provider is 'apns'")
	}
	if body.Provider == domain.PushProviderFirebase && body.Firebase == nil {
		return fiber.NewError(http.StatusBadRequest, "Firebase configuration is required when provider is 'firebase'")
	}

	// Check if app config exists
	_, err = handler.appStore.Get(context.Background(), orgID, appID)
	exists := err == nil

	appConfig := domain.AppConfig{
		ID:       appID,
		OrgID:    orgID,
		Name:     body.Name,
		Provider: body.Provider,
		Metadata: body.Metadata,
		UpdatedAt: time.Now(),
	}

	if body.APNS != nil {
		appConfig.APNS = &domain.APNSConfig{
			KeyID:      body.APNS.KeyID,
			TeamID:     body.APNS.TeamID,
			BundleID:   body.APNS.BundleID,
			KeyPath:    body.APNS.KeyPath,
			Production: body.APNS.Production,
		}
	}

	if body.Firebase != nil {
		appConfig.Firebase = &domain.FirebaseConfig{
			ProjectID:  body.Firebase.ProjectID,
			Credential: body.Firebase.Credential,
		}
	}

	if err := handler.appStore.Put(context.Background(), appConfig); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Fetch the app config to get CreatedAt/UpdatedAt set by the store
	updated, err := handler.appStore.Get(context.Background(), orgID, appID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve app config: "+err.Error())
	}

	statusCode := http.StatusOK
	if !exists {
		statusCode = http.StatusCreated
	}

	return fiberCtx.Status(statusCode).JSON(updated)
}

// PatchAppConfig godoc
// @Summary Partially update an app configuration
// @Tags app-configs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "App Config ID"
// @Param appConfig body AppConfigPatchRequest true "App configuration data (all fields optional)"
// @Success 200 {object} domain.AppConfig
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /app-configs/{id} [patch]
func (handler *AppConfigHandler) Patch(fiberCtx *fiber.Ctx) error {
	orgID, err := handler.getOrgID(fiberCtx)
	if err != nil {
		return err
	}

	appID := fiberCtx.Params("id")
	if appID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing app config id")
	}

	existing, err := handler.appStore.Get(context.Background(), orgID, appID)
	if err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "app config not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	// Validate patch body and get raw bytes for merge patch
	patchBytes, err := httpx.ValidatePatchBody[AppConfigPatchRequest](fiberCtx)
	if err != nil {
		return err
	}

	// Apply merge patch (automatically protects immutable fields)
	if err := httpx.MergePatch(&existing, patchBytes); err != nil {
		return fiber.NewError(http.StatusBadRequest, "failed to apply patch: "+err.Error())
	}

	existing.UpdatedAt = time.Now()

	if err := handler.appStore.Put(context.Background(), existing); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	updated, err := handler.appStore.Get(context.Background(), orgID, appID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "failed to retrieve updated app config: "+err.Error())
	}

	return fiberCtx.JSON(updated)
}

// DeleteAppConfig godoc
// @Summary Delete an app configuration
// @Tags app-configs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "App Config ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /app-configs/{id} [delete]
func (handler *AppConfigHandler) Delete(fiberCtx *fiber.Ctx) error {
	orgID, err := handler.getOrgID(fiberCtx)
	if err != nil {
		return err
	}

	appID := fiberCtx.Params("id")
	if appID == "" {
		return fiber.NewError(http.StatusBadRequest, "missing app config id")
	}

	if err := handler.appStore.Delete(context.Background(), orgID, appID); err != nil {
		if err == storage.ErrNotFound {
			return fiber.NewError(http.StatusNotFound, "app config not found")
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.SendStatus(http.StatusNoContent)
}

// ListAppConfigs godoc
// @Summary List app configurations
// @Tags app-configs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (0-based)" default(0)
// @Param pageSize query int false "Page size" default(20)
// @Param sortBy query string false "Sort fields (e.g., name:asc,createdAt:desc)"
// @Success 200 {object} domain.ListResult[domain.AppConfig]
// @Failure 500 {object} map[string]string
// @Router /app-configs [get]
func (handler *AppConfigHandler) List(fiberCtx *fiber.Ctx) error {
	orgID, err := handler.getOrgID(fiberCtx)
	if err != nil {
		return err
	}

	opts := httpx.ParseListOptions(fiberCtx, "")
	// Filter by orgID
	if opts.Filter == nil {
		opts.Filter = make(map[string]string)
	}
	opts.Filter["orgId"] = orgID

	result, err := handler.appStore.List(context.Background(), opts)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return fiberCtx.JSON(result)
}

