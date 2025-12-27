package middleware

import (
	"net/http"
	"strings"

	"github.com/eth-trading/internal/auth"
	"github.com/eth-trading/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	authService *auth.Service
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *auth.Service) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// contextKey is a custom type for context keys
type contextKey string

const (
	// UserContextKey is the key for user claims in context
	UserContextKey contextKey = "user"
)

// Authenticate is middleware that validates JWT tokens
func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
		}

		token := parts[1]

		// Validate token
		claims, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
		}

		// Store claims in context
		c.Set(string(UserContextKey), claims)

		return next(c)
	}
}

// RequireRole is middleware that checks if user has required role
func (m *AuthMiddleware) RequireRole(roles ...models.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user claims from context
			claims, ok := c.Get(string(UserContextKey)).(*models.JWTClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(c)
		}
	}
}

// RequireOwnership is middleware that checks if user owns the resource
func (m *AuthMiddleware) RequireOwnership(getUserID func(c echo.Context) (uuid.UUID, error)) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user claims from context
			claims, ok := c.Get(string(UserContextKey)).(*models.JWTClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
			}

			// Get resource owner ID
			ownerID, err := getUserID(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid resource ID")
			}

			// Admins can access any resource
			if claims.Role == models.RoleAdmin {
				return next(c)
			}

			// Check if user owns the resource
			if claims.UserID != ownerID {
				return echo.NewHTTPError(http.StatusForbidden, "access denied")
			}

			return next(c)
		}
	}
}

// GetUserClaims retrieves user claims from echo context
func GetUserClaims(c echo.Context) (*models.JWTClaims, error) {
	claims, ok := c.Get(string(UserContextKey)).(*models.JWTClaims)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}
	return claims, nil
}

// GetUserID retrieves user ID from echo context
func GetUserID(c echo.Context) (uuid.UUID, error) {
	claims, err := GetUserClaims(c)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

// Optional is middleware that validates JWT tokens but doesn't require them
func (m *AuthMiddleware) Optional(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return next(c) // No token, continue without authentication
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return next(c) // Invalid format, continue without authentication
		}

		token := parts[1]

		// Validate token
		claims, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			return next(c) // Invalid token, continue without authentication
		}

		// Store claims in context
		c.Set(string(UserContextKey), claims)

		return next(c)
	}
}
