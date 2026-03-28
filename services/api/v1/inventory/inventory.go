package inventory

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"slices"
	"strconv"

	v1 "github.com/cph-collectibles/api/v1"
	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/tables/inventory"
)

func HandleWith(db *db.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("foo", r.URL.Query().Get("random"))

		switch r.Method {
		case http.MethodGet:
			ctx := context.Background()

			id := r.PathValue("id")
			if id != "" {
				data, err := inventory.SelectOne(db, ctx, id)
				if err != nil {
					v1.Error(w, http.StatusInternalServerError, fmt.Errorf("inventory.HandleWith: %w", err))
					return
				}
				v1.Success(w, http.StatusOK, data)
				return
			}
			q := r.URL.Query()

			switch {
			case q.Has("random"):
				n, err := strconv.ParseInt(q.Get("random"), 10, 32)
				if err != nil {
					v1.Error(w, http.StatusBadRequest, fmt.Errorf("inventory.HandleWith: %w", err))
					return
				}
				data, err := inventory.SelectAll(db, ctx)
				if err != nil {
					v1.Error(w, http.StatusInternalServerError, fmt.Errorf("inventory.HandleWith: %w", err))
					return
				}
				seen := []*inventory.Data{}
				for n > -1 {
					idx := rand.Intn(len(data))
					if slices.Contains(seen, data[idx]) {
						continue
					}
					seen = append(seen, data[idx])
					n -= 1
				}
				v1.Success(w, http.StatusOK, seen...)
			case q.Has("search"):
				data, err := inventory.Query(db, ctx, q.Get("search"))
				if err != nil {
					v1.Error(w, http.StatusInternalServerError, fmt.Errorf("inventory.HandleWith: %w", err))
					return
				}

				v1.Success(w, http.StatusOK, data...)
			default:
				data, err := inventory.SelectAll(db, ctx)
				if err != nil {
					v1.Error(w, http.StatusInternalServerError, fmt.Errorf("inventory.HandleWith: %w", err))
					return
				}
				v1.Success(w, http.StatusOK, data...)
			}
		}
	}
}
