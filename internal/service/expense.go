package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strings"
	"time"

	"expense-tracker-bot/internal/repository"
	"expense-tracker-bot/pkg/parser"
)

type ExpenseService struct {
	expRepo  *repository.ExpenseRepo
	userRepo *repository.UserRepo
}

func NewExpenseService(exp *repository.ExpenseRepo, usr *repository.UserRepo) *ExpenseService {
	return &ExpenseService{expRepo: exp, userRepo: usr}
}

func (s *ExpenseService) RegisterUser(ctx context.Context, tgID int64, name string) error {
	_, err := s.userRepo.GetOrCreate(ctx, tgID, name)
	return err
}

func (s *ExpenseService) AddExpense(ctx context.Context,
	tgID int64, firstName, raw string) (string, error) {

	uid, err := s.userRepo.GetOrCreate(ctx, tgID, firstName)
	if err != nil {
		return "", err
	}

	pe, err := parser.ParseExpense(raw)
	if err != nil {
		return "", err
	}
	pe.UserID = uid

	if _, err = s.expRepo.Create(ctx, &pe); err != nil {
		return "", err
	}
	return fmt.Sprintf("Записал %s — %.2f %s", pe.Note, pe.Amount, pe.Currency), nil
}

func (s *ExpenseService) UndoLast(ctx context.Context,
	tgID int64, firstName string) (string, error) {

	uid, err := s.internalID(ctx, tgID, firstName)
	if err != nil {
		return "", err
	}

	err = s.expRepo.DeleteLast(ctx, uid)
	switch {
	case err == nil:
		return "Последняя запись удалена ✅", nil
	case errors.Is(err, pgx.ErrNoRows):
		return "Нет записей для удаления.", nil
	default:
		return "", err
	}
}

func (s *ExpenseService) internalID(ctx context.Context, tgID int64, name string) (int64, error) {
	return s.userRepo.GetOrCreate(ctx, tgID, name)
}

func (s *ExpenseService) Stats(ctx context.Context, userID int64, firstName, period string) (string, error) {
	uid, err := s.internalID(ctx, userID, firstName)
	if err != nil {
		return "", err
	}

	to := time.Now()
	var from time.Time
	switch period {
	case "day":
		from = to.AddDate(0, 0, -1)
	case "week":
		from = to.AddDate(0, 0, -7)
	default:
		from = to.AddDate(0, -1, 0)
	}
	rows, err := s.expRepo.Stats(ctx, uid, from, to)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "Пока пусто 🚀", nil
	}
	var total float64
	for _, r := range rows {
		total += r.Sum
	}
	return fmt.Sprintf("С %s по %s всего: %.2f €", from.Format("02.01"), to.Format("02.01"), total), nil
}

func (s *ExpenseService) List(
	ctx context.Context, tgID int64, name, period string) (string, error) {

	uid, err := s.internalID(ctx, tgID, name)
	if err != nil {
		return "", err
	}

	from, to := periodWindow(period)
	rows, err := s.expRepo.ListByPeriod(ctx, uid, from, to)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "Пока пусто 🚀", nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Покупки с %s по %s:\n",
		from.Format("02.01"), to.Format("02.01"))
	for _, e := range rows {
		fmt.Fprintf(&b, "• %s — %.2f %s (%s)\n",
			e.Date.Format("02.01 15:04"),
			e.Amount, e.Currency, e.Note)
	}
	return b.String(), nil
}

func (s *ExpenseService) StatsCurrency(
	ctx context.Context, tgID int64, name, period, cur string) (string, error) {

	uid, err := s.internalID(ctx, tgID, name)
	if err != nil {
		return "", err
	}

	from, to := periodWindow(period)
	m, err := s.expRepo.StatsByPeriodCurrency(ctx, uid, from, to, cur)
	if err != nil {
		return "", err
	}
	if len(m) == 0 {
		return "Пока пусто 🚀", nil
	}

	if cur != "" {
		return fmt.Sprintf("За период %s–%s: %.2f %s",
			from.Format("02.01"), to.Format("02.01"), m[cur], cur), nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Статистика с %s по %s:\n", from.Format("02.01"), to.Format("02.01"))
	for c, s := range m {
		fmt.Fprintf(&b, "• %s — %.2f\n", c, s)
	}
	return b.String(), nil
}

func periodWindow(p string) (time.Time, time.Time) {
	to := time.Now()
	switch p {
	case "day":
		return to.AddDate(0, 0, -1), to
	case "week":
		return to.AddDate(0, 0, -7), to
	default:
		return to.AddDate(0, -1, 0), to
	}
}
