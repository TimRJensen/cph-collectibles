package checkout

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	v1 "github.com/cph-collectibles/api/v1"
	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/tables/orderitems"
	"github.com/cph-collectibles/db/tables/orders"
	"github.com/cph-collectibles/db/tables/posters"
	"github.com/cph-collectibles/db/wrappers"
	"github.com/oklog/ulid/v2"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/paymentintent"
)

func HandleWith(conn *db.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			v1.Success(w, http.StatusOK, struct {
				PublishableKey string `json:"publishableKey"`
			}{
				PublishableKey: os.Getenv("STRIPE_PUBLISH_KEY"),
			})
		case http.MethodPost:
			cart := struct {
				Items []string `json:"items"`
			}{}
			err := json.NewDecoder(r.Body).Decode(&cart)
			if err != nil {
				v1.Error(w, http.StatusInternalServerError, fmt.Errorf("checkout.HandleWith: %w", err))
				return
			}

			ctx := context.Background()
			tx := conn.Transaction(ctx)
			defer tx.Rollback(ctx)
			data, err := posters.Query(tx, ctx, cart.Items)
			if err != nil {

				v1.Error(w, http.StatusInternalServerError, fmt.Errorf("checkout.HandleWith: %w", err))
				return
			}

			total := int64(0)
			for _, d := range data {
				total += d.Cost.MinorTotal
			}

			order := orders.Data{
				Id:     wrappers.ULID(ulid.Make()),
				Total:  total,
				Status: "pending",
			}
			if err = orders.Insert(tx, ctx, &order); err != nil {
				v1.Error(w, http.StatusInternalServerError, fmt.Errorf("checkout.HandleWith: %w", err))
				return
			}

			for _, d := range data {
				item := &orderitems.Data{OrderId: order.Id, PosterId: d.Id}
				if err = orderitems.Insert(tx, ctx, item); err != nil {
					v1.Error(w, http.StatusInternalServerError, fmt.Errorf("checkout.HandleWith: %w", err))
					return
				}
			}

			params := &stripe.PaymentIntentParams{
				Amount:   stripe.Int64(total),
				Currency: stripe.String("EUR"),
				AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
					Enabled: stripe.Bool(true),
				},
			}
			params.AddMetadata("order_id", order.Id.String())

			pi, err := paymentintent.New(params)
			if err != nil {
				// Try to safely cast a generic error to a stripe.Error so that we can get at
				// some additional Stripe-specific information about what went wrong.
				if stripeErr, ok := err.(*stripe.Error); ok {
					log.Printf("Stripe error occurred: %v\n", stripeErr.Error())
				} else {
					log.Printf("Other error occurred: %v\n", err)
				}
				v1.Error(w, http.StatusBadRequest, fmt.Errorf("checkout.HandleWith: %w", err))
				return
			}

			order.PaymentIntentId = pi.ID
			if err = orders.Update(tx, ctx, &order); err != nil {
				v1.Error(w, http.StatusInternalServerError, fmt.Errorf("checkout.HandleWith: %w", err))
				return
			}
			tx.Commit(ctx)

			v1.Success(w, http.StatusOK, struct {
				OrderID      string `json:"orderId"`
				ClientSecret string `json:"clientSecret"`
			}{
				OrderID:      order.Id.String(),
				ClientSecret: pi.ClientSecret},
			)
		}
	}
}
