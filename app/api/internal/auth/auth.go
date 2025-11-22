package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	keyfunc "github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/notifyx/httpx"
)

// Claims is an alias to httpx.Claims for backward compatibility
type Claims = httpx.Claims

// AuthValidator is an alias to httpx.AuthValidator for backward compatibility
type AuthValidator = httpx.AuthValidator

// Use httpx error constants
var (
	ErrMissingToken      = httpx.ErrMissingToken
	ErrInvalidToken      = httpx.ErrInvalidToken
	ErrInsufficientScope  = httpx.ErrInsufficientScope
)

type JWKSValidator struct {
	issuer   string
	audience []string
	jwks     *keyfunc.JWKS
}

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

func (validator *JWKSValidator) Validate(ctx context.Context, token string) (httpx.Claims, error) {
	if token == "" {
		return httpx.Claims{}, ErrMissingToken
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
		return httpx.Claims{}, ErrInvalidToken
	}
	if !parsedToken.Valid {
		return httpx.Claims{}, ErrInvalidToken
	}

	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return httpx.Claims{}, ErrInvalidToken
	}

	orgID := getStringClaim(mapClaims["orgId"])
	if orgID == "" {
		return httpx.Claims{}, errors.New("auth: orgId missing in token")
	}

	return httpx.Claims{
		OrgID:   orgID,
		Scopes:  extractScopes(mapClaims),
		Subject: getStringClaim(mapClaims["sub"]),
	}, nil
}

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

func getStringClaim(value any) string {
	if strValue, ok := value.(string); ok {
		return strings.TrimSpace(strValue)
	}
	return ""
}

// HasScopes is an alias to httpx.HasScopes for backward compatibility
var HasScopes = httpx.HasScopes
