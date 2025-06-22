package repository

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"expense-tracker-bot/internal/db"
	"expense-tracker-bot/pkg/parser"
)

type StatRow struct {
	Day time.Time
	Sum float64
}

type ExpenseRepo struct {
	pool *pgxpool.Pool
	qb   sq.StatementBuilderType
}

func NewExpenseRepo(conn *db.Conn) *ExpenseRepo {
	return &ExpenseRepo{
		pool: conn.Pool,
		qb:   conn.QB,
	}
}

func (r *ExpenseRepo) Create(ctx context.Context, p *parser.ParsedExpense) (int64, error) {
	sql, args, err := r.qb.
		Insert("transactions").
		Columns("user_id", "amount", "currency", "note", "created_at").
		Values(p.UserID, p.Amount, p.Currency, p.Note, time.Now()).
		Suffix("RETURNING id").
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

func (r *ExpenseRepo) DeleteLast(ctx context.Context, uid int64) error {
	sql, args, err := r.qb.
		Delete("transactions").
		Where(sq.Expr(`
		    id = (
		      SELECT id FROM transactions
		      WHERE  user_id = ?
		      ORDER  BY id DESC
		      LIMIT 1
		    )`, uid)).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	var id int64
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return err
	}
	return nil
}

func (r *ExpenseRepo) Stats(ctx context.Context, uid int64, from, to time.Time) ([]StatRow, error) {
	sql, args, err := r.qb.
		Select("date_trunc('day', created_at) AS day", "SUM(amount)").
		From("transactions").
		Where(sq.And{
			sq.Eq{"user_id": uid},
			sq.GtOrEq{"created_at": from},
			sq.LtOrEq{"created_at": to},
		}).
		GroupBy("day").
		OrderBy("day").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []StatRow
	for rows.Next() {
		var s StatRow
		if err := rows.Scan(&s.Day, &s.Sum); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type Expense struct {
	ID       int64
	Date     time.Time
	Amount   float64
	Currency string
	Note     string
}

func (r *ExpenseRepo) ListByPeriod(ctx context.Context,
	uid int64, from, to time.Time) ([]Expense, error) {

	sql, args, err := r.qb.
		Select("id", "created_at", "amount", "currency", "note").
		From("transactions").
		Where(sq.And{
			sq.Eq{"user_id": uid},
			sq.GtOrEq{"created_at": from},
			sq.LtOrEq{"created_at": to},
		}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Expense
	for rows.Next() {
		var e Expense
		if err := rows.Scan(&e.ID, &e.Date, &e.Amount, &e.Currency, &e.Note); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *ExpenseRepo) StatsByPeriodCurrency(
	ctx context.Context, uid int64, from, to time.Time,
	cur string) (map[string]float64, error) {

	q := r.qb.
		Select("currency", "SUM(amount)").
		From("transactions").
		Where(sq.And{
			sq.Eq{"user_id": uid},
			sq.GtOrEq{"created_at": from},
			sq.LtOrEq{"created_at": to},
		}).
		GroupBy("currency")

	if cur != "" {
		q = q.Where(sq.Eq{"currency": cur})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]float64)
	for rows.Next() {
		var c string
		var s float64
		if err := rows.Scan(&c, &s); err != nil {
			return nil, err
		}
		out[c] = s
	}
	return out, rows.Err()
}
