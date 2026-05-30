package main

import (
	"fmt"
	"os"
	"time"

	"football_bot/internal/telegram"
)

func main() {

	bot := telegram.New(
		os.Getenv("BOT_TOKEN"),
	)

	fmt.Println("bot started")

	offset := 0

	for {

		updates, err := bot.Updates(offset)
		if err != nil {

			fmt.Println(err)

			time.Sleep(
				5 * time.Second,
			)

			continue
		}

		for _, update := range updates.Result {

			offset = update.UpdateID + 1

			text := update.Message.Text

			switch text {

			case "/start":

				bot.Send(
					update.Message.Chat.ID,
					"football_bot ready",
				)

			case "/big":

				bot.Send(
					update.Message.Chat.ID,
					"BIG pressed",
				)

			case "/stop":

				bot.Send(
					update.Message.Chat.ID,
					"STOP pressed",
				)
			}
		}

		time.Sleep(
			2 * time.Second,
		)
	}
}
