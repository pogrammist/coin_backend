package wallet

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"coin-app/internal/lib/logger/sl"
	"coin-app/internal/storage"
)

type Wallet struct {
	log           *slog.Logger
	depositSaver  DepositSaver
	withdrawSaver WithdrawSaver
}

type DepositSaver interface {
	SaveDeposit(
		ctx context.Context,
		walletId string,
		amount int,
	) (err error)
}

type WithdrawSaver interface {
	SaveWithdraw(
		ctx context.Context,
		walletId string,
		amount int,
	) (err error)
}

var (
	ErrWalletNotExists = errors.New("wallet not exists")
)

// New returns a new instance of the Wallet service.
func New(
	log *slog.Logger,
	depositSaver DepositSaver,
	withdrawSaver WithdrawSaver,

) *Wallet {
	return &Wallet{
		log:           log,
		depositSaver:  depositSaver,
		withdrawSaver: withdrawSaver,
	}
}

// SaveDeposit adds deposit in the wallet.
// If wallet with given uuid not exists, returns error.
func (w *Wallet) SaveDeposit(ctx context.Context, walletId string, amount int) error {
	const op = "Wallet.SaveDeposit"

	log := w.log.With(
		slog.String("op", op),
		slog.String("walletId", walletId),
	)

	log.Info("depositing money")

	err := w.depositSaver.SaveDeposit(ctx, walletId, amount)
	if err != nil {
		if errors.Is(err, storage.ErrWalletNotExists) {
			log.Warn("wallet not exists", sl.Err(err))

			return fmt.Errorf("%s: %w", op, ErrWalletNotExists)
		}
		log.Error("failed to save deposit", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("deposit saved")

	return nil
}

func (w *Wallet) SaveWithdraw(ctx context.Context, walletId string, amount int) error {
	const op = "Wallet.SaveWithdraw"

	log := w.log.With(
		slog.String("op", op),
		slog.String("walletId", walletId),
	)

	log.Info("withdrawing money")

	err := w.withdrawSaver.SaveWithdraw(ctx, walletId, amount)
	if err != nil {
		if errors.Is(err, storage.ErrWalletNotExists) {
			log.Warn("wallet not exists", sl.Err(err))

			return fmt.Errorf("%s: %w", op, ErrWalletNotExists)
		}
		log.Error("failed to save withdraw", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("withdraw saved")

	return nil
}
