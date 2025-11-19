package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/api/internal/auth"
)

type mockValidator struct {
	claims auth.Claims
	err    error
}

func (m *mockValidator) Validate(ctx context.Context, token string) (auth.Claims, error) {
	if m.err != nil {
		return auth.Claims{}, m.err
	}
	return m.claims, nil
}

func TestRequireAuth(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{
		claims: auth.Claims{
			OrgID:  "test-org",
			Scopes: []string{"notify:read", "notify:write"},
		},
	}

	app.Get("/test", RequireAuth(validator, "notify:read"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_MissingToken(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{}

	app.Get("/test", RequireAuth(validator, "notify:read"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestOrgIDFromCtx(t *testing.T) {
	app := fiber.New()
	validator := &mockValidator{
		claims: auth.Claims{
			OrgID:  "test-org",
			Scopes: []string{"notify:read"},
		},
	}

	app.Get("/test", RequireAuth(validator, "notify:read"), func(c *fiber.Ctx) error {
		orgID := OrgIDFromCtx(c)
		return c.SendString(orgID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
