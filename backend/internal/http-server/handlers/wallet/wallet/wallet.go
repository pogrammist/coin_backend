package wallet

import (
	"coin-app/internal/lib/logger/sl"
	"context"
	"errors"
	"io"
	"net/http"

	"log/slog"

	resp "coin-app/internal/lib/api/response"
	"coin-app/internal/services/wallet"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request struct {
	WalletId      uuid.UUID `json:"walletId"`
	OperationType string    `json:"operationType"`
	Amount        int       `json:"amount"`
}

type Response struct {
	resp.Response
	TransactionId uuid.UUID `json:"transactionId"`
}

// TODO(pogrammist): не используется
type OperationType string

const (
	Deposit  OperationType = "DEPOSIT"
	Withdraw OperationType = "WITHDRAW"
)

type TransactionSaver interface {
	SaveTransaction(
		ctx context.Context,
		walletId uuid.UUID,
		operationType string,
		amount int,
	) (transactionId uuid.UUID, err error)
}

func New(log *slog.Logger, transactionSaver TransactionSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.wallet.deposit.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		transactionId, err := transactionSaver.SaveTransaction(r.Context(), req.WalletId, req.OperationType, req.Amount)
		if errors.Is(err, wallet.ErrWalletNotExists) {
			log.Warn("wallet not exists", slog.String("walletId", req.WalletId.String()))

			render.JSON(w, r, resp.Error("wallet not exists"))

			return
		}
		if err != nil {
			log.Error("failed to save user", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to save transaction"))

			return
		}

		log.Info("transaction added", slog.String("id", transactionId.String()))

		responseOK(w, r, transactionId)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, transactionId uuid.UUID) {
	render.JSON(w, r, Response{
		Response:      resp.OK(),
		TransactionId: transactionId,
	})
}
