package orderitems

import (
	"context"
	_ "embed"

	"github.com/cph-collectibles/db"
	"github.com/cph-collectibles/db/wrappers"
)

type Data struct {
	OrderId  wrappers.ULID `json:"orderId"`
	PosterId wrappers.ULID `json:"posterId"`
}

//go:embed insert.sql
var insert string

func Insert(db db.Connection, ctx context.Context, data *Data) error {
	_, err := db.Exec(ctx, insert,
		data.OrderId.String(),
		data.PosterId.String(),
	)
	return err
}
