package orderitems

import (
	"context"
	_ "embed"

	"github.com/cph-collectibles/db/wrappers"
	"github.com/jackc/pgx/v5"
)

type Data struct {
	OrderId  wrappers.ULID `json:"orderId"`
	PosterId wrappers.ULID `json:"posterId"`
}

//go:embed insert.sql
var insert string

func Insert(conn pgx.Tx, ctx context.Context, data *Data) error {
	_, err := conn.Exec(ctx, insert,
		data.OrderId.String(),
		data.PosterId.String(),
	)
	return err
}
