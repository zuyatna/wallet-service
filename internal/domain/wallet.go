package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Wallet represents the current balance snapshot of an e-wallet account.
type Wallet struct {
	ID           string          `json:"id" db:"id"`
	UserID       string          `json:"user_id" db:"user_id"`
	WalletType   string          `json:"wallet_type" db:"wallet_type"`
	WalletNumber string          `json:"wallet_number" db:"wallet_number"`
	Currency     string          `json:"currency" db:"currency"`
	Balance      decimal.Decimal `json:"balance" db:"balance"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}
