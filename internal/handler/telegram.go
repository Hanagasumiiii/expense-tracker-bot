package handler

import (
	"context"
	"log"
	"strings"

	"expense-tracker-bot/internal/service"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	api     *tgbot.BotAPI
	svc     *service.ExpenseService
	updates tgbot.UpdatesChannel
}

func NewTelegram(token string, svc *service.ExpenseService) (*TelegramBot, error) {
	api, err := tgbot.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	u := tgbot.NewUpdate(0)
	u.Timeout = 30
	updates := api.GetUpdatesChan(u)

	return &TelegramBot{
		api:     api,
		svc:     svc,
		updates: updates,
	}, nil
}

func (b *TelegramBot) Run(ctx context.Context) error {
	log.Printf("bot authorised as @%s", b.api.Self.UserName)

	for {
		select {
		case <-ctx.Done():
			return nil
		case upd := <-b.updates:
			if upd.Message == nil {
				continue
			}
			go b.handleMessage(ctx, upd.Message)
		}
	}
}

func (b *TelegramBot) Stop() {
	b.api.StopReceivingUpdates()
}

func (b *TelegramBot) handleMessage(ctx context.Context, m *tgbot.Message) {
	uid := m.From.ID
	text := m.Text

	var resp string
	var err error

	switch {
	case text == "/start":
		if err := b.svc.RegisterUser(ctx, uid, m.From.FirstName); err != nil {
			resp = "Ошибка регистрации: " + err.Error()
		} else {
			resp = "Привет! Шли: \n- 12.5 RUB кофе\n/list week\n/stats week USD"
		}

	case hasPrefix(text, "/add", "-"):
		rest := dropCmd(text)
		resp, err = b.svc.AddExpense(ctx,
			uid,
			m.From.FirstName,
			rest)

	case hasPrefix(text, "/undo"):
		resp, err = b.svc.UndoLast(ctx,
			uid,
			m.From.FirstName)

	case hasPrefix(text, "/list"):
		rest := strings.Fields(dropCmd(text))
		period := "week"
		if len(rest) > 0 {
			period = rest[0]
		}
		resp, err = b.svc.List(ctx, m.From.ID, m.From.FirstName, period)

	case hasPrefix(text, "/stats"):
		parts := strings.Fields(dropCmd(text))
		per, cur := "day", ""
		if len(parts) >= 1 {
			per = parts[0]
		}
		if len(parts) == 2 {
			cur = strings.ToUpper(parts[1])
		}
		resp, err = b.svc.StatsCurrency(ctx, m.From.ID, m.From.FirstName, per, cur)

	default:
		resp = "Не понял 🤖. Попробуй /add, /undo, /stats."
	}

	if err != nil {
		resp = "Ошибка: " + err.Error()
	}

	msg := tgbot.NewMessage(m.Chat.ID, resp)
	b.api.Send(msg)
}

func hasPrefix(s string, cmds ...string) bool {
	for _, c := range cmds {
		if len(s) >= len(c) && s[:len(c)] == c {
			return true
		}
	}
	return false
}

func dropCmd(s string) string {
	if i := indexAny(s, ' ', '\n', '\t'); i >= 0 {
		return s[i+1:]
	}
	return ""
}

func indexAny(s string, cs ...rune) int {
	for i, r := range s {
		for _, c := range cs {
			if r == c {
				return i
			}
		}
	}
	return -1
}
