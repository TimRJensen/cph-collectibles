package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	v1 "github.com/cph-collectibles/api/v1"
	"github.com/cph-collectibles/api/v1/checkout"
	"github.com/cph-collectibles/api/v1/checkout/confirm"
	"github.com/cph-collectibles/api/v1/posters"
	"github.com/cph-collectibles/api/v1/webhook"
	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/seed"
	"github.com/cph-collectibles/db/tables/files"
	"github.com/cph-collectibles/db/wrappers"
	"github.com/oklog/ulid/v2"
	"github.com/stripe/stripe-go/v84"
)

func init() {
	stripe.Key = os.Getenv("STRIPE_PRIVATE_KEY")
}

func seedDB(p *db.Pool) {
	cwd, _ := os.Getwd()

	r, err := zip.OpenReader(filepath.Join(cwd, "data/posters.zip"))
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	rexp := regexp.MustCompile(`\w+/(\w+)[.]png`)
	imgs := map[string]*zip.File{}
	for _, f := range r.File {
		matches := rexp.FindStringSubmatch(f.Name)
		imgs[matches[1]] = f
	}

	data, err := seed.SeedCSV("./data/posters.csv", p)
	if err != nil {
		log.Printf("seedDB: %v", fmt.Errorf("%w", err))
		return
	}

	filedir, err := filepath.Abs(filepath.Join(cwd, "assets"))
	if err != nil {
		log.Fatalf("seedDB: %v", fmt.Errorf("%w", err))
	}

	ctx := context.Background()
	tx := p.Transaction(ctx)
	for _, d := range data {
		id, raw := d.Id.String(), d.Meta.RawId

		path := filepath.Join(filedir, id)
		err := os.MkdirAll(path, 0o777)
		if err != nil {
			log.Fatalf("seedDB: %v", fmt.Errorf("%w", err))
		}

		if v, ok := imgs[d.Meta.RawId]; ok {
			dst, err := os.Create(filepath.Join(path, fmt.Sprintf("%s.png", raw)))
			if err != nil {
				log.Fatalf("seedDB: %v", fmt.Errorf("%w", err))
			}

			src, err := v.Open()
			if err != nil {
				log.Fatalf("seedDB: %v", fmt.Errorf("%w", err))
			}

			_, err = io.Copy(dst, src)
			if err != nil {
				log.Fatalf("seedDB: %v", fmt.Errorf("%w", err))
			}
			dst.Close()
			src.Close()

			data := files.Data{}
			data.Id = (wrappers.ULID)(ulid.Make())
			data.PosterId = (wrappers.ULID)(d.Id)
			data.URL = fmt.Sprintf("%s/assets/%s/%s.png", os.Getenv("FS_URL"), id, raw)
			err = files.Insert(tx, ctx, &data)
			if err != nil {
				log.Fatalf("seedDB: %v", fmt.Errorf("%w", err))
			}
		}
	}
	tx.Commit(ctx)
}

func main() {
	p, err := db.NewPool(db.LoadConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()
	seedDB(p)

	r := http.NewServeMux()
	r.HandleFunc("/api/v1/posters/{id}", posters.HandleWith(p))
	r.HandleFunc("/api/v1/posters", posters.HandleWith(p))
	r.HandleFunc("/api/v1/checkout", checkout.HandleWith(p))
	r.HandleFunc("/api/v1/checkout/{id}", confirm.HandleWith(p))
	r.HandleFunc("/api/v1/checkout/webhook", webhook.HandleWith(p))
	http.ListenAndServe(":8080", v1.WithCors(r))
}
