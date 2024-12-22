package create

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
	UserId uuid.UUID `json:"userId"`
	Amount int       `json:"amount"`
}

type Response struct {
	resp.Response
	WalletId uuid.UUID `json:"walletId"`
}

type WalletSaver interface {
	SaveWallet(
		ctx context.Context,
		UserId uuid.UUID,
		amount int,
	) (walletId uuid.UUID, err error)
}

func New(log *slog.Logger, walletSaver WalletSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.wallet.create.New"

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

		walletId, err := walletSaver.SaveWallet(r.Context(), req.UserId, req.Amount)
		if errors.Is(err, wallet.ErrWalletExists) {
			log.Warn("wallet already exists", slog.String("walletId", walletId.String()))

			render.JSON(w, r, resp.Error("wallet already exists"))

			return
		}
		if err != nil {
			log.Error("failed to save wallet", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to save wallet"))

			return
		}

		log.Info("wallet added", slog.String("id", walletId.String()))

		responseOK(w, r, walletId)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, walletId uuid.UUID) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		WalletId: walletId,
	})
}
