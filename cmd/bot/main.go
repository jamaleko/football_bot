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
	watchedMatches := map[int]int64{} // simpan match ID yang di-watch

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
			   "%d. %s vs %s\n/watch %d\n\n",
			   i+1,
			   m.HomeTeam.Name,
			   m.AwayTeam.Name,
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
			   "%s vs %s\n/watch %d\n\n",
			   m.HomeTeam.Name,
			   m.AwayTeam.Name,
			   m.ID,
			  )
			 }
			
			 bot.Send(update.Message.Chat.ID, msg)

			case strings.HasPrefix(text, "/watch"):
				parts := strings.Split(text, " ")
				if len(parts) == 2 {
					id, err := strconv.Atoi(parts[1])
					if err == nil {
						watchedMatches[id] = update.Message.Chat.ID
						bot.Send(update.Message.Chat.ID, fmt.Sprintf("Watching match %d", id))
					}
				}

			case text == "/stop":
				bot.Send(update.Message.Chat.ID, "STOP pressed")
				watchedMatches = map[int]int64{} // reset semua watch
			}
		}

		// Auto refresh untuk watched match
		for id, chatID := range watchedMatches {
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
				bot.Send(chatID, msg)
			}
		}

		time.Sleep(9 * time.Minute)
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
