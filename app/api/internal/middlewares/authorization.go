package middlewares

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/notifyx/api/internal/auth"
)

const (
	scopesContextKey = "scopes"
)

func RequireAuth(validator auth.Validator, scopes ...string) fiber.Handler {
	return func(fiberCtx *fiber.Ctx) error {
		token, err := extractBearer(fiberCtx.Get(fiber.HeaderAuthorization))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		claims, err := validator.Validate(fiberCtx.UserContext(), token)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) {
				return fiber.NewError(fiber.StatusUnauthorized, err.Error())
			}
			return fiber.NewError(fiber.StatusBadGateway, "failed to validate token")
		}

		if !auth.HasScopes(claims, scopes...) {
			return fiber.NewError(fiber.StatusForbidden, auth.ErrInsufficientScope.Error())
		}

		fiberCtx.Locals(orgContextKey, claims.OrgID)
		fiberCtx.Locals(scopesContextKey, claims.Scopes)
		return fiberCtx.Next()
	}
}

func extractBearer(header string) (string, error) {
	if header == "" {
		return "", auth.ErrMissingToken
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", auth.ErrMissingToken
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", auth.ErrMissingToken
	}
	return token, nil
}

