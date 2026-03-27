package posters

import (
	"context"
	"fmt"
	"net/http"

	v1 "github.com/cph-collectibles/api/v1"
	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/tables/posters"
)

func HandleWith(conn *db.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ctx := context.Background()
			tx := conn.Transaction(ctx)

			id := r.PathValue("id")
			if id != "" {
				data, err := posters.SelectOne(tx, ctx, id)
				if err != nil {
					tx.Rollback(ctx)
					v1.Error(w, http.StatusInternalServerError, fmt.Errorf("posters.HandleWith: %w", err))
					return
				}
				defer tx.Commit(ctx)
				v1.Success(w, http.StatusOK, data)
			} else {
				data, err := posters.SelectAll(tx, ctx)
				if err != nil {
					tx.Rollback(ctx)
					v1.Error(w, http.StatusInternalServerError, fmt.Errorf("posters.HandleWith: %w", err))
					return
				}
				defer tx.Commit(ctx)
				v1.Success(w, http.StatusOK, data...)
			}
		}
	}
}
