package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type mockValidator struct {
	claims Claims
	err    error
}

func (validator *mockValidator) Validate(ctx context.Context, token string) (Claims, error) {
	if validator.err != nil {
		return Claims{}, validator.err
	}
	return validator.claims, nil
}

func TestRequireAuth(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{
		claims: Claims{
			CustomerID: "test-customer",
			Scopes:    []string{"notify:read", "notify:write"},
		},
	}

	app.Get("/test", RequireAuth(validator, "notify:read"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireAuth_MissingToken(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{}

	app.Get("/test", RequireAuth(validator, "notify:read"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{
		err: ErrInvalidToken,
	}

	app.Get("/test", RequireAuth(validator, "notify:read"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAuth_InsufficientScope(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{
		claims: Claims{
			CustomerID: "test-customer",
			Scopes:    []string{"notify:read"},
		},
	}

	app.Get("/test", RequireAuth(validator, "notify:write"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRequireAuth_NoScopesRequired(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{
		claims: Claims{
			CustomerID: "test-customer",
			Scopes:    []string{"notify:read"},
		},
	}

	app.Get("/test", RequireAuth(validator), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCustomerIDFromCtx(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{
		claims: Claims{
			CustomerID: "test-customer",
			Scopes:    []string{"notify:read"},
		},
	}

	app.Get("/test", RequireAuth(validator, "notify:read"), func(c *fiber.Ctx) error {
		customerID := CustomerIDFromCtx(c)
		return c.SendString(customerID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireAuth_InvalidBearerFormat(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{
		claims: Claims{
			CustomerID: "test-customer",
			Scopes:    []string{"notify:read"},
		},
	}

	app.Get("/test", RequireAuth(validator, "notify:read"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "missing header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "case insensitive bearer",
			authHeader:     "bearer test-token",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestHasScopes(t *testing.T) {
	tests := []struct {
		name     string
		claims   Claims
		required []string
		expected bool
	}{
		{
			name: "all required scopes present",
			claims: Claims{
				Scopes: []string{"notify:read", "notify:write"},
			},
			required: []string{"notify:read"},
			expected: true,
		},
		{
			name: "multiple required scopes all present",
			claims: Claims{
				Scopes: []string{"notify:read", "notify:write", "notify:admin"},
			},
			required: []string{"notify:read", "notify:write"},
			expected: true,
		},
		{
			name: "missing required scope",
			claims: Claims{
				Scopes: []string{"notify:read"},
			},
			required: []string{"notify:write"},
			expected: false,
		},
		{
			name: "no required scopes",
			claims: Claims{
				Scopes: []string{"notify:read"},
			},
			required: []string{},
			expected: true,
		},
		{
			name: "empty claims scopes",
			claims: Claims{
				Scopes: []string{},
			},
			required: []string{"notify:read"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasScopes(tt.claims, tt.required...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequireAuth_ValidatorError(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{
		err: errors.New("validation error"),
	}

	app.Get("/test", RequireAuth(validator, "notify:read"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}
