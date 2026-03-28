package main

import (
	"log"
	"net/http"
	"os"

	v1 "github.com/cph-collectibles/api/v1"
	"github.com/cph-collectibles/api/v1/checkout"
	"github.com/cph-collectibles/api/v1/checkout/confirm"
	"github.com/cph-collectibles/api/v1/inventory"
	"github.com/cph-collectibles/api/v1/webhook"
	"github.com/cph-collectibles/db"
)

func init() {
	if _, ok := os.LookupEnv("STRIPE_PRIVATE_KEY"); !ok {
		log.Fatalf("env %s is nil", "STRIPE_PRIVATE_KEY")
	}
	if _, ok := os.LookupEnv("STRIPE_PUBLISH_KEY"); !ok {
		log.Fatalf("env %s is nil", "STRIPE_PUBLISH_KEY")
	}
	if _, ok := os.LookupEnv("STRIPE_WEBHOOK_KEY"); !ok {
		log.Fatalf("env %s is nil", "STRIPE_WEBHOOK_KEY")
	}
	if _, ok := os.LookupEnv("RESEND_API_KEY"); !ok {
		log.Fatalf("env %s is nil", "RESEND_API_KEY")
	}
	if _, ok := os.LookupEnv("RESEND_FROM"); !ok {
		log.Fatalf("env %s is nil", "RESEND_FROM")
	}
	if _, ok := os.LookupEnv("HOST"); !ok {
		log.Fatalf("env %s is nil", "HOST")
	}
}

func main() {
	p, err := db.NewPool(db.LoadConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	r := http.NewServeMux()
	r.HandleFunc("/api/v1/inventory/{id}", inventory.HandleWith(p))
	r.HandleFunc("/api/v1/inventory", inventory.HandleWith(p))
	r.HandleFunc("/api/v1/checkout", checkout.HandleWith(p))
	r.HandleFunc("/api/v1/checkout/{id}", confirm.HandleWith(p))
	r.HandleFunc("/api/v1/checkout/webhook", webhook.HandleWith(p))
	http.ListenAndServe(":8080", v1.WithCors(r))
}
