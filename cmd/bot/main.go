package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
)

func main() {

	b, err := bot.New(os.Getenv("BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/start",
		bot.MatchTypePrefix,
		func(ctx context.Context, b *bot.Bot, update *bot.Update) {

			b.SendMessage(
				ctx,
				&bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "football_bot ready",
				},
			)
		},
	)

	fmt.Println("bot started")

	b.Start(context.Background())
}
