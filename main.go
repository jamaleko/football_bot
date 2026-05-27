package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const baseURL = "https://v3.football.api-sports.io"

var (
	isWatching     = false
	selectedMatch  = 0
	lastGoals      = 0
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

func sendTelegram(message string) {

	token := os.Getenv("BOT_TOKEN")
	chatID := os.Getenv("CHAT_ID")

	url := fmt.Sprintf(
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

func sendMenu() {

	message :=
		"⚽ FOOTBALL BOT\n\n" +
			"/startbot\n" +
			"/stopbot"

	sendTelegram(message)
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

func sendUpcomingMatches() {

	now := time.Now()

	matchIndex := 1

	message :=
		"📅 BIG MATCHES\n\n"

	matchMap := map[int]int{}

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

			message += fmt.Sprintf(
				"%d. 🌍 %s\n"+
					"🏆 %s\n"+
					"⚽ %s vs %s\n"+
					"🕒 %s WIB\n\n",

				matchIndex,

				match.League.Country,
				match.League.Name,

				match.Teams.Home.Name,
				match.Teams.Away.Name,

				wib.Format(
					"02 Jan 15:04",
				),
			)

			matchMap[matchIndex] =
				match.Fixture.ID

			matchIndex++

			if matchIndex > 10 {
				break
			}
		}
	}

	sendTelegram(message)

	selectMatch(matchMap)
}

func selectMatch(matchMap map[int]int) {

	fmt.Println("\nPilih nomor match:")

	var input string

	fmt.Scanln(&input)

	number, err :=
		strconv.Atoi(input)

	if err != nil {
		fmt.Println("input salah")
		return
	}

	matchID :=
		matchMap[number]

	if matchID == 0 {
		fmt.Println("match tidak ditemukan")
		return
	}

	selectedMatch =
		matchID

	lastGoals = 0

	isWatching = true

	sendTelegram(
		fmt.Sprintf(
			"👀 Watching Match #%d",
			number,
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

	if totalGoals > lastGoals {

		lastGoals = totalGoals

		sendTelegram(
			"⚽ GOAL!\n\n" + message,
		)

	} else {

		sendTelegram(message)
	}
}

func main() {

	godotenv.Load()

	sendTelegram(
		"⚽ Football Bot Started",
	)

	sendMenu()

	for {

		fmt.Println(
			"\n/startbot atau /stopbot",
		)

		var input string

		fmt.Scanln(&input)

		switch input {

		case "/startbot":

			if isWatching {

				fmt.Println(
					"bot already running",
				)

				continue
			}

			sendUpcomingMatches()

		case "/stopbot":

			isWatching = false

			selectedMatch = 0

			sendTelegram(
				"🛑 Bot Stopped",
			)

			fmt.Println(
				"bot stopped",
			)
		}
	}
}
