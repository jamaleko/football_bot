package main

import (
	"os"

	"football_bot/internal/footballdata"
	"football_bot/internal/telegram"
)

func main() {

	bot := telegram.New(
		os.Getenv("BOT_TOKEN"),
	)

	football := footballdata.New(
		os.Getenv("TOKEN"),
	)

	offset := LoadOffset()

	ProcessCommands(
		bot,
		football,
		&offset,
	)

	SaveOffset(offset)

	CheckMatches(
		bot,
		football,
	)
}
