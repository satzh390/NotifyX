package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"github.com/notifyx/api/internal/routes"
	"github.com/notifyx/core/storage"
	"github.com/notifyx/httpx"

	_ "github.com/notifyx/api/docs" // swagger docs
)

type Config struct {
	Addr string
}

type Server struct {
	cfg       Config
	app       *fiber.App
	stores    storage.Stores
	validator httpx.AuthValidator
}

func New(cfg Config, stores storage.Stores, validator httpx.AuthValidator) *Server {
	// Create Fiber app with custom config
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	srv := &Server{
		cfg:       cfg,
		app:       app,
		stores:    stores,
		validator: validator,
	}

	app.Use(requestid.New(requestid.Config{
		Header: "X-Request-ID",
		Generator: func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		},
	}))
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			// Log stack trace to stderr
			slog.Error("panic recovered", slog.Any("error", e))
		},
	}))
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path} | ${ip} | ${locals:requestid}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
		Output:     os.Stdout,
		// Skip logging for health checks and swagger
		Next: func(c *fiber.Ctx) bool {
			path := c.Path()
			return path == "/swagger" || path == "/swagger/*" || path == "/health"
		},
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "notifyx-api",
		})
	})
	app.Get("/swagger/*", swagger.HandlerDefault)
	routes.RegisterRoutes(srv.app, srv.stores, srv.validator)
	return srv
}

// customErrorHandler handles errors and logs them appropriately
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Log error details
	slog.Error("request error",
		slog.String("method", c.Method()),
		slog.String("path", c.Path()),
		slog.Int("status", code),
		slog.String("error", err.Error()),
	)

	return c.Status(code).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
}

func (server *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.app.Listen(server.cfg.Addr)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutting down api server")
		_ = server.app.Shutdown()
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return err
		}
		return nil
	}
}
