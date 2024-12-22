package withdraw

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
}

type WithdrawSaver interface {
	Withdraw(
		ctx context.Context,
		walletId uuid.UUID,
		operationType string,
		amount int,
	) (err error)
}

func New(log *slog.Logger, withdrawSaver WithdrawSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.wallet.withdraw.New"

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

		err = withdrawSaver.Withdraw(r.Context(), req.WalletId, req.OperationType, req.Amount)
		if errors.Is(err, wallet.ErrWalletNotExists) {
			log.Warn("wallet not exists", slog.String("walletId", req.WalletId.String()))

			render.JSON(w, r, resp.Error("wallet not exists"))

			return
		}
		if err != nil {
			log.Error("failed to withdraw funds", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to withdraw funds"))

			return
		}

		log.Info("withdrawal processed")

		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
