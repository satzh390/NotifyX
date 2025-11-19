package server

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/api/internal/auth"
	"github.com/notifyx/api/internal/routes"
	"github.com/notifyx/core/storage"
)

type Config struct {
	Addr string
}

type Server struct {
	cfg       Config
	app       *fiber.App
	stores    storage.Stores
	validator auth.Validator
}

func New(cfg Config, stores storage.Stores, validator auth.Validator) *Server {
	app := fiber.New()
	srv := &Server{
		cfg:       cfg,
		app:       app,
		stores:    stores,
		validator: validator,
	}

	routes.RegisterRoutes(srv.app, srv.stores, srv.validator)
	return srv
}

func (server *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.app.Listen(server.cfg.Addr)
	}()

	select {
	case <-ctx.Done():
		_ = server.app.Shutdown()
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("fiber listen: %w", err)
		}
		return nil
	}
}
