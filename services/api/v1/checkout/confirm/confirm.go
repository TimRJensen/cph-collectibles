package confirm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cph-collectibles/api/mailer"
	v1 "github.com/cph-collectibles/api/v1"
	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/tables/orders"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/paymentintent"
)

type shipping struct {
	Name      string `json:"name"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Address   struct {
		Line1      string `json:"line1"`
		Line2      string `json:"line2"`
		City       string `json:"city"`
		State      string `json:"state"`
		PostalCode string `json:"postal_code"`
		Country    string `json:"country"`
	} `json:"address"`
}

type Request struct {
	OrderID             string   `json:"orderId"`
	ConfirmationTokenID string   `json:"confirmationTokenId"`
	Email               string   `json:"email"`
	Shipping            shipping `json:"shipping"`
}

type Response struct {
	Status       string `json:"status"`
	RedirectURL  string `json:"redirectURL,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

func validAction(v *stripe.PaymentIntentNextAction) bool {
	switch {
	case v == nil:
		return false
	case v.Type != stripe.PaymentIntentNextActionTypeRedirectToURL:
		return false
	case v.RedirectToURL == nil:
		return false
	default:
		return true
	}
}

func HandleWith(conn *db.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			id := r.PathValue("id")
			if id == "" {
				v1.Error(w, http.StatusBadRequest, fmt.Errorf("checkout.HandleWith: path param 'id' missing"))
				return
			}

			req := Request{}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				v1.Error(w, http.StatusInternalServerError, fmt.Errorf("checkout.HandleWith: %w", err))
				return
			}

			ctx := context.Background()
			tx := conn.Transaction(ctx)
			defer tx.Rollback(ctx)

			order, err := orders.SelectOne(tx, ctx, id)
			if err != nil {
				v1.Error(w, http.StatusInternalServerError, fmt.Errorf("checkout.HandleWith: %w", err))
				return
			}
			tx.Commit(ctx)

			params := &stripe.PaymentIntentConfirmParams{
				ConfirmationToken: stripe.String(req.ConfirmationTokenID),
				ReceiptEmail:      stripe.String(req.Email),
				ReturnURL:         stripe.String("https://localhost" + "/checkout/complete"),
			}
			params.Shipping = &stripe.ShippingDetailsParams{
				Name: stripe.String(req.Shipping.Name),
				Address: &stripe.AddressParams{
					Line1:      stripe.String(req.Shipping.Address.Line1),
					Line2:      stripe.String(req.Shipping.Address.Line2),
					City:       stripe.String(req.Shipping.Address.City),
					PostalCode: stripe.String(req.Shipping.Address.PostalCode),
					Country:    stripe.String(req.Shipping.Address.Country),
				},
			}

			pi, err := paymentintent.Confirm(order.PaymentIntentId, params)
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

			switch pi.Status {
			case stripe.PaymentIntentStatusSucceeded:
				// mark order as paid / authorized-success
				v1.Success(w, http.StatusOK, Response{
					Status: "succeeded",
				})
				m, err := mailer.New()
				if err != nil {
					log.Println(fmt.Errorf("checkout.HandleWith: %w", err))
					return
				}

				msgID, err := m.SendOrderConfirmation(r.Context(), req.Email, order.Id.String())
				if err != nil {
					log.Println(fmt.Errorf("checkout.HandleWith: %w", err))
					return
				}
				_ = msgID
			case stripe.PaymentIntentStatusProcessing:
				// mark order as pending_payment
				v1.Success(w, http.StatusOK, Response{
					Status: "processing",
				})
			case stripe.PaymentIntentStatusRequiresAction:
				resp := Response{
					Status:       "requires_action",
					ClientSecret: pi.ClientSecret,
				}
				if validAction(pi.NextAction) {
					resp.RedirectURL = pi.NextAction.RedirectToURL.URL
				}
				v1.Success(w, http.StatusOK, resp)
			case stripe.PaymentIntentStatusRequiresPaymentMethod:
				v1.Success(w, http.StatusOK, Response{
					Status: "requires_payment_method",
				})
			case stripe.PaymentIntentStatusRequiresCapture:
				v1.Success(w, http.StatusOK, Response{
					Status: "requires_capture",
				})
			case stripe.PaymentIntentStatusCanceled:
				v1.Success(w, http.StatusOK, Response{
					Status: "canceled",
				})
			default:
				v1.Success(w, http.StatusOK, Response{
					Status: string(pi.Status),
				})
			}
		}
	}
}
