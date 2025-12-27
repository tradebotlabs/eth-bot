package models

import (
	"time"

	"github.com/google/uuid"
)

// AccountType represents the type of trading account
type AccountType string

const (
	AccountTypeDemo AccountType = "demo"
	AccountTypeLive AccountType = "live"
)

// TradingMode represents the trading mode
type TradingMode string

const (
	TradingModePaper TradingMode = "paper"
	TradingModeLive  TradingMode = "live"
)

// TradingAccount represents a user's trading account (demo or live)
type TradingAccount struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	UserID      uuid.UUID   `json:"user_id" db:"user_id"`
	AccountType AccountType `json:"account_type" db:"account_type"`
	AccountName string      `json:"account_name" db:"account_name"`

	// Demo account fields
	DemoInitialCapital  *float64 `json:"demo_initial_capital,omitempty" db:"demo_initial_capital"`
	DemoCurrentBalance  *float64 `json:"demo_current_balance,omitempty" db:"demo_current_balance"`

	// Live account fields (Binance)
	BinanceAPIKey          *string `json:"binance_api_key,omitempty" db:"binance_api_key"`
	BinanceSecretEncrypted *string `json:"-" db:"binance_secret_key_encrypted"` // Never expose
	BinanceTestnet         bool    `json:"binance_testnet" db:"binance_testnet"`

	// Trading configuration
	TradingSymbol      string      `json:"trading_symbol" db:"trading_symbol"`
	TradingMode        TradingMode `json:"trading_mode" db:"trading_mode"`
	EnabledStrategies  []string    `json:"enabled_strategies" db:"enabled_strategies"`

	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TradingAccountCreateRequest represents a request to create a new trading account
type TradingAccountCreateRequest struct {
	AccountType AccountType `json:"account_type" validate:"required,oneof=demo live"`
	AccountName string      `json:"account_name" validate:"required,min=3,max=100"`

	// Demo account fields (required if account_type = demo)
	DemoInitialCapital *float64 `json:"demo_initial_capital,omitempty" validate:"omitempty,gte=1000,lte=100000"`

	// Live account fields (required if account_type = live)
	BinanceAPIKey    *string `json:"binance_api_key,omitempty" validate:"required_if=AccountType live"`
	BinanceSecretKey *string `json:"binance_secret_key,omitempty" validate:"required_if=AccountType live"`
	BinanceTestnet   bool    `json:"binance_testnet"`

	// Trading configuration
	TradingSymbol     string   `json:"trading_symbol" validate:"required"`
	EnabledStrategies []string `json:"enabled_strategies"`
}

// TradingAccountUpdateRequest represents an account update request
type TradingAccountUpdateRequest struct {
	AccountName       *string  `json:"account_name,omitempty" validate:"omitempty,min=3,max=100"`
	TradingSymbol     *string  `json:"trading_symbol,omitempty"`
	EnabledStrategies []string `json:"enabled_strategies,omitempty"`
	IsActive          *bool    `json:"is_active,omitempty"`
}

// TradingAccountResponse is the public response for a trading account
type TradingAccountResponse struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	AccountType AccountType `json:"account_type"`
	AccountName string      `json:"account_name"`

	// Demo account fields
	DemoInitialCapital *float64 `json:"demo_initial_capital,omitempty"`
	DemoCurrentBalance *float64 `json:"demo_current_balance,omitempty"`

	// Live account info (masked for security)
	BinanceAPIKeyMasked *string `json:"binance_api_key_masked,omitempty"` // Only show last 4 chars
	BinanceTestnet      bool    `json:"binance_testnet"`

	// Trading configuration
	TradingSymbol     string      `json:"trading_symbol"`
	TradingMode       TradingMode `json:"trading_mode"`
	EnabledStrategies []string    `json:"enabled_strategies"`

	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts TradingAccount to TradingAccountResponse
func (a *TradingAccount) ToResponse() *TradingAccountResponse {
	resp := &TradingAccountResponse{
		ID:                a.ID,
		UserID:            a.UserID,
		AccountType:       a.AccountType,
		AccountName:       a.AccountName,
		TradingSymbol:     a.TradingSymbol,
		TradingMode:       a.TradingMode,
		EnabledStrategies: a.EnabledStrategies,
		IsActive:          a.IsActive,
		BinanceTestnet:    a.BinanceTestnet,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}

	// Include demo account fields if demo
	if a.AccountType == AccountTypeDemo {
		resp.DemoInitialCapital = a.DemoInitialCapital
		resp.DemoCurrentBalance = a.DemoCurrentBalance
	}

	// Mask API key for security (show only last 4 characters)
	if a.BinanceAPIKey != nil && len(*a.BinanceAPIKey) > 4 {
		masked := "****" + (*a.BinanceAPIKey)[len(*a.BinanceAPIKey)-4:]
		resp.BinanceAPIKeyMasked = &masked
	}

	return resp
}

// Validate validates a TradingAccountCreateRequest
func (r *TradingAccountCreateRequest) Validate() error {
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
