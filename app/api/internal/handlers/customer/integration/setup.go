//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/api/internal/handlers/customer"
	mongoadapter "github.com/notifyx/core/adapters/mongo"
)

func setupCustomerIntegrationApp(t *testing.T) (*fiber.App, func()) {
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

	customerHandler := customer.NewCustomerHandler(stores.Customers)
	customers := api.Group("/customers")
	customers.Post("", customerHandler.Create)
	customers.Get("/:id", customerHandler.Get)
	customers.Put("/:id", customerHandler.Put)
	customers.Patch("/:id", customerHandler.Patch)
	customers.Delete("/:id", customerHandler.Delete)
	customers.Get("", customerHandler.List)

	return app, func() {
		if cleanup != nil {
			_ = cleanup(ctx)
		}
	}
}
