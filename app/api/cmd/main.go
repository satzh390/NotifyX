// @title NotifyX API
// @version 1.0
// @description API for templates, groups, subscribers, and rules
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"log"
	"os"

	"github.com/notifyx/api/config"
	"github.com/notifyx/api/internal/server"
	mongoadapter "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/httpx"
)

func main() {
	configPath := os.Getenv("NOTIFYX_API_CONFIG")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	stores, cleanup, err := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      cfg.Storage.Mongo.URI,
		Database: cfg.Storage.Mongo.Database,
	})
	if err != nil {
		log.Fatalf("failed to initialize mongo store: %v", err)
	}
	defer func() {
		if cleanup != nil {
			_ = cleanup(context.Background())
		}
	}()

	validator, err := httpx.NewJWKSValidator(ctx, cfg.OAuth.Issuer, cfg.OAuth.JWKS, cfg.OAuth.Audiences)
	if err != nil {
		log.Fatalf("failed to initialize auth validator: %v", err)
	}

	srv := server.New(server.Config{
		Addr: cfg.HTTP.Addr,
	}, stores, validator)

	if err := srv.Run(ctx); err != nil {
		log.Fatalf("notifyx-api exited: %v", err)
	}
}
