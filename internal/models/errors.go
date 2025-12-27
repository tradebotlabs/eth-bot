package models

import "errors"

var (
	// User errors
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserInactive         = errors.New("user account is inactive")
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrWeakPassword         = errors.New("password does not meet requirements")
	ErrPasswordMismatch     = errors.New("current password is incorrect")

	// Account errors
	ErrAccountNotFound          = errors.New("trading account not found")
	ErrAccountAlreadyExists     = errors.New("account with this name already exists")
	ErrUnauthorizedAccount      = errors.New("unauthorized to access this account")
	ErrDemoCapitalRequired      = errors.New("demo initial capital is required for demo accounts")
	ErrDemoCapitalOutOfRange    = errors.New("demo capital must be between $1,000 and $100,000")
	ErrBinanceAPIKeyRequired    = errors.New("binance API key is required for live accounts")
	ErrBinanceSecretRequired    = errors.New("binance secret key is required for live accounts")
	ErrAccountInactive          = errors.New("account is inactive")

	// Session errors
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")

	// General errors
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInternalError    = errors.New("internal server error")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
