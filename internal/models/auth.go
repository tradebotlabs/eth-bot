package models

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a user session with refresh token
type Session struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2"`

	// Account setup (demo or live)
	AccountType AccountType `json:"account_type" validate:"required,oneof=demo live"`
	AccountName string      `json:"account_name" validate:"required,min=3,max=100"`

	// Demo account fields (required if account_type = demo)
	DemoInitialCapital *float64 `json:"demo_initial_capital,omitempty" validate:"omitempty,gte=1000,lte=100000"`

	// Live account fields (required if account_type = live)
	BinanceAPIKey    *string `json:"binance_api_key,omitempty"`
	BinanceSecretKey *string `json:"binance_secret_key,omitempty"`
	BinanceTestnet   bool    `json:"binance_testnet"`
}

// Validate validates the RegisterRequest
func (r *RegisterRequest) Validate() error {
	// Demo account must have initial capital
	if r.AccountType == AccountTypeDemo {
		if r.DemoInitialCapital == nil {
			return ErrDemoCapitalRequired
		}
		if *r.DemoInitialCapital < 1000 || *r.DemoInitialCapital > 100000 {
			return ErrDemoCapitalOutOfRange
		}
	}

	// Live account must have Binance credentials
	if r.AccountType == AccountTypeLive {
		if r.BinanceAPIKey == nil || *r.BinanceAPIKey == "" {
			return ErrBinanceAPIKeyRequired
		}
		if r.BinanceSecretKey == nil || *r.BinanceSecretKey == "" {
			return ErrBinanceSecretRequired
		}
	}

	return nil
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	User         *UserResponse           `json:"user"`
	AccessToken  string                  `json:"access_token"`
	RefreshToken string                  `json:"refresh_token"`
	ExpiresIn    int64                   `json:"expires_in"` // seconds
	Accounts     []*TradingAccountResponse `json:"accounts"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   UserRole  `json:"role"`
	Exp    int64     `json:"exp"` // Expiration time
	Iat    int64     `json:"iat"` // Issued at
}

// TwoFactorSetupRequest represents 2FA setup initiation
type TwoFactorSetupRequest struct {
	Password string `json:"password" validate:"required"`
}

// TwoFactorSetupResponse contains QR code and backup codes
type TwoFactorSetupResponse struct {
	QRCode      string   `json:"qr_code"`       // Base64 encoded QR code image
	Secret      string   `json:"secret"`        // TOTP secret for manual entry
	BackupCodes []string `json:"backup_codes"`  // One-time backup codes
}

// TwoFactorVerifyRequest represents 2FA verification
type TwoFactorVerifyRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// TwoFactorAuth represents user's 2FA settings
type TwoFactorAuth struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Secret      string    `json:"-" db:"secret"` // Never expose
	BackupCodes []string  `json:"-" db:"backup_codes"` // Hashed codes
	IsEnabled   bool      `json:"is_enabled" db:"is_enabled"`
	EnabledAt   *time.Time `json:"enabled_at" db:"enabled_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// APIKey represents a programmatic API access key
type APIKey struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	KeyName      string     `json:"key_name" db:"key_name"`
	APIKeyHash   string     `json:"-" db:"api_key_hash"`      // Never expose
	APISecretHash string    `json:"-" db:"api_secret_hash"`   // Never expose
	Permissions  []string   `json:"permissions" db:"permissions"`
	IPWhitelist  []string   `json:"ip_whitelist" db:"ip_whitelist"`
	LastUsedAt   *time.Time `json:"last_used_at" db:"last_used_at"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// APIKeyCreateRequest represents API key creation request
type APIKeyCreateRequest struct {
	KeyName     string    `json:"key_name" validate:"required,min=3,max=100"`
	Permissions []string  `json:"permissions" validate:"required,min=1"`
	IPWhitelist []string  `json:"ip_whitelist,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// APIKeyCreateResponse contains the created API key (only shown once)
type APIKeyCreateResponse struct {
	ID        uuid.UUID  `json:"id"`
	APIKey    string     `json:"api_key"`    // Only shown once during creation
	APISecret string     `json:"api_secret"` // Only shown once during creation
	KeyName   string     `json:"key_name"`
	Permissions []string `json:"permissions"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
