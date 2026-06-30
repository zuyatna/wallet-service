package service

import (
	"context"
	"errors"
	"log"
	"time"
	"wallet-service/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type EventPublisher interface {
	PublishTopUpSuccess(ctx context.Context, tx *domain.Transaction) error
}

// TransactionService is the struct that hold our repository interface
type TransactionService struct {
	repo        domain.WalletRepository
	redisClient *redis.Client
	publisher   EventPublisher
}

// NewTransactionService is the complete flow for adding money to a user's wallet
func NewTransactionService(repo domain.WalletRepository, redisClient *redis.Client, publisher EventPublisher) *TransactionService {
	return &TransactionService{
		repo:        repo,
		redisClient: redisClient,
		publisher:   publisher,
	}
}

func (s *TransactionService) ProcessTopUp(ctx context.Context, req domain.TopUpRequest) error {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("top-up amount must be greater than zero")
	}

	lockKey := "wallet_lock:" + req.DestinationWalletNumber

	locked, err := s.redisClient.SetNX(ctx, lockKey, "locked", 5*time.Second).Result()
	if err != nil || !locked {
		return errors.New("wallet is currently processing another transaction, please try again")
	}
	defer s.redisClient.Del(ctx, lockKey)

	var completedTx *domain.Transaction

	// The ACID transaction block
	// Everything inside this statement will succeed together or fail together
	err = s.repo.ExecuteTx(ctx, func(txRepo domain.WalletRepository) error {
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
		completedTx = &domain.Transaction{
			ID:              transactionID,
			RefenceID:       req.ReferenceID,
			WalletID:        wallet.ID,
			TransactionType: "TOPUP",
			Amount:          req.Amount,
			Status:          "Success", // Marking success as we process instantly
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := txRepo.CreateTransaction(ctx, completedTx); err != nil {
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

	// If the database transaction was successful (err == nil), publish the async event!
	if err == nil && completedTx != nil {
		go func() {
			if pubErr := s.publisher.PublishTopUpSuccess(context.Background(), completedTx); pubErr != nil {
				log.Printf("Failed to publish to RabbitMQ, but DB saved successfully: %v\n", pubErr)
			}
		}()
	}

	return err
}
