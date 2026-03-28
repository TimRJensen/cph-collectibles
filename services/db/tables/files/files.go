package files

import (
	"context"
	_ "embed"

	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/wrappers"
)

type Data struct {
	Id        wrappers.ULID `json:"id"`
	PosterId  wrappers.ULID `json:"posterId"`
	URL       string        `json:"url"`
	CreatedAt wrappers.Time `json:"createdAt"`
	UpdatedAt wrappers.Time `json:"updatedAt"`
}

//go:embed insert.sql
var insert string

func Insert(db db.Connection, ctx context.Context, data *Data) error {
	_, err := db.Exec(ctx, insert,
		data.Id.String(),
		data.PosterId.String(),
		data.URL,
	)
	return err
}
