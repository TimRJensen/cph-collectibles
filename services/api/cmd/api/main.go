package main

import (
	"log"
	"net/http"
	"os"

	v1 "github.com/cph-collectibles/api/v1"
	"github.com/cph-collectibles/api/v1/checkout"
	"github.com/cph-collectibles/api/v1/checkout/confirm"
	"github.com/cph-collectibles/api/v1/posters"
	"github.com/cph-collectibles/api/v1/webhook"
	"github.com/cph-collectibles/db"
	"github.com/stripe/stripe-go/v84"
)

func init() {
	stripe.Key = os.Getenv("STRIPE_PRIVATE_KEY")
}

func main() {
	p, err := db.NewPool(db.LoadConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	r := http.NewServeMux()
	r.HandleFunc("/api/v1/posters/{id}", posters.HandleWith(p))
	r.HandleFunc("/api/v1/posters", posters.HandleWith(p))
	r.HandleFunc("/api/v1/checkout", checkout.HandleWith(p))
	r.HandleFunc("/api/v1/checkout/{id}", confirm.HandleWith(p))
	r.HandleFunc("/api/v1/checkout/webhook", webhook.HandleWith(p))
	http.ListenAndServe(":8080", v1.WithCors(r))
}
