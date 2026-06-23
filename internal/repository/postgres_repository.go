package repository

import (
	"context"
	"database/sql"
	"errors"
	"wallet-service/internal/domain"

	"github.com/shopspring/decimal"
)

type PostgresWalletRepository struct {
	db *sql.DB
	tx *sql.Tx // This will be non-nil if we are inside a transaction block
}

// NewPostgresWalletRepository is our constructor
func NewPostgresWalletRepository(db *sql.DB) *PostgresWalletRepository {
	return &PostgresWalletRepository{
		db: db,
	}
}

// ExecuteTx creates a safe environment for multiple queries to run together
func (r *PostgresWalletRepository) ExecuteTx(ctx context.Context, fn func(wallet domain.WalletRepository) error) error {
	// Start the transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Create a TEMPORARY copy of repository that uses the 'tx' instead of 'db'
	txRepo := &PostgresWalletRepository{
		db: r.db,
		tx: tx,
	}

	// Run the business logic function using this temporary repository
	err = fn(txRepo)

	// If the function returned an error, we undo everything!
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(rbErr, err) // Return both errors if rollback also fails
		}
		return err
	}

	// If there were zero errors, save everything permanently!
	return tx.Commit()
}

// exec checks if we are in a transaction. If yes, it uses 'tx'. If no, it uses 'db'.
func (r *PostgresWalletRepository) exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if r.tx != nil {
		return r.tx.ExecContext(ctx, query, args...)
	}
	return r.db.ExecContext(ctx, query, args...)
}

// GetWalletByNumber fetches the wallet by its wallet_number
func (r *PostgresWalletRepository) GetWalletByNumber(ctx context.Context, walletNumber string) (*domain.Wallet, error) {
	query := `
		SELECT id, user_id, wallet_type, wallet_number, currency, balance, created_at, updated_at
		FROM wallets 
		WHERE wallet_number = $1
	`

	var row *sql.Row
	if r.tx != nil {
		row = r.tx.QueryRowContext(ctx, query, walletNumber)
	} else {
		row = r.db.QueryRowContext(ctx, query, walletNumber)
	}

	var w domain.Wallet
	err := row.Scan(
		&w.ID,
		&w.UserID,
		&w.WalletType,
		&w.WalletNumber,
		&w.Currency,
		&w.Balance,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}

	return &w, nil
}

// CreateTransaction saves the intent
func (r *PostgresWalletRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (id, reference_id, wallet_id, transaction_type, amount, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.exec(ctx, query,
		tx.ID, tx.RefenceID, tx.WalletID, tx.TransactionType, tx.Amount, tx.Status, tx.Metadata,
	)
	return err
}

// CreateLedger creates the immutable history log
func (r *PostgresWalletRepository) CreateLedger(ctx context.Context, ledger *domain.WalletLedger) error {
	query := `
		INSERT INTO wallet_ledgers (id, wallet_id, transaction_id, amount, balance_after, description)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.exec(ctx, query,
		ledger.ID, ledger.WalletID, ledger.TransactionID, ledger.Amount, ledger.BalanceAfter, ledger.Description,
	)
	return err
}

// UpdateWalletBalance adjusts the user's cached balance
func (r *PostgresWalletRepository) UpdateWalletBalance(ctx context.Context, walletID string, amount decimal.Decimal) error {
	query := `
		UPDATE wallets 
		SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	// PostgreSQL will mathematically add the decimal amount accurately!
	_, err := r.exec(ctx, query, amount, walletID)
	return err
}
