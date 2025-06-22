package main

import (
	"context"
	"expense-tracker-bot/internal/app"
	"log"
	"os/signal"
	"syscall"

	"expense-tracker-bot/internal/config"
)

func main() {
	cfg := config.MustLoad()

	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := a.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
