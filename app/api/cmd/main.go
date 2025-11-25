// @title NotifyX API
// @version 1.0
// @description API for organizations, customers, templates, groups, subscribers, and rules
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	stores, cleanup, err := mongoadapter.NewStoreSet(ctx, mongoadapter.Options{
		URI:      cfg.Storage.Mongo.URI,
		Database: cfg.Storage.Mongo.Database,
	})
	if err != nil {
		logger.Error("failed to initialize mongo store", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if cleanup != nil {
			_ = cleanup(context.Background())
		}
	}()

	validator, err := httpx.NewJWKSValidator(ctx, cfg.OAuth.Issuer, cfg.OAuth.JWKS, cfg.OAuth.Audiences)
	if err != nil {
		logger.Error("failed to initialize auth validator", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv := server.New(server.Config{
		Addr: cfg.HTTP.Addr,
	}, stores, validator)

	logger.Info("api started", slog.String("addr", cfg.HTTP.Addr))

	if err := srv.Run(ctx); err != nil {
		logger.Error("api exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("api exited cleanly")
}
