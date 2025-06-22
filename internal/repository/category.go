package repository

import (
	"context"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"expense-tracker-bot/internal/db"
)

type CategoryRepo struct {
	pool *pgxpool.Pool
	qb   sq.StatementBuilderType
}

func NewCategoryRepo(conn *db.Conn) *CategoryRepo {
	return &CategoryRepo{
		pool: conn.Pool,
		qb:   conn.QB,
	}
}

func (r *CategoryRepo) GetOrCreate(ctx context.Context, uid int64, name string) (int64, error) {
	name = strings.ToLower(strings.TrimSpace(name))

	sqlSel, argsSel, _ := r.qb.
		Select("id").
		From("categories").
		Where(sq.Eq{"user_id": uid, "name": name}).
		Limit(1).
		ToSql()
	var id int64
	if err := r.pool.QueryRow(ctx, sqlSel, argsSel...).Scan(&id); err == nil {
		return id, nil
	}
	
	sqlIns, argsIns, err := r.qb.
		Insert("categories").
		Columns("user_id", "name").
		Values(uid, name).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, err
	}
	if err := r.pool.QueryRow(ctx, sqlIns, argsIns...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
