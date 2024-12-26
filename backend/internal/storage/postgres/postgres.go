package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"coin-app/internal/domain/models"
	"coin-app/internal/storage"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New() (*Storage, error) {
	const op = "storage.postgres.New"

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", "dbase", user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// SaveWallet saves wallet to db.
func (s *Storage) SaveWallet(ctx context.Context, walletId uuid.UUID, userId uuid.UUID, balance int) (uuid.UUID, error) {
	const op = "storage.postgres.SaveWallet"

	stmt, err := s.db.Prepare("INSERT INTO wallets(id, user_id, balance) VALUES($1, $2, $3) RETURNING id")
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	err = stmt.QueryRowContext(ctx, walletId, userId, balance).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return uuid.UUID{}, fmt.Errorf("%s: %w", op, storage.ErrWalletExists)
		}

		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// SaveTransaction saves deposit to db.
func (s *Storage) SaveTransaction(ctx context.Context, transactionId uuid.UUID, walletId uuid.UUID, operationType string, amount int) (uuid.UUID, error) {
	const op = "storage.postgres.SaveDeposit"

	stmt, err := s.db.Prepare("INSERT INTO transactions(id, wallet_id, operation_type, amount) VALUES($1, $2, $3, $4) RETURNING id")
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	err = stmt.QueryRowContext(ctx, transactionId, walletId, operationType, amount).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return uuid.UUID{}, fmt.Errorf("%s: %w", op, storage.ErrWalletNotExists)
		}

		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// UpdateWallet updates wallet to db.
func (s *Storage) UpdateBalance(ctx context.Context, walletId uuid.UUID, amount int) error {
	const op = "storage.postgres.UpdateWallet"

	stmt, err := s.db.Prepare("UPDATE wallets SET balance = balance + $1 WHERE id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, amount, walletId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetWallet retrieves wallet from db.
func (s *Storage) GetWallet(ctx context.Context, walletId uuid.UUID) (models.Wallet, error) {
	const op = "storage.postgres.GetWallet"

	stmt, err := s.db.Prepare("SELECT id, user_id, balance, created_at, updated_at FROM wallets WHERE id = $1")
	if err != nil {
		return models.Wallet{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var wallet models.Wallet
	err = stmt.QueryRowContext(ctx, walletId).Scan(&wallet.Id, &wallet.UserId, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Wallet{}, fmt.Errorf("%s: %w", op, storage.ErrWalletNotExists)
		}
		return models.Wallet{}, fmt.Errorf("%s: %w", op, err)
	}

	return wallet, nil
}
