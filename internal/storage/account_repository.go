package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eth-trading/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// TradingAccountRepository implements trading account data access
type TradingAccountRepository struct {
	db *sqlx.DB
}

// NewTradingAccountRepository creates a new trading account repository
func NewTradingAccountRepository(db *sqlx.DB) *TradingAccountRepository {
	return &TradingAccountRepository{db: db}
}

// Create creates a new trading account
func (r *TradingAccountRepository) Create(account *models.TradingAccount) error {
	query := `
		INSERT INTO trading_accounts (
			id, user_id, account_type, account_name,
			demo_initial_capital, demo_current_balance,
			binance_api_key, binance_secret_key_encrypted, binance_testnet,
			trading_symbol, trading_mode, enabled_strategies,
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := r.db.Exec(
		query,
		account.ID,
		account.UserID,
		account.AccountType,
		account.AccountName,
		account.DemoInitialCapital,
		account.DemoCurrentBalance,
		account.BinanceAPIKey,
		account.BinanceSecretEncrypted,
		account.BinanceTestnet,
		account.TradingSymbol,
		account.TradingMode,
		pq.Array(account.EnabledStrategies),
		account.IsActive,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("insert trading account: %w", err)
	}

	return nil
}

// GetByID retrieves a trading account by ID
func (r *TradingAccountRepository) GetByID(id uuid.UUID) (*models.TradingAccount, error) {
	query := `
		SELECT id, user_id, account_type, account_name,
		       demo_initial_capital, demo_current_balance,
		       binance_api_key, binance_secret_key_encrypted, binance_testnet,
		       trading_symbol, trading_mode, enabled_strategies,
		       is_active, created_at, updated_at
		FROM trading_accounts
		WHERE id = $1
	`

	var account models.TradingAccount
	err := r.db.Get(&account, query, id)
	if err == sql.ErrNoRows {
		return nil, models.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get trading account by id: %w", err)
	}

	return &account, nil
}

// GetByUserID retrieves all trading accounts for a user
func (r *TradingAccountRepository) GetByUserID(userID uuid.UUID) ([]*models.TradingAccount, error) {
	query := `
		SELECT id, user_id, account_type, account_name,
		       demo_initial_capital, demo_current_balance,
		       binance_api_key, binance_secret_key_encrypted, binance_testnet,
		       trading_symbol, trading_mode, enabled_strategies,
		       is_active, created_at, updated_at
		FROM trading_accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var accounts []*models.TradingAccount
	err := r.db.Select(&accounts, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get trading accounts by user id: %w", err)
	}

	return accounts, nil
}

// Update updates a trading account
func (r *TradingAccountRepository) Update(account *models.TradingAccount) error {
	query := `
		UPDATE trading_accounts
		SET account_name = $2,
		    demo_current_balance = $3,
		    binance_api_key = $4,
		    binance_secret_key_encrypted = $5,
		    binance_testnet = $6,
		    trading_symbol = $7,
		    trading_mode = $8,
		    enabled_strategies = $9,
		    is_active = $10,
		    updated_at = $11
		WHERE id = $1
	`

	account.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		account.ID,
		account.AccountName,
		account.DemoCurrentBalance,
		account.BinanceAPIKey,
		account.BinanceSecretEncrypted,
		account.BinanceTestnet,
		account.TradingSymbol,
		account.TradingMode,
		pq.Array(account.EnabledStrategies),
		account.IsActive,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("update trading account: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrAccountNotFound
	}

	return nil
}

// Delete deletes a trading account (soft delete)
func (r *TradingAccountRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE trading_accounts
		SET is_active = false, updated_at = $2
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id, time.Now())
	if err != nil {
		return fmt.Errorf("delete trading account: %w", err)
	}

	return nil
}

// UpdateBalance updates the current balance for a demo account
func (r *TradingAccountRepository) UpdateBalance(accountID uuid.UUID, newBalance float64) error {
	query := `
		UPDATE trading_accounts
		SET demo_current_balance = $2, updated_at = $3
		WHERE id = $1 AND account_type = 'demo'
	`

	_, err := r.db.Exec(query, accountID, newBalance, time.Now())
	if err != nil {
		return fmt.Errorf("update account balance: %w", err)
	}

	return nil
}
