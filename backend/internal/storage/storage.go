package storage

import "errors"

var (
	ErrWalletExists    = errors.New("wallet already exists")
	ErrWalletNotExists = errors.New("wallet not exists")
)
