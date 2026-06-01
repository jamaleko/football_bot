package main

import (
	"fmt"
	"strconv"
	"strings"

	"football_bot/internal/footballdata"
	"football_bot/internal/telegram"
)

func ProcessCommands(
	bot *telegram.Client,
	football *footballdata.Client,
	offset *int,
) {

	watch := LoadWatch()

	updates, err := bot.Updates(*offset)
	if err != nil {
		return
	}

	for _, update := range updates.Result {

		*offset = update.UpdateID + 1

		text := update.Message.Text

		switch {

		case text == "/start":

			bot.Send(
				update.Message.Chat.ID,
				"football_bot ready",
			)

		case strings.HasPrefix(text, "/watch"):

			parts := strings.Split(text, " ")

			if len(parts) != 2 {
				continue
			}

			id, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}

			watch[id] = update.Message.Chat.ID

			bot.Send(
				update.Message.Chat.ID,
				fmt.Sprintf(
					"Watching match %d",
					id,
				),
			)

		case strings.HasPrefix(text, "/stop"):

			parts := strings.Split(text, " ")

			if len(parts) != 2 {
				continue
			}

			id, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}

			delete(
				watch,
				id,
			)

			bot.Send(
				update.Message.Chat.ID,
				fmt.Sprintf(
					"Stopped watching %d",
					id,
				),
			)
		}
	}

	SaveWatch(watch)
}
