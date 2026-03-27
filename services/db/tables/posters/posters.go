package posters

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/cph-collectibles/db/wrappers"
	"github.com/jackc/pgx/v5"
)

type meta struct {
	RawId     string        `json:"rawId"`
	CreatedAt wrappers.Time `json:"createdAt"`
	UpdatedAt wrappers.Time `json:"updatedAt"`
}

type origin struct {
	Source string `json:"source"`
	Year   string `json:"year"`
}

type details struct {
	Heading string  `json:"heading"`
	Body    string  `json:"body"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Origin  origin  `json:"origin"`
}

type cost struct {
	RawAmount   float64 `json:"rawAmount"`
	RawVAT      float64 `json:"rawVAT"`
	MinorAmount int64   `json:"minorAmount"`
	MinorVAT    int64   `json:"minorVAT"`
	RawTotal    float64 `json:"rawTotal"`
	MinorTotal  int64   `json:"minorTotal"`
}

type rating struct {
	Rating string `json:"rating"`
	Notes  string `json:"notes"`
}

type file struct {
	Id  wrappers.ULID `json:"id"`
	URL string        `json:"url"`
}

type Data struct {
	Id        wrappers.ULID `json:"id"`
	Meta      meta          `json:"meta"`
	Cost      cost          `json:"cost"`
	Condition rating        `json:"condition"`
	Detail    details       `json:"detail"`
	Files     []file        `json:"files"`
}

func (d *Data) ToSql() string {
	sb := strings.Builder{}
	fmt.Fprintf(&sb, "'%s'::varchar(26),", d.Id.String())
	fmt.Fprintf(&sb, "'%s'::varchar(16),", d.Meta.RawId)
	fmt.Fprintf(&sb, "'%s',", d.Detail.Heading)
	fmt.Fprintf(&sb, "'%s',", d.Detail.Body)
	fmt.Fprintf(&sb, "%.2f::numeric(6, 2),", d.Detail.Width)
	fmt.Fprintf(&sb, "%.2f::numeric(6, 2),", d.Detail.Height)
	fmt.Fprintf(&sb, "'%s'::varchar(32),", d.Detail.Origin.Source)
	fmt.Fprintf(&sb, "'%s'::varchar(16),", d.Detail.Origin.Year)
	fmt.Fprintf(&sb, "'%s'::rating,", d.Condition.Rating)
	fmt.Fprintf(&sb, "'%s'", d.Condition.Notes)
	return sb.String()
}

//go:embed insert.sql
var insert string

func Insert(conn pgx.Tx, ctx context.Context, data *Data) error {
	_, err := conn.Exec(ctx, insert,
		data.Id.String(),
		data.Meta.RawId,
		data.Cost.RawAmount,
		data.Cost.RawVAT,
		data.Detail.Heading,
		data.Detail.Body,
		data.Detail.Width,
		data.Detail.Height,
		data.Detail.Origin.Source,
		data.Detail.Origin.Year,
		data.Condition.Rating,
		data.Condition.Notes,
	)
	return err
}

//go:embed select-all.sql
var all string

func SelectAll(conn pgx.Tx, ctx context.Context) ([]*Data, error) {
	records, err := conn.Query(ctx, all)
	if err != nil {
		return nil, err
	}
	defer records.Close()

	data := []*Data{}
	for records.Next() {
		d := &Data{}
		err = records.Scan(&d.Id, &d.Meta, &d.Cost, &d.Detail, &d.Condition, &d.Files)
		if err != nil {
			return nil, err
		} else {
			data = append(data, d)
		}
	}
	return data, nil
}

//go:embed select-one.sql
var one string

func SelectOne(conn pgx.Tx, ctx context.Context, id string) (Data, error) {
	record := conn.QueryRow(ctx, one, id)
	d := Data{}
	err := record.Scan(&d.Id, &d.Meta, &d.Cost, &d.Detail, &d.Condition, &d.Files)
	if err != nil {
		return d, err
	}
	return d, nil
}

//go:embed query.sql
var query string

func Query(conn pgx.Tx, ctx context.Context, queries []string) ([]*Data, error) {
	records, err := conn.Query(ctx, query, queries)
	if err != nil {
		return nil, err
	}
	defer records.Close()

	data := []*Data{}
	for records.Next() {
		d := &Data{}
		err = records.Scan(&d.Id, &d.Meta, &d.Cost, &d.Detail, &d.Condition, &d.Files)
		if err != nil {
			return nil, err
		} else {
			data = append(data, d)
		}
	}
	return data, nil
}
