package main

import (
	"fmt"
	"os"
	"time"
	"strconv"
	"strings"

	"football_bot/internal/telegram"
	"football_bot/internal/footballdata"
)

func main() {

	bot := telegram.New(os.Getenv("BOT_TOKEN"))
	football := footballdata.New(os.Getenv("TOKEN"))

	fmt.Println("bot started")

	offset := 0
	watchedMatches := map[int]bool{} // simpan match ID yang di-watch

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
				bot.Send(update.Message.Chat.ID, "football_bot ready")

			case text == "/big":
				msg := "BIG MATCHES:\n\n"

				// WC
				wcMatches, _ := football.WorldCupMatches()
				for i, m := range wcMatches {
					msg += fmt.Sprintf("%d. %s vs %s\n/watch %d\n\n", i+1, m.HomeTeam.Name, m.AwayTeam.Name, m.ID)
				}

				// CL
				clMatches, _ := football.ChampionsLeagueMatches()
				for i, m := range clMatches {
					msg += fmt.Sprintf("%d. %s vs %s\n/watch %d\n\n", i+1, m.HomeTeam.Name, m.AwayTeam.Name, m.ID)
				}

				bot.Send(update.Message.Chat.ID, msg)

			case strings.HasPrefix(text, "/watch"):
				parts := strings.Split(text, " ")
				if len(parts) == 2 {
					id, err := strconv.Atoi(parts[1])
					if err == nil {
						watchedMatches[id] = true
						bot.Send(update.Message.Chat.ID, fmt.Sprintf("Watching match %d", id))
					}
				}

			case text == "/stop":
				bot.Send(update.Message.Chat.ID, "STOP pressed")
				watchedMatches = map[int]bool{} // reset semua watch
			}
		}

		// Auto refresh untuk watched match
		for id := range watchedMatches {
			match, err := football.Match(id)
			if err != nil {
				continue
			}

			// jika status berubah, kirim ke Telegram
			if match.Status == "IN_PLAY" || match.Status == "PAUSED" || match.Status == "FINISHED" {
				msg := fmt.Sprintf(
					"%s vs %s\nStatus: %s\n",
					match.HomeTeam.Name,
					match.AwayTeam.Name,
					match.Status,
				)
				bot.Send(os.Getenv("TELEGRAM_CHAT_ID"), msg)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
