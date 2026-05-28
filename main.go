package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const BOT_TOKEN = "8727181698:AAGAehZUWyQmh7CBiWNt_x6Xsevh2hrcuCE"

type Response struct {
	Events []Event `json:"events"`
}

type Event struct {
	Tournament Tournament `json:"tournament"`
	HomeTeam   Team       `json:"homeTeam"`
	AwayTeam   Team       `json:"awayTeam"`
	HomeScore  Score      `json:"homeScore"`
	AwayScore  Score      `json:"awayScore"`
	Status     Status     `json:"status"`
	Time       Time       `json:"time"`
}

type Tournament struct {
	Name string `json:"name"`
}

type Team struct {
	Name string `json:"name"`
}

type Score struct {
	Current int `json:"current"`
}

type Status struct {
	Type string `json:"type"`
}

type Time struct {
	CurrentPeriodStartTimestamp int64 `json:"currentPeriodStartTimestamp"`
}

func getLiveMatches() string {

	url := "https://www.sofascore.com/api/v1/sport/football/events/live"

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "❌ Request error"
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "❌ API error"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "❌ Read body error"
	}

	// DEBUG
	fmt.Println(string(body))

	var data Response

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "❌ JSON parse error"
	}

	if len(data.Events) == 0 {
		return "❌ Tidak ada live match"
	}

	var result strings.Builder

	result.WriteString("⚽ LIVE MATCH\n\n")

	count := 0

	for _, match := range data.Events {

		// hanya ambil yang live
		if match.Status.Type == "inprogress" {

			count++

			result.WriteString(
				fmt.Sprintf(
					"🏆 %s\n"+
						"%s %d - %d %s\n\n",

					match.Tournament.Name,
					match.HomeTeam.Name,
					match.HomeScore.Current,
					match.AwayScore.Current,
					match.AwayTeam.Name,
				),
			)
		}
	}

	if count == 0 {
		return "❌ Tidak ada live match"
	}

	return result.String()
}

func getBigMatch() string {

	url := "https://www.sofascore.com/api/v1/sport/football/events/live"

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "❌ Request error"
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "❌ API error"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "❌ Read body error"
	}

	var data Response

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "❌ JSON parse error"
	}

	if len(data.Events) == 0 {
		return "🔥 BIG MATCH coming soon"
	}

	bigLeagues := []string{
		"Premier League",
		"LaLiga",
		"Serie A",
		"Bundesliga",
		"Ligue 1",
		"Champions League",
	}

	var result strings.Builder

	result.WriteString("🔥 BIG MATCH\n\n")

	found := false

	for _, match := range data.Events {

		for _, league := range bigLeagues {

			if strings.Contains(
				strings.ToLower(match.Tournament.Name),
				strings.ToLower(league),
			) {

				found = true

				result.WriteString(
					fmt.Sprintf(
						"🏆 %s\n"+
							"%s %d - %d %s\n\n",

						match.Tournament.Name,
						match.HomeTeam.Name,
						match.HomeScore.Current,
						match.AwayScore.Current,
						match.AwayTeam.Name,
					),
				)
			}
		}
	}

	if !found {
		return "🔥 BIG MATCH coming soon"
	}

	return result.String()
}

func main() {

	bot, err := tgbotapi.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("BOT RUNNING...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil {
			continue
		}

		text := strings.ToUpper(update.Message.Text)

		var reply string

		switch text {

		case "/START":

			reply = `⚽ FOOTBALL BOT

/LIVE
/BIG
/RANDOM
/REFRESH
/STOP`

		case "/LIVE":

			reply = getLiveMatches()

		case "/BIG":

			reply = getBigMatch()

		case "/RANDOM":

			reply = getLiveMatches()

		case "/REFRESH":

			reply = "🔄 Refreshed\n\n" + getLiveMatches()

		case "/STOP":

			reply = "🛑 Bot stopped"

		default:

			reply = "❌ Command tidak dikenal"
		}

		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			reply,
		)

		bot.Send(msg)
	}
}
