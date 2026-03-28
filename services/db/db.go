package db

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func init() {
	if _, ok := os.LookupEnv("DB_HOST"); !ok {
		log.Fatalf("env %s is nil", "DB_HOST")
	}
	if _, ok := os.LookupEnv("DB_PORT"); !ok {
		log.Fatalf("env %s is nil", "DB_PORT")
	}
	if _, ok := os.LookupEnv("DB_USER"); !ok {
		log.Fatalf("env %s is nil", "DB_USER")
	}
	if _, ok := os.LookupEnv("DB_PASSWORD"); !ok {
		log.Fatalf("env %s is nil", "DB_PASSWORD")
	}
	if _, ok := os.LookupEnv("DB_NAME"); !ok {
		log.Fatalf("env %s is nil", "DB_NAME")
	}
}

type Connection interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func LoadConfig() Config {
	return Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}
}

func (c Config) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, c.Port),
		Path:   c.Name,
	}

	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return u.String()
}

type DB struct {
	pool *pgxpool.Pool
}

func (p *DB) Close() {
	p.pool.Close()
}

func (p *DB) Transaction(ctx context.Context) pgx.Tx {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Println(fmt.Errorf("db.Transaction: %w", err))
	}
	return tx
}

func (p *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, sql, args...)
}

func (p *DB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

func (p *DB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func NewPool(cfg Config) (*DB, error) {
	p := new(DB)
	if pool, err := pgxpool.New(context.Background(), cfg.DSN()); err != nil {
		return nil, err
	} else {
		p.pool = pool
	}
	return p, nil
}
