package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"football_bot/internal/footballdata"
	"football_bot/internal/telegram"
)

func main() {

	bot := telegram.New(os.Getenv("BOT_TOKEN"))
	football := footballdata.New(os.Getenv("TOKEN"))

	fmt.Println("bot started")

	offset := 0

	watchedMatches := map[int]int64{}
	lastScore := map[int]string{}

	go func() {

		for {

			for id, chatID := range watchedMatches {

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
				
				currentScore := home + "-" + away
				
				if lastScore[id] != currentScore {
				
				 lastScore[id] = currentScore
				
				 msg := fmt.Sprintf(
				  "⚽ %s %s-%s %s\nStatus: %s",
				  match.HomeTeam.Name,
				  home,
				  away,
				  match.AwayTeam.Name,
				  match.Status,
				 )
				
				 bot.Send(chatID, msg)
				
				 if match.Status == "FINISHED" {
				  delete(watchedMatches, id)
				  delete(lastScore, id)
				 }
				}
			}

			time.Sleep(9 * time.Minute)
		}
	}()

	for {

		updates, err := bot.Updates(offset)
		if err != nil {
			fmt.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates.Result {

			offset = update.UpdateID + 1

			text := update.Message.Text

			switch {

			case text == "/start":

				bot.Send(
					update.Message.Chat.ID,
					"football_bot ready",
				)

			case text == "/big":

				msg := "🏆 WORLD CUP\n\n"

				wcMatches, err := football.WorldCupMatches()
				if err != nil {
					bot.Send(update.Message.Chat.ID, err.Error())
					continue
				}

				limit := 10
				if len(wcMatches) < limit {
					limit = len(wcMatches)
				}

				for i := 0; i < limit; i++ {

					m := wcMatches[i]

					msg += fmt.Sprintf(
						"%d. %s vs %s\n%s\n/watch %d\n\n",
						i+1,
						m.HomeTeam.Name,
						m.AwayTeam.Name,
						toWib(m.UTCDate),
						m.ID,
					)
				}

				msg += "\n🏆 CHAMPIONS LEAGUE\n\n"

				clMatches, err := football.ChampionsLeagueMatches()
				if err != nil {
					bot.Send(update.Message.Chat.ID, err.Error())
					continue
				}

				for _, m := range clMatches {

					msg += fmt.Sprintf(
						"%s vs %s\n%s\n/watch %d\n\n",
						m.HomeTeam.Name,
						m.AwayTeam.Name,
						toWib(m.UTCDate),
						m.ID,
					)
				}

				bot.Send(update.Message.Chat.ID, msg)

			case strings.HasPrefix(text, "/watch"):

				parts := strings.Split(text, " ")

				if len(parts) != 2 {
					continue
				}

				id, err := strconv.Atoi(parts[1])
				if err != nil {
					continue
				}

				watchedMatches[id] = update.Message.Chat.ID

				bot.Send(
					update.Message.Chat.ID,
					fmt.Sprintf("Watching match %d", id),
				)

			case text == "/stop":

				watchedMatches = map[int]int64{}
				lastScore = map[int]string{}

				bot.Send(
					update.Message.Chat.ID,
					"STOP pressed",
				)
			}
		}

		time.Sleep(2 * time.Second)
	}
}

func toWib(utc string) string {

	t, err := time.Parse(
		time.RFC3339,
		utc,
	)
	if err != nil {
		return utc
	}

	return t.In(
		time.FixedZone("WIB", 7*3600),
	).Format(
		"02 Jan 2006 15:04 WIB",
	)
}
