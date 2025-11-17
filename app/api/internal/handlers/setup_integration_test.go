//go:build integration
// +build integration

package handlers

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
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

	// Mock middleware that sets orgID
	api.Use(func(c *fiber.Ctx) error {
		c.Locals("orgId", "test-org-integration")
		return c.Next()
	})

	return stores, app, func() {
		if cleanup != nil {
			_ = cleanup(ctx)
		}
	}
}

func setupSubscriberIntegrationApp(t *testing.T) (*fiber.App, func()) {
	stores, app, cleanup := setupIntegrationTest(t)
	api := app.Group("/api/v1")

	subHandler := NewSubscriberHandler(stores.Subscribers)
	subscribers := api.Group("/subscribers")
	subscribers.Post("", subHandler.Create)
	subscribers.Get("/:id", subHandler.Get)
	subscribers.Put("/:id", subHandler.Update)
	subscribers.Delete("/:id", subHandler.Delete)
	subscribers.Get("", subHandler.List)

	return app, cleanup
}

func setupGroupIntegrationApp(t *testing.T) (*fiber.App, func()) {
	stores, app, cleanup := setupIntegrationTest(t)
	api := app.Group("/api/v1")

	groupHandler := NewGroupHandler(stores.Groups)
	groups := api.Group("/groups")
	groups.Post("", groupHandler.Create)
	groups.Get("/:id", groupHandler.Get)
	groups.Put("/:id", groupHandler.Update)
	groups.Delete("/:id", groupHandler.Delete)
	groups.Get("", groupHandler.List)

	return app, cleanup
}

func setupRuleIntegrationApp(t *testing.T) (*fiber.App, func()) {
	stores, app, cleanup := setupIntegrationTest(t)
	api := app.Group("/api/v1")

	ruleHandler := NewRuleHandler(stores.Rules)
	rules := api.Group("/rules")
	rules.Post("", ruleHandler.Create)
	rules.Get("/:eventType", ruleHandler.Get)
	rules.Put("/:eventType", ruleHandler.Update)
	rules.Delete("/:eventType", ruleHandler.Delete)
	rules.Get("", ruleHandler.List)

	return app, cleanup
}


