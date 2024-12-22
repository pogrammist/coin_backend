package models

import "github.com/google/uuid"

type Wallet struct {
	ID       int64
	WalletId uuid.UUID
	UserId   uuid.UUID
}
