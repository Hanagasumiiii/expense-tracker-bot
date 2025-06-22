package app

import (
	"context"
	"log"

	"expense-tracker-bot/internal/config"
	"expense-tracker-bot/internal/db"
	"expense-tracker-bot/internal/handler"
	"expense-tracker-bot/internal/repository"
	"expense-tracker-bot/internal/service"
)

type App struct {
	dbConn *db.Conn
	bot    *handler.TelegramBot
}

func New(cfg config.Config) (*App, error) {
	ctx := context.Background()

	dbConn, err := db.New(ctx, db.Config{
		Host: cfg.DBHost,
		Port: cfg.DBPort,
		User: cfg.DBUser,
		Pass: cfg.DBPass,
		Name: cfg.DBName,
	})
	if err != nil {
		return nil, err
	}

	usrRepo := repository.NewUserRepo(dbConn)
	expRepo := repository.NewExpenseRepo(dbConn)
	svc := service.NewExpenseService(expRepo, usrRepo)
	bot, err := handler.NewTelegram(cfg.TgToken, svc)
	if err != nil {
		dbConn.Close()
		return nil, err
	}

	return &App{dbConn: dbConn, bot: bot}, nil
}

func (a *App) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		a.Shutdown()
	}()
	return a.bot.Run(ctx)
}

func (a *App) Shutdown() {
	log.Println("shutting down...")
	a.bot.Stop()
	a.dbConn.Close()
}
