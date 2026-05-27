package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	baseURL = "https://v3.football.api-sports.io"
)

var (
	selectedMatch int
	isWatching    bool
	lastGoals     int
	lastUpdateID  int
)

var priorityLeagues = []string{
	"Premier League",
	"La Liga",
	"Serie A",
	"Bundesliga",
	"Ligue 1",

	"UEFA Champions League",
	"UEFA Europa League",

	"FIFA World Cup",
	"UEFA Euro",
	"Copa America",

	"Liga 1",
}

type FixtureResponse struct {
	Response []struct {

		League struct {
			Name    string `json:"name"`
			Country string `json:"country"`
		} `json:"league"`

		Fixture struct {
			ID int `json:"id"`

			Status struct {
				Elapsed int `json:"elapsed"`
			} `json:"status"`

			Date string `json:"date"`
		} `json:"fixture"`

		Teams struct {
			Home struct {
				Name string `json:"name"`
			} `json:"home"`

			Away struct {
				Name string `json:"name"`
			} `json:"away"`
		} `json:"teams"`

		Goals struct {
			Home int `json:"home"`
			Away int `json:"away"`
		} `json:"goals"`

		Statistics []struct {

			Team struct {
				Name string `json:"name"`
			} `json:"team"`

			Statistics []struct {
				Type  string      `json:"type"`
				Value interface{} `json:"value"`
			} `json:"statistics"`

		} `json:"statistics"`

	} `json:"response"`
}

type TelegramResponse struct {
	Ok bool `json:"ok"`

	Result []struct {

		UpdateID int `json:"update_id"`

		Message struct {

			Text string `json:"text"`

			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`

		} `json:"message"`

	} `json:"result"`
}

type MatchItem struct {
	ID    int
	Title string
}

var currentMatches []MatchItem

func sendTelegram(message string) {

	token := os.Getenv("BOT_TOKEN")
	chatID := os.Getenv("CHAT_ID")

	url :=
		fmt.Sprintf(
			"https://api.telegram.org/bot%s/sendMessage",
			token,
		)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
	}

	body, _ := json.Marshal(payload)

	http.Post(
		url,
		"application/json",
		bytes.NewBuffer(body),
	)
}

func apiRequest(url string) ([]byte, error) {

	req, _ := http.NewRequest(
		"GET",
		url,
		nil,
	)

	req.Header.Set(
		"x-apisports-key",
		os.Getenv("FOOTBALL_API_KEY"),
	)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func isPriorityLeague(name string) bool {

	for _, league := range priorityLeagues {

		if strings.Contains(
			strings.ToLower(name),
			strings.ToLower(league),
		) {
			return true
		}
	}

	return false
}

func sendMenu() {

	message :=
		"⚽ FOOTBALL BOT\n\n" +
			"/startbot\n" +
			"/random\n" +
			"/stopbot"

	sendTelegram(message)
}

func sendUpcomingMatches() {

	currentMatches = nil

	now := time.Now()

	message :=
		"📅 BIG MATCHES\n\n"

	index := 1

	for day := 0; day < 3; day++ {

		date :=
			now.AddDate(
				0,
				0,
				day,
			).Format("2006-01-02")

		url :=
			fmt.Sprintf(
				"%s/fixtures?date=%s",
				baseURL,
				date,
			)

		body, err :=
			apiRequest(url)

		if err != nil {
			continue
		}

		var data FixtureResponse

		json.Unmarshal(
			body,
			&data,
		)

		for _, match := range data.Response {

			if !isPriorityLeague(
				match.League.Name,
			) {
				continue
			}

			matchTime, _ :=
				time.Parse(
					time.RFC3339,
					match.Fixture.Date,
				)

			if matchTime.Before(now) {
				continue
			}

			wib :=
				matchTime.In(
					time.FixedZone(
						"WIB",
						7*3600,
					),
				)

			title :=
				fmt.Sprintf(
					"%d. 🌍 %s\n"+
						"🏆 %s\n"+
						"⚽ %s vs %s\n"+
						"🕒 %s WIB\n\n",

					index,

					match.League.Country,
					match.League.Name,

					match.Teams.Home.Name,
					match.Teams.Away.Name,

					wib.Format(
						"02 Jan 15:04",
					),
				)

			message += title

			currentMatches =
				append(
					currentMatches,
					MatchItem{
						ID:    match.Fixture.ID,
						Title: title,
					},
				)

			index++

			if index > 10 {
				break
			}
		}
	}

	message +=
		"\nReply nomor match.\nContoh: 1"

	sendTelegram(message)
}

func watchRandomLiveMatch() {

	url :=
		fmt.Sprintf(
			"%s/fixtures?live=all",
			baseURL,
		)

	body, err :=
		apiRequest(url)

	if err != nil {
		return
	}

	var data FixtureResponse

	json.Unmarshal(
		body,
		&data,
	)

	if len(data.Response) == 0 {

		sendTelegram(
			"❌ Tidak ada live match",
		)

		return
	}

	randomIndex :=
		rand.Intn(
			len(data.Response),
		)

	match :=
		data.Response[randomIndex]

	selectedMatch =
		match.Fixture.ID

	lastGoals = 0

	isWatching = true

	sendTelegram(
		fmt.Sprintf(
			"🎲 RANDOM MATCH\n\n"+
				"🌍 %s\n"+
				"🏆 %s\n"+
				"⚽ %s %d - %d %s",

			match.League.Country,
			match.League.Name,

			match.Teams.Home.Name,
			match.Goals.Home,

			match.Goals.Away,
			match.Teams.Away.Name,
		),
	)

	go watchLoop()
}

func watchLoop() {

	for isWatching {

		watchMatch()

		time.Sleep(
			5 * time.Minute,
		)
	}
}

func watchMatch() {

	if selectedMatch == 0 {
		return
	}

	url :=
		fmt.Sprintf(
			"%s/fixtures?id=%d",
			baseURL,
			selectedMatch,
		)

	body, err :=
		apiRequest(url)

	if err != nil {
		return
	}

	var data FixtureResponse

	json.Unmarshal(
		body,
		&data,
	)

	if len(data.Response) == 0 {
		return
	}

	match :=
		data.Response[0]

	getStat := func(statType string) (string, string) {

		home := "-"
		away := "-"

		for i, teamStats := range match.Statistics {

			for _, stat :=
				range teamStats.Statistics {

				if stat.Type == statType {

					value :=
						fmt.Sprintf(
							"%v",
							stat.Value,
						)

					if i == 0 {
						home = value
					} else {
						away = value
					}
				}
			}
		}

		return home, away
	}

	posH, posA :=
		getStat("Ball Possession")

	shotH, shotA :=
		getStat("Shots on Goal")

	cornerH, cornerA :=
		getStat("Corner Kicks")

	offH, offA :=
		getStat("Offsides")

	foulH, foulA :=
		getStat("Fouls")

	yellowH, yellowA :=
		getStat("Yellow Cards")

	redH, redA :=
		getStat("Red Cards")

	totalGoals :=
		match.Goals.Home +
			match.Goals.Away

	message :=
		fmt.Sprintf(
			"🌍 %s\n"+
				"🏆 %s\n\n"+

				"⚽ %s %d - %d %s\n\n"+

				"⏱ %d'\n\n"+

				"📊 Possession\n"+
				"%s vs %s\n\n"+

				"🎯 Shots\n"+
				"%s vs %s\n\n"+

				"🚩 Corners\n"+
				"%s vs %s\n\n"+

				"🚷 Offsides\n"+
				"%s vs %s\n\n"+

				"🤕 Fouls\n"+
				"%s vs %s\n\n"+

				"🟨 Yellow\n"+
				"%s vs %s\n\n"+

				"🟥 Red\n"+
				"%s vs %s",

			match.League.Country,
			match.League.Name,

			match.Teams.Home.Name,
			match.Goals.Home,

			match.Goals.Away,
			match.Teams.Away.Name,

			match.Fixture.Status.Elapsed,

			posH, posA,
			shotH, shotA,
			cornerH, cornerA,
			offH, offA,
			foulH, foulA,
			yellowH, yellowA,
			redH, redA,
		)

	fmt.Println(
		"\n====================",
	)

	fmt.Println(message)

	if totalGoals > lastGoals {

		lastGoals = totalGoals

		sendTelegram(
			"⚽ GOAL!\n\n" + message,
		)

	} else {

		sendTelegram(message)
	}
}

func getTelegramUpdates() {

	token := os.Getenv("BOT_TOKEN")

	for {

		url :=
			fmt.Sprintf(
				"https://api.telegram.org/bot%s/getUpdates?offset=%d",
				token,
				lastUpdateID+1,
			)

		resp, err :=
			http.Get(url)

		if err != nil {

			time.Sleep(
				3 * time.Second,
			)

			continue
		}

		body, _ :=
			io.ReadAll(resp.Body)

		resp.Body.Close()

		var data TelegramResponse

		json.Unmarshal(
			body,
			&data,
		)

		for _, update := range data.Result {

			lastUpdateID =
				update.UpdateID

			text :=
				strings.TrimSpace(
					update.Message.Text,
				)

			fmt.Println(
				"telegram:",
				text,
			)

			switch {

			case text == "/startbot":

				sendUpcomingMatches()

			case text == "/random":

				watchRandomLiveMatch()

			case text == "/stopbot":

				isWatching = false
				selectedMatch = 0

				sendTelegram(
					"🛑 Bot stopped",
				)

			default:

				number, err :=
					strconv.Atoi(text)

				if err != nil {
					continue
				}

				if number < 1 ||
					number > len(currentMatches) {

					sendTelegram(
						"❌ Nomor tidak valid",
					)

					continue
				}

				selectedMatch =
					currentMatches[number-1].ID

				lastGoals = 0

				isWatching = true

				sendTelegram(
					"👀 Watching\n\n" +
						currentMatches[number-1].Title,
				)

				go watchLoop()
			}
		}

		time.Sleep(
			2 * time.Second,
		)
	}
}

func main() {

	rand.Seed(
		time.Now().UnixNano(),
	)

	godotenv.Load()

	sendTelegram(
		"⚽ Football Bot Started",
	)

	sendMenu()

	log.Println(
		"Bot running...",
	)

	getTelegramUpdates()
}
