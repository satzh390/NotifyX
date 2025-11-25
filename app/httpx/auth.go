package httpx

import (
	"context"
	"errors"
	"strings"
	"time"

	keyfunc "github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
)

const (
	CustomerIDKey = "customerId"
	ScopesKey    = "scopes"
)

var (
	ErrMissingToken      = errors.New("auth: missing bearer token")
	ErrInvalidToken      = errors.New("auth: token invalid")
	ErrInsufficientScope = errors.New("auth: insufficient scope")
)

// Claims represents JWT token claims
type Claims struct {
	CustomerID string
	Scopes     []string
	Subject    string
}

// AuthValidator is an interface for validating authentication tokens
type AuthValidator interface {
	Validate(ctx context.Context, token string) (Claims, error)
}

// RequireAuth creates a middleware that validates authentication tokens
// It extracts the Bearer token, validates it, checks scopes, and stores claims in context
func RequireAuth(validator AuthValidator, scopes ...string) fiber.Handler {
	return func(fiberCtx *fiber.Ctx) error {
		token, err := extractBearer(fiberCtx.Get(fiber.HeaderAuthorization))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		claims, err := validator.Validate(fiberCtx.UserContext(), token)
		if err != nil {
			if errors.Is(err, ErrInvalidToken) {
				return fiber.NewError(fiber.StatusUnauthorized, err.Error())
			}
			return fiber.NewError(fiber.StatusBadGateway, "failed to validate token")
		}

		if !HasScopes(claims, scopes...) {
			return fiber.NewError(fiber.StatusForbidden, ErrInsufficientScope.Error())
		}

		fiberCtx.Locals(CustomerIDKey, claims.CustomerID)
		fiberCtx.Locals(ScopesKey, claims.Scopes)
		return fiberCtx.Next()
	}
}

// extractBearer extracts the Bearer token from the Authorization header
func extractBearer(header string) (string, error) {
	if header == "" {
		return "", ErrMissingToken
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrMissingToken
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", ErrMissingToken
	}
	return token, nil
}

// HasScopes checks if the claims contain all required scopes
func HasScopes(claims Claims, required ...string) bool {
	if len(required) == 0 {
		return true
	}
	for _, requiredScope := range required {
		match := false
		for _, scope := range claims.Scopes {
			if scope == requiredScope {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}

// CustomerIDFromCtx extracts the customer ID from Fiber context
func CustomerIDFromCtx(fiberCtx *fiber.Ctx) string {
	if value := fiberCtx.Locals(CustomerIDKey); value != nil {
		if customerID, ok := value.(string); ok {
			return customerID
		}
	}
	return ""
}

// JWKSValidator validates JWT tokens using JWKS (JSON Web Key Set)
type JWKSValidator struct {
	issuer   string
	audience []string
	jwks     *keyfunc.JWKS
}

// NewJWKSValidator creates a new JWKS validator
func NewJWKSValidator(ctx context.Context, issuer, jwksURL string, audience []string) (*JWKSValidator, error) {
	if issuer == "" {
		return nil, errors.New("auth: issuer is required")
	}
	if jwksURL == "" {
		return nil, errors.New("auth: jwks url is required")
	}

	opts := keyfunc.Options{
		Ctx: ctx,
		RefreshErrorHandler: func(err error) {
			// ignore refresh errors, will retry on next interval
		},
		RefreshInterval: 15 * time.Minute,
		RefreshTimeout:  5 * time.Second,
	}

	jwks, err := keyfunc.Get(jwksURL, opts)
	if err != nil {
		return nil, err
	}

	return &JWKSValidator{
		issuer:   issuer,
		audience: audience,
		jwks:     jwks,
	}, nil
}

// Validate validates a JWT token and returns the claims
func (validator *JWKSValidator) Validate(ctx context.Context, token string) (Claims, error) {
	if token == "" {
		return Claims{}, ErrMissingToken
	}

	parserOpts := []jwt.ParserOption{
		jwt.WithIssuer(validator.issuer),
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}),
	}
	if len(validator.audience) > 0 {
		parserOpts = append(parserOpts, jwt.WithAudience(validator.audience...))
	}

	parsedToken, err := jwt.Parse(token, validator.jwks.Keyfunc, parserOpts...)
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	if !parsedToken.Valid {
		return Claims{}, ErrInvalidToken
	}

	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, ErrInvalidToken
	}

	customerID := getStringClaim(mapClaims["customerId"])
	if customerID == "" {
		return Claims{}, errors.New("auth: customerId missing in token")
	}

	return Claims{
		CustomerID: customerID,
		Scopes:     extractScopes(mapClaims),
		Subject:    getStringClaim(mapClaims["sub"]),
	}, nil
}

// extractScopes extracts scopes from JWT claims
func extractScopes(claims jwt.MapClaims) []string {
	if scopeStr, ok := claims["scope"].(string); ok {
		return splitScopes(scopeStr)
	}
	if scopes, ok := claims["scp"].([]any); ok {
		result := make([]string, 0, len(scopes))
		for _, item := range scopes {
			if scopeStr, ok := item.(string); ok && scopeStr != "" {
				result = append(result, scopeStr)
			}
		}
		return result
	}
	return nil
}

// splitScopes splits a space-separated scope string into a slice
func splitScopes(scope string) []string {
	if scope == "" {
		return nil
	}
	parts := strings.Fields(scope)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// getStringClaim extracts a string claim from JWT claims
func getStringClaim(value any) string {
	if strValue, ok := value.(string); ok {
		return strings.TrimSpace(strValue)
	}
	return ""
}

