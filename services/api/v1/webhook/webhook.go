package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	v1 "github.com/cph-collectibles/api/v1"
	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/tables/orders"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

var secret = os.Getenv("STRIPE_WEBHOOK_KEY")

func HandleWith(db *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const maxBodyBytes = int64(1 << 20) // 1 MB
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
			return
		}

		sig := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(payload, sig, secret)
		if err != nil {
			v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
			return
		}

		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
			return
		}

		if pi.Metadata["order_id"] == "" {
			v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
			return
		}

		switch event.Type {
		case "payment_intent.succeeded":
			ctx := r.Context()
			tx := db.Transaction(ctx)
			defer tx.Rollback(ctx)

			order, err := orders.SelectOne(db, ctx, pi.Metadata["order_id"])
			if err != nil {
				v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
				return
			}
			order.Status = "paid"

			err = orders.Update(db, ctx, &order)
			if err != nil {
				v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
				return
			}

			if err := tx.Commit(ctx); err != nil {
				v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
				return
			}

			v1.Success(w, http.StatusOK, []any{})
		case "payment_intent.payment_failed":
			ctx := r.Context()
			tx := db.Transaction(ctx)
			defer tx.Rollback(ctx)

			order, err := orders.SelectOne(db, ctx, pi.Metadata["order_id"])
			if err != nil {
				v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
				return
			}
			order.Status = "cancelled"

			err = orders.Update(db, ctx, &order)
			if err != nil {
				v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
				return
			}

			if err := tx.Commit(ctx); err != nil {
				v1.Error(w, http.StatusBadRequest, fmt.Errorf("webhook.HandleWith: %w", err))
				return
			}

			v1.Success(w, http.StatusOK, []any{})
		default:
			// ignore unneeded events
		}
	}
}
