package main

import (
	"fmt"

	"football_bot/internal/footballdata"
	"football_bot/internal/telegram"
)

func CheckMatches(
	bot *telegram.Client,
	football *footballdata.Client,
) {

	watch := LoadWatch()
	state := LoadState()

	for id, chatID := range watch {

		match, err := football.Match(id)
		if err != nil {
			continue
		}

		home := "-"
		away := "-"

		if match.Score.FullTime.Home != nil {
			home = fmt.Sprintf("%d", *match.Score.FullTime.Home)
		}

		if match.Score.FullTime.Away != nil {
			away = fmt.Sprintf("%d", *match.Score.FullTime.Away)
		}

		status := MatchStatus(match)

		current := home + "-" + away + "|" + status

		if state[id] == current {
			continue
		}

		state[id] = current

		msg := fmt.Sprintf(
			"⚽ %s %s-%s %s\nStatus: %s",
			match.HomeTeam.Name,
			home,
			away,
			match.AwayTeam.Name,
			status,
		)

		bot.Send(chatID, msg)

		if match.Status == "FINISHED" {
			delete(watch, id)
			delete(state, id)
		}
	}

	SaveWatch(watch)
	SaveState(state)
}
