package db

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Pass     string
	Name     string
	MaxConns int32

	ConnLifetime time.Duration
}

type Conn struct {
	*pgxpool.Pool
	QB sq.StatementBuilderType
}

func New(ctx context.Context, c Config) (*Conn, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Pass, c.Host, c.Port, c.Name)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}
	cfg.MaxConns = 4
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping DB: %w", err)
	}

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return &Conn{Pool: pool, QB: qb}, nil
}

func (p *Conn) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
