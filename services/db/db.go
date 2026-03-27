package db

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

type Pool struct {
	pool *pgxpool.Pool
}

func (p *Pool) Close() {
	p.pool.Close()
}

func (p *Pool) Transaction(ctx context.Context) pgx.Tx {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Println(fmt.Errorf("db.Transaction: %w", err))
	}
	return tx
}

func NewPool(cfg Config) (*Pool, error) {
	p := new(Pool)
	if pool, err := pgxpool.New(context.Background(), cfg.DSN()); err != nil {
		return nil, err
	} else {
		p.pool = pool
	}
	return p, nil
}

/*


 */
