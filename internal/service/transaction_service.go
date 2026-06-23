package service

import (
	"context"
	"errors"
	"time"
	"wallet-service/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionService is the struct that hold our repository interface
type TransactionService struct {
	repo domain.WalletRepository
}

// NewTransactionService is the complete flow for adding money to a user's wallet
func NewTransactionService(repo domain.WalletRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) ProcessTopUp(ctx context.Context, req domain.TopUpRequest) error {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("top-up amount must be greater than zero")
	}

	// The ACID transaction block
	// Everything inside this statement will succeed together or fail together
	err := s.repo.ExecuteTx(ctx, func(txRepo domain.WalletRepository) error {
		// Verify the wallet exists
		wallet, err := txRepo.GetWalletByNumber(ctx, req.DestinationWalletNumber)
		if err != nil {
			return errors.New("destination wallet not found")
		}

		// Create the transaction record
		transactionUUID, err := uuid.NewV7()
		if err != nil {
			return err
		}

		transactionID := transactionUUID.String()
		tx := &domain.Transaction{
			ID:              transactionID,
			RefenceID:       req.ReferenceID,
			WalletID:        wallet.ID,
			TransactionType: "TOPUP",
			Amount:          req.Amount,
			Status:          "Success", // Marking success as we process instantly
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := txRepo.CreateTransaction(ctx, tx); err != nil {
			return err // Triggers a ROLLBACK
		}

		// Create the immutable ledger entry
		ledgerUUID, err := uuid.NewV7()
		if err != nil {
			return err
		}

		ledgerID := ledgerUUID.String()

		// Incrementing a balance safely
		newBalance := wallet.Balance.Add(req.Amount)

		ledger := &domain.WalletLedger{
			ID:            ledgerID,
			WalletID:      wallet.ID,
			TransactionID: transactionID,
			Amount:        req.Amount,
			BalanceAfter:  newBalance,
			Description:   "Top-up from bank " + req.SourceBankCode,
			CreatedAt:     time.Now(),
		}

		if err := txRepo.CreateLedger(ctx, ledger); err != nil {
			return err // Triggers a ROLLBACK
		}

		// Update the cached wallet balance
		if err := txRepo.UpdateWalletBalance(ctx, wallet.ID, req.Amount); err != nil {
			return err // Triggers a ROLLBACK
		}

		return nil // Triggers a COMMIT
	})

	return err
}
