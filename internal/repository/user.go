package repository

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"expense-tracker-bot/internal/db"
)

type UserRepo struct {
	pool *pgxpool.Pool
	qb   sq.StatementBuilderType
}

func NewUserRepo(conn *db.Conn) *UserRepo {
	return &UserRepo{
		pool: conn.Pool,
		qb:   conn.QB,
	}
}

func (r *UserRepo) GetOrCreate(ctx context.Context, tgID int64, name string) (int64, error) {
	sql, args, err := r.qb.
		Insert("users").
		Columns("tg_id", "first_name").
		Values(tgID, name).
		Suffix(`
		    ON CONFLICT (tg_id) DO UPDATE SET first_name = EXCLUDED.first_name
		    RETURNING id`).
		ToSql()
	if err != nil {
		return 0, err
	}

	var id int64
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
