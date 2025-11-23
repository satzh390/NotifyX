package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ParseAndValidateBody parses and validates the request body
// Returns the parsed body or an error
func ParseAndValidateBody[T any](fiberCtx *fiber.Ctx) (*T, error) {
	var body T

	// Parse body
	if err := fiberCtx.BodyParser(&body); err != nil {
		return nil, fiber.NewError(http.StatusBadRequest, "invalid body: "+err.Error())
	}

	// Validate struct
	if err := validate.Struct(&body); err != nil {
		validationErrors := formatValidationErrors(err)
		return nil, fiber.NewError(http.StatusBadRequest, validationErrors)
	}

	return &body, nil
}

// ValidatePatchBody validates the patch body from raw bytes and returns the raw bytes for merge patch
// This allows validation before applying merge patch while preserving the original JSON structure
// PatchRequest types should not have required validation tags - it's the developer's responsibility
func ValidatePatchBody[T any](fiberCtx *fiber.Ctx) ([]byte, error) {
	// Get raw body bytes first (before parsing)
	patchData := fiberCtx.Body()
	if len(patchData) == 0 {
		return nil, fiber.NewError(http.StatusBadRequest, "patch data is required")
	}

	// Parse and validate the body from raw bytes
	var body T
	if err := json.Unmarshal(patchData, &body); err != nil {
		return nil, fiber.NewError(http.StatusBadRequest, "invalid patch body: "+err.Error())
	}

	// Validate struct (PatchRequest types don't have required tags, so this is safe)
	if err := validate.Struct(&body); err != nil {
		validationErrors := formatValidationErrors(err)
		return nil, fiber.NewError(http.StatusBadRequest, validationErrors)
	}

	// Return a copy of the raw bytes for merge patch to avoid issues with Fiber's body buffer
	patchDataCopy := make([]byte, len(patchData))
	copy(patchDataCopy, patchData)
	return patchDataCopy, nil
}

// formatValidationErrors formats validator errors into a readable string
func formatValidationErrors(err error) string {
	var messages []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			field := fieldError.Field()
			reason := getValidationReason(fieldError.Tag(), fieldError.Param())
			message := fmt.Sprintf("%s is invalid, %s", field, reason)
			messages = append(messages, message)
		}
	} else {
		messages = append(messages, err.Error())
	}

	return strings.Join(messages, "; ")
}

// getValidationReason returns a human-readable reason for a validation failure
func getValidationReason(tag, param string) string {
	// Map of validation tags to their reason messages
	reasonMap := map[string]string{
		"required": "required",
		"email":    "must be a valid email address",
		"url":      "must be a valid URL",
		"min":      fmt.Sprintf("must be at least %s", param),
		"max":      fmt.Sprintf("must be at most %s", param),
		"gte":      fmt.Sprintf("must be greater than or equal to %s", param),
		"lte":      fmt.Sprintf("must be less than or equal to %s", param),
		"oneof":    fmt.Sprintf("must be one of: %s", param),
		"gt":       fmt.Sprintf("must be greater than %s", param),
		"lt":       fmt.Sprintf("must be less than %s", param),
		"eq":       fmt.Sprintf("must be equal to %s", param),
		"ne":       fmt.Sprintf("must not be equal to %s", param),
		"len":      fmt.Sprintf("must have length %s", param),
		"alpha":    "must contain only alphabetic characters",
		"alphanum": "must contain only alphanumeric characters",
		"numeric":  "must be numeric",
		"uuid":     "must be a valid UUID",
		"ip":       "must be a valid IP address",
		"ipv4":     "must be a valid IPv4 address",
		"ipv6":     "must be a valid IPv6 address",
	}

	if reason, exists := reasonMap[tag]; exists {
		return reason
	}

	// Generic fallback for unknown tags
	if param != "" {
		return fmt.Sprintf("failed validation rule '%s' with parameter '%s'", tag, param)
	}
	return fmt.Sprintf("failed validation rule '%s'", tag)
}
