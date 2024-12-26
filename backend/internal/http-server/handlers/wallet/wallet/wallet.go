package wallet

import (
	"coin-app/internal/lib/logger/sl"
	"context"
	"net/http"

	"log/slog"

	"coin-app/internal/domain/models"
	resp "coin-app/internal/lib/api/response"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

// type Request struct {
// 	WalletId uuid.UUID `json:"walletId"`
// }

type Response struct {
	resp.Response
	TransactionId uuid.UUID     `json:"transactionId"`
	Wallet        models.Wallet `json:"wallet"`
}

type WalletProvider interface {
	GetWallet(
		ctx context.Context,
		walletId uuid.UUID,
	) (wallet models.Wallet, err error)
}

func New(log *slog.Logger, walletProvider WalletProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.wallet.wallet.New"

		// Извлечение walletId из URL параметров
		walletIdParam := chi.URLParam(r, "walletId")
		walletId, err := uuid.Parse(walletIdParam) // Парсинг walletId из строки
		if err != nil {
			log.Error("invalid walletId", sl.Err(err))
			render.JSON(w, r, resp.Error("invalid walletId"))
			return
		}

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("walletId extracted", slog.String("walletId", walletId.String()))

		// Получение кошелька
		wallet, err := walletProvider.GetWallet(r.Context(), walletId)
		if err != nil {
			log.Error("failed to get wallet", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to get wallet"))
			return
		}

		log.Info("wallet retrieved", slog.Any("wallet", wallet))

		responseOK(w, r, wallet)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, wallet models.Wallet) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Wallet:   wallet,
	})
}
