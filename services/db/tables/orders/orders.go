package orders

import (
	"context"
	_ "embed"

	"github.com/cph-collectibles/db/wrappers"
	"github.com/jackc/pgx/v5"
)

type Data struct {
	Id              wrappers.ULID `json:"id"`
	PaymentIntentId string        `json:"paymentIntentId"`
	Status          string        `json:"status"`
	Total           int64         `json:"total"`
	CreatedAt       wrappers.Time `json:"createdAt"`
	UpdatedAt       wrappers.Time `json:"updatedAt"`
}

//go:embed insert.sql
var insert string

func Insert(conn pgx.Tx, ctx context.Context, data *Data) error {
	_, err := conn.Exec(ctx, insert,
		data.Id.String(),
		data.Total,
	)
	return err
}

//go:embed update.sql
var update string

func Update(conn pgx.Tx, ctx context.Context, data *Data) error {
	_, err := conn.Exec(ctx, update, data.Id.String(), data.PaymentIntentId, data.Status)
	return err
}

//go:embed select-one.sql
var one string

func SelectOne(conn pgx.Tx, ctx context.Context, id string) (Data, error) {
	record := conn.QueryRow(ctx, one, id)
	d := Data{}
	err := record.Scan(&d.Id, &d.PaymentIntentId, &d.Status, &d.Total, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return d, err
	}
	return d, nil
}
