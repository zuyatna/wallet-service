package domain

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// Transaction represents the intent of a transaction request.
type Transaction struct {
	ID              string          `json:"id" db:"id"`
	RefenceID       string          `json:"refence_id" db:"refence_id"`
	WalletID        string          `json:"wallet_id" db:"wallet_id"`
	TransactionType string          `json:"transaction_type" db:"transaction_type"`
	Amount          decimal.Decimal `json:"amount" db:"amount"`
	Status          string          `json:"status" db:"status"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// WalletLedger represents the absolute, immutable history entry for mutations.
type WalletLedger struct {
	ID            string          `json:"id" db:"id"`
	WalletID      string          `json:"wallet_id" db:"wallet_id"`
	TransactionID string          `json:"transaction_id" db:"transaction_id"`
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	BalanceAfter  decimal.Decimal `json:"balance_after" db:"balance_after"`
	Description   string          `json:"description" db:"description"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

type TopUpRequest struct {
	ReferenceID             string          `json:"reference_id" validate:"required,uuid"`
	SourceBankCode          string          `json:"source_bank" validate:"required"`
	SourceAccountNumber     string          `json:"source_account_number" validate:"required"`
	DestinationWalletType   string          `json:"destination_wallet_type" validate:"required"`
	DestinationWalletNumber string          `json:"destination_wallet_number" validate:"required"`
	Amount                  decimal.Decimal `json:"amount" validate:"required,gt=0"`        // Must be greater than 0
	Currency                string          `json:"currency" validate:"required,oneof=IDR"` // Only accept specific currency
}
