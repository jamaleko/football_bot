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

const (
	baseURL = "https://v3.football.api-sports.io"
)

var (
	lastGoals     = 0
	selectedMatch = 0
	matchMap      = map[int]int{}
)

type FixtureResponse struct {
	Response []struct {
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

	payload := map[string]string{
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

func sendTodayMatches() {

	today :=
		time.Now().Format("2006-01-02")

	url :=
		fmt.Sprintf(
			"%s/fixtures?date=%s",
			baseURL,
			today,
		)

	body, err := apiRequest(url)

	if err != nil {
		log.Println(err)
		return
	}

	var data FixtureResponse

	json.Unmarshal(body, &data)

	if len(data.Response) == 0 {
		return
	}

	message :=
		"📅 Match Hari Ini\n\n"

	matchMap = map[int]int{}

	index := 1

	for _, match := range data.Response {

		matchTime, _ :=
			time.Parse(
				time.RFC3339,
				match.Fixture.Date,
			)

		wib :=
			matchTime.In(
				time.FixedZone(
					"WIB",
					7*3600,
				),
			)

		message += fmt.Sprintf(
			"%d. ⚽ %s vs %s\n🕒 %s WIB\n\n",

			index,

			match.Teams.Home.Name,
			match.Teams.Away.Name,

			wib.Format("15:04"),
		)

		matchMap[index] =
			match.Fixture.ID

		index++

		if index > 10 {
			break
		}
	}

	message +=
		"Reply angka untuk pilih match"

	sendTelegram(message)
}

func sendLiveMatches() bool {

	url :=
		fmt.Sprintf(
			"%s/fixtures?live=all",
			baseURL,
		)

	body, err := apiRequest(url)

	if err != nil {
		log.Println(err)
		return false
	}

	var data FixtureResponse

	json.Unmarshal(body, &data)

	if len(data.Response) == 0 {
		return false
	}

	message :=
		"🔥 LIVE NOW\n\n"

	matchMap = map[int]int{}

	index := 1

	for _, match := range data.Response {

		message += fmt.Sprintf(
			"%d. ⚽ %s %d - %d %s\n⏱ %d'\n\n",

			index,

			match.Teams.Home.Name,
			match.Goals.Home,

			match.Goals.Away,
			match.Teams.Away.Name,

			match.Fixture.Status.Elapsed,
		)

		matchMap[index] =
			match.Fixture.ID

		index++

		if index > 10 {
			break
		}
	}

	message +=
		"Reply angka untuk watch"

	sendTelegram(message)

	return true
}

func getTelegramUpdates() {

	token := os.Getenv("BOT_TOKEN")

	url :=
		fmt.Sprintf(
			"https://api.telegram.org/bot%s/getUpdates",
			token,
		)

	resp, err := http.Get(url)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}

	json.Unmarshal(body, &result)

	resultsRaw, ok :=
		result["result"]

	if !ok {
		return
	}

	results :=
		resultsRaw.([]interface{})

	if len(results) == 0 {
		return
	}

	last :=
		results[len(results)-1].(map[string]interface{})

	messageRaw, ok :=
		last["message"]

	if !ok {
		return
	}

	message :=
		messageRaw.(map[string]interface{})

	textRaw, ok :=
		message["text"]

	if !ok {
		return
	}

	text :=
		textRaw.(string)

	fmt.Println(
		"telegram text:",
		text,
	)

	number, err :=
		strconv.Atoi(
			strings.TrimSpace(text),
		)

	if err != nil {
		return
	}

	matchID :=
		matchMap[number]

	if matchID == 0 {
		return
	}

	selectedMatch = matchID

	lastGoals = 0

	sendTelegram(
		fmt.Sprintf(
			"✅ Watching match #%d",
			number,
		),
	)
}

func getLiveMatches() int {

	url :=
		fmt.Sprintf(
			"%s/fixtures?live=all",
			baseURL,
		)

	body, err := apiRequest(url)

	if err != nil {
		log.Println(err)
		return 0
	}

	var data FixtureResponse

	json.Unmarshal(body, &data)

	return len(data.Response)
}

func watchSelectedMatch() {

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

	totalGoals :=
		match.Goals.Home +
			match.Goals.Away

	homePossession := "-"
	awayPossession := "-"

	for _, teamStats :=
		range match.Statistics {

		for _, stat :=
			range teamStats.Statistics {

			if stat.Type ==
				"Ball Possession" {

				if teamStats.Team.Name ==
					match.Teams.Home.Name {

					homePossession =
						fmt.Sprintf(
							"%v",
							stat.Value,
						)

				 } else {

					awayPossession =
						fmt.Sprintf(
							"%v",
							stat.Value,
						)
				}
			}
		}
	}

	if totalGoals > lastGoals {

		lastGoals = totalGoals

		sendTelegram(
			fmt.Sprintf(
				"⚽ GOAL!\n\n"+
					"%s %d - %d %s\n\n"+
					"⏱ %d'\n\n"+
					"📊 Possession\n"+
					"%s vs %s",

				match.Teams.Home.Name,
				match.Goals.Home,

				match.Goals.Away,
				match.Teams.Away.Name,

				match.Fixture.Status.Elapsed,

				homePossession,
				awayPossession,
			),
		)

	} else {

		sendTelegram(
			fmt.Sprintf(
				"📊 LIVE MATCH\n\n"+
					"%s %d - %d %s\n\n"+
					"⏱ %d'\n\n"+
					"📊 Possession\n"+
					"%s vs %s",

				match.Teams.Home.Name,
				match.Goals.Home,

				match.Goals.Away,
				match.Teams.Away.Name,

				match.Fixture.Status.Elapsed,

				homePossession,
				awayPossession,
			),
		)
	}
}

func main() {

	godotenv.Load()

	sendTelegram(
		"⚽ Football Bot Started",
	)

	// cek live dulu
	hasLive :=
		sendLiveMatches()

	// kalau tidak ada live
	if !hasLive {

		sendTodayMatches()
	}

	// telegram polling cepat
	go func() {

		for {

			getTelegramUpdates()

			time.Sleep(
				10 * time.Second,
			)
		}

	}()

	// football loop
	for {

		liveCount :=
			getLiveMatches()

		if liveCount == 0 {

			log.Println(
				"no live match",
			)

			time.Sleep(
				30 * time.Minute,
			)

		} else {

			watchSelectedMatch()

			time.Sleep(
				5 * time.Minute,
			)
		}
	}
}
