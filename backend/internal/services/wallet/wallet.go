package wallet

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"coin-app/internal/lib/logger/sl"
	"coin-app/internal/storage"
)

type Wallet struct {
	log              *slog.Logger
	walletSaver      WalletSaver
	transactionSaver TransactionSaver
}

type WalletSaver interface {
	SaveWallet(
		ctx context.Context,
		walletId uuid.UUID,
		userId uuid.UUID,
		balance int,
	) (id uuid.UUID, err error)
}

type TransactionSaver interface {
	SaveTransaction(
		ctx context.Context,
		transactionId uuid.UUID,
		walletId uuid.UUID,
		operationType string,
		amount int,
	) (id uuid.UUID, err error)
}

var (
	ErrWalletExists    = errors.New("wallet already exists")
	ErrWalletNotExists = errors.New("wallet not exists")
)

// New returns a new instance of the Wallet service.
func New(
	log *slog.Logger,
	walletSaver WalletSaver,
	transactionSaver TransactionSaver,
) *Wallet {
	return &Wallet{
		log:              log,
		walletSaver:      walletSaver,
		transactionSaver: transactionSaver,
	}
}

func (w *Wallet) SaveWallet(ctx context.Context, userId uuid.UUID, balance int) (uuid.UUID, error) {
	const op = "Wallet.SaveWallet"

	walletId := uuid.New()

	log := w.log.With(
		slog.String("op", op),
		slog.String("walletId", walletId.String()),
		slog.String("userId", userId.String()),
		slog.Int("balance", balance),
	)

	log.Info("creating new wallet")

	id, err := w.walletSaver.SaveWallet(ctx, walletId, userId, balance)
	if err != nil {
		if errors.Is(err, storage.ErrWalletExists) {
			log.Warn("wallet already exists", sl.Err(err))

			return uuid.UUID{}, fmt.Errorf("%s: %w", op, ErrWalletExists)
		}
		log.Error("failed to save wallet", sl.Err(err))

		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("new wallet created successfully")
	return id, nil
}

// SaveTransaction adds deposit or withdraw in the wallet.
// If wallet with given uuid not exists, returns error.
func (w *Wallet) SaveTransaction(ctx context.Context, walletId uuid.UUID, operationType string, amount int) (uuid.UUID, error) {
	const op = "Wallet.SaveTransaction"

	transactionId := uuid.New()

	log := w.log.With(
		slog.String("op", op),
		slog.String("transactionId", transactionId.String()),
		slog.String("walletId", walletId.String()),
		slog.String("operationType", operationType),
		slog.Int("amount", amount),
	)

	log.Info("depositing money")

	id, err := w.transactionSaver.SaveTransaction(ctx, transactionId, walletId, operationType, amount)
	if err != nil {
		if errors.Is(err, storage.ErrWalletNotExists) {
			log.Warn("wallet not exists", sl.Err(err))

			return uuid.UUID{}, fmt.Errorf("%s: %w", op, ErrWalletNotExists)
		}
		log.Error("failed to save transaction", sl.Err(err))

		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("transaction saved successfully")

	return id, nil
}