//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/api/internal/handlers/group"
	mongoadapter "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/core/storage"
)

func setupIntegrationTest(t *testing.T) (storage.Stores, *fiber.App, func()) {
	// Connect to test MongoDB (should be running via docker-compose)
	ctx := context.Background()
	stores, cleanup, err := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      "mongodb://localhost:27017",
		Database: "notifyx_test",
	})
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	app := fiber.New()
	api := app.Group("/api/v1")

	// Mock middleware that sets customerID
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("customerId", "test-customer-integration")
		return c.Next()
	})

	return stores, app, func() {
		if cleanup != nil {
			_ = cleanup(ctx)
		}
	}
}

func setupGroupIntegrationApp(t *testing.T) (*fiber.App, func()) {
	stores, app, cleanup := setupIntegrationTest(t)
	api := app.Group("/api/v1")

	groupHandler := group.NewGroupHandler(stores.Groups)
	groups := api.Group("/groups")
	groups.Post("", groupHandler.Create)
	groups.Get("/:id", groupHandler.Get)
	groups.Put("/:id", groupHandler.Put)
	groups.Patch("/:id", groupHandler.Patch)
	groups.Delete("/:id", groupHandler.Delete)
	groups.Get("", groupHandler.List)

	return app, cleanup
}
