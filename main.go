package main

import (
	"bufio"
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
	selectedMatch int
	lastGoals     int
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

func getLiveMatches() []struct {
	ID    int
	Title string
} {

	url :=
		fmt.Sprintf(
			"%s/fixtures?live=all",
			baseURL,
		)

	body, err := apiRequest(url)

	if err != nil {
		log.Println(err)
		return nil
	}

	var data FixtureResponse

	json.Unmarshal(body, &data)

	var matches []struct {
		ID    int
		Title string
	}

	fmt.Println("\n🔥 LIVE MATCHES\n")

	for i, match := range data.Response {

		title :=
			fmt.Sprintf(
				"%s %d - %d %s (%d')",

				match.Teams.Home.Name,
				match.Goals.Home,

				match.Goals.Away,
				match.Teams.Away.Name,

				match.Fixture.Status.Elapsed,
			)

		fmt.Printf(
			"%d. %s\n",
			i+1,
			title,
		)

		matches =
			append(
				matches,
				struct {
					ID    int
					Title string
				}{
					ID:    match.Fixture.ID,
					Title: title,
				},
			)
	}

	return matches
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
		log.Println(err)
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

	message :=
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
	}
}

func main() {

	godotenv.Load()

	sendTelegram(
		"⚽ Football Bot Started",
	)

	matches :=
		getLiveMatches()

	if len(matches) == 0 {

		fmt.Println(
			"No live matches",
		)

		return
	}

	fmt.Print(
		"\nPilih nomor match: ",
	)

	reader :=
		bufio.NewReader(os.Stdin)

	input, _ :=
		reader.ReadString('\n')

	input =
		strings.TrimSpace(input)

	number, err :=
		strconv.Atoi(input)

	if err != nil {

		fmt.Println(
			"Input salah",
		)

		return
	}

	if number < 1 ||
		number > len(matches) {

		fmt.Println(
			"Nomor tidak valid",
		)

		return
	}

	selectedMatch =
		matches[number-1].ID

	fmt.Println(
		"\n✅ Watching:",
		matches[number-1].Title,
	)

	sendTelegram(
		"👀 Watching:\n\n" +
			matches[number-1].Title,
	)

	for {

		watchMatch()

		time.Sleep(
			5 * time.Minute,
		)
	}
}
