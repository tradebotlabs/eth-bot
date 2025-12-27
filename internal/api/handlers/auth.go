package handlers

import (
	"net/http"

	"github.com/eth-trading/internal/api/middleware"
	"github.com/eth-trading/internal/auth"
	"github.com/eth-trading/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Register user
	user, account, err := h.authService.Register(&req)
	if err != nil {
		if err == models.ErrUserAlreadyExists {
			return echo.NewHTTPError(http.StatusConflict, "email already registered")
		}
		if err == models.ErrDemoCapitalRequired || err == models.ErrDemoCapitalOutOfRange {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err == models.ErrBinanceAPIKeyRequired || err == models.ErrBinanceSecretRequired {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		log.Error().Err(err).Msg("Failed to register user")
		return echo.NewHTTPError(http.StatusInternalServerError, "registration failed")
	}

	log.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Str("account_type", string(account.AccountType)).
		Msg("User registered successfully")

	// Auto-login after registration
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	loginResp, err := h.authService.Login(user.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		// Registration succeeded but login failed - still return success
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "registration successful, please login",
			"user":    user.ToResponse(),
			"account": account.ToResponse(),
		})
	}

	return c.JSON(http.StatusCreated, loginResp)
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	// Login user
	resp, err := h.authService.Login(req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		if err == models.ErrInvalidCredentials {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
		}
		if err == models.ErrUserInactive {
			return echo.NewHTTPError(http.StatusForbidden, "account is inactive")
		}

		log.Error().Err(err).Str("email", req.Email).Msg("Login failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "login failed")
	}

	log.Info().
		Str("user_id", resp.User.ID.String()).
		Str("email", resp.User.Email).
		Str("ip", ipAddress).
		Msg("User logged in successfully")

	return c.JSON(http.StatusOK, resp)
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	// Logout user (delete all sessions)
	if err := h.authService.Logout(userID); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Logout failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "logout failed")
	}

	log.Info().Str("user_id", userID.String()).Msg("User logged out successfully")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req models.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Refresh token
	resp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if err == models.ErrSessionNotFound {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
		}
		if err == models.ErrSessionExpired {
			return echo.NewHTTPError(http.StatusUnauthorized, "refresh token expired")
		}

		log.Error().Err(err).Msg("Token refresh failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "token refresh failed")
	}

	return c.JSON(http.StatusOK, resp)
}

// GetMe returns the current authenticated user
// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c echo.Context) error {
	// Get user claims from context
	claims, err := middleware.GetUserClaims(c)
	if err != nil {
		return err
	}

	// In a real implementation, we'd fetch fresh user data from database
	// For now, return claims as user info
	user := &models.UserResponse{
		ID:       claims.UserID,
		Email:    claims.Email,
		Role:     claims.Role,
		IsActive: true,
	}

	return c.JSON(http.StatusOK, user)
}

// ChangePassword handles password change
// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	var req models.PasswordChangeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Change password
	if err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		if err == models.ErrPasswordMismatch {
			return echo.NewHTTPError(http.StatusBadRequest, "current password is incorrect")
		}

		log.Error().Err(err).Str("user_id", userID.String()).Msg("Password change failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "password change failed")
	}

	log.Info().Str("user_id", userID.String()).Msg("Password changed successfully")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "password changed successfully, please login again",
	})
}

// RequestPasswordReset handles password reset request
// POST /api/v1/auth/password-reset
func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	var req models.PasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// TODO: Implement password reset email sending
	// For now, just return success
	log.Info().Str("email", req.Email).Msg("Password reset requested")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "if the email exists, a password reset link has been sent",
	})
}

// ConfirmPasswordReset handles password reset confirmation
// POST /api/v1/auth/password-reset/confirm
func (h *AuthHandler) ConfirmPasswordReset(c echo.Context) error {
	var req models.PasswordResetConfirm
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// TODO: Implement password reset token validation and password update
	log.Info().Msg("Password reset confirmed")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "password reset successfully, please login",
	})
}
