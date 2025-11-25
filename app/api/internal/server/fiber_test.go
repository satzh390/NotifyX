package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/httpx"
)

type mockValidator struct{}

func (validator *mockValidator) Validate(ctx context.Context, token string) (httpx.Claims, error) {
	return httpx.Claims{
		CustomerID: "test-customer",
		Scopes:     []string{"notify:read", "notify:write"},
	}, nil
}

func TestServer_Run(t *testing.T) {
	stores := storage.Stores{}
	validator := &mockValidator{}

	srv := New(Config{
		Addr: ":0", // Use port 0 for testing
	}, stores, validator)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := srv.Run(ctx)
	if err != context.DeadlineExceeded && err != nil {
		t.Errorf("Expected context deadline or nil, got %v", err)
	}
}

func TestServer_HealthCheck(t *testing.T) {
	stores := storage.Stores{}
	validator := &mockValidator{}

	srv := New(Config{
		Addr: ":0",
	}, stores, validator)

	// Add a simple health check route for testing
	srv.app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := srv.app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
