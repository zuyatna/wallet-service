package domain

import (
	"context"

	"github.com/shopspring/decimal"
)

// WalletRepository defines all the database operations we need.
type WalletRepository interface {
	// ExecuteTx handles the ACID transaction block (BEGIN, COMMIT, ROLLBACK)
	ExecuteTx(ctx context.Context, fn func(repo WalletRepository) error) error

	GetWalletByNumber(ctx context.Context, walletNumber string) (*Wallet, error)
	CreateTransaction(ctx context.Context, tx *Transaction) error
	CreateLedger(ctx context.Context, ledger *WalletLedger) error
	UpdateWalletBalance(ctx context.Context, walletID string, amount decimal.Decimal) error
}
