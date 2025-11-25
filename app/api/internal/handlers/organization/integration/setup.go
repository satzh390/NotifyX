//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/api/internal/handlers/organization"
	mongoadapter "github.com/notifyx/core/adapters/mongo"
)

func setupOrganizationIntegrationApp(t *testing.T) (*fiber.App, func()) {
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

	organizationHandler := organization.NewOrganizationHandler(stores.Organizations)
	organizations := api.Group("/organizations")
	organizations.Post("", organizationHandler.Create)
	organizations.Get("/:id", organizationHandler.Get)
	organizations.Put("/:id", organizationHandler.Put)
	organizations.Patch("/:id", organizationHandler.Patch)
	organizations.Delete("/:id", organizationHandler.Delete)
	organizations.Get("", organizationHandler.List)

	return app, func() {
		if cleanup != nil {
			_ = cleanup(ctx)
		}
	}
}

