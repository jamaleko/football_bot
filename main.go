package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	BOT_TOKEN = "8727181698:AAGAehZUWyQmh7CBiWNt_x6Xsevh2hrcuCE"
)

var TELEGRAM_API =
	"https://api.telegram.org/bot" + BOT_TOKEN

type TelegramResponse struct {
	Ok     bool `json:"ok"`
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

type SofaLiveResponse struct {
	Events []Event `json:"events"`
}

type Event struct {
	ID     int `json:"id"`
	Status struct {
		Description string `json:"description"`
		Type        string `json:"type"`
	} `json:"status"`

	Tournament struct {
		Name string `json:"name"`

		Category struct {
			Name string `json:"name"`
		} `json:"category"`
	} `json:"tournament"`

	HomeTeam struct {
		Name string `json:"name"`
	} `json:"homeTeam"`

	AwayTeam struct {
		Name string `json:"name"`
	} `json:"awayTeam"`

	HomeScore struct {
		Current int `json:"current"`
	} `json:"homeScore"`

	AwayScore struct {
		Current int `json:"current"`
	} `json:"awayScore"`
}

type IncidentResponse struct {
	Incidents []Incident `json:"incidents"`
}

type Incident struct {
	Time int    `json:"time"`
	Type string `json:"incidentType"`

	Player struct {
		Name string `json:"name"`
	} `json:"player"`

	ScoringTeam string `json:"scoringTeam"`
}

type UserSession struct {
	Watching bool
	MatchID  int
}

var userSessions =
	map[int64]*UserSession{}

var lastUpdateID = 0

func main() {

	rand.Seed(time.Now().UnixNano())

	fmt.Println("⚽ SofaScore Bot Started")

	for {

		getUpdates()

		time.Sleep(2 * time.Second)
	}
}

func getUpdates() {

	endpoint :=
		fmt.Sprintf(
			"%s/getUpdates?timeout=30&offset=%d",
			TELEGRAM_API,
			lastUpdateID+1,
		)

	resp, err :=
		http.Get(endpoint)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, _ :=
		io.ReadAll(resp.Body)

	var result TelegramResponse

	json.Unmarshal(body, &result)

	for _, update := range result.Result {

		lastUpdateID = update.UpdateID

		text :=
			strings.TrimSpace(
				update.Message.Text,
			)

		chatID :=
			update.Message.Chat.ID

		handleCommand(
			chatID,
			text,
		)
	}
}

func handleCommand(
	chatID int64,
	text string,
) {

	text =
		strings.ToUpper(text)

	switch {

	case text == "/START":

		sendMainMenu(chatID)

	case text == "/LIVE" || text == "LIVE":

		sendLiveMatches(chatID)

	case text == "/BIG" || text == "BIG":

		sendBigMatches(chatID)

	case text == "/RANDOM" || text == "RANDOM":

		watchRandom(chatID)

	case text == "/REFRESH" || text == "REFRESH":

		refreshMatch(chatID)

	case text == "/STOP" || text == "STOP":

		stopWatch(chatID)

	case strings.HasPrefix(
		text,
		"/WATCH",
	):

		parts :=
			strings.Split(text, " ")

		if len(parts) < 2 {
			return
		}

		number, err :=
			strconv.Atoi(parts[1])

		if err != nil {
			return
		}

		watchByNumber(
			chatID,
			number,
		)
	}
}

func sendMainMenu(chatID int64) {

	msg :=
		`⚽ FOOTBALL BOT

/LIVE
/BIG
/RANDOM
/REFRESH
/STOP`

	sendTelegram(
		chatID,
		msg,
		mainKeyboard(),
	)
}

func mainKeyboard() map[string]interface{} {

	return map[string]interface{}{
		"keyboard": [][]map[string]string{

			{
				{
					"text": "/LIVE",
				},
				{
					"text": "/BIG",
				},
			},

			{
				{
					"text": "/RANDOM",
				},
				{
					"text": "/REFRESH",
				},
			},

			{
				{
					"text": "/STOP",
				},
			},
		},

		"resize_keyboard": true,
	}
}

var cachedMatches []Event

func fetchLiveMatches() []Event {

	url :=
		"https://www.sofascore.com/api/v1/sport/football/events/live"

	req, _ :=
		http.NewRequest(
			"GET",
			url,
			nil,
		)

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0",
	)

	client := &http.Client{}

	resp, err :=
		client.Do(req)

	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	body, _ :=
		io.ReadAll(resp.Body)

	var result SofaLiveResponse

	json.Unmarshal(body, &result)

	return result.Events
}

func sendLiveMatches(chatID int64) {

	matches :=
		fetchLiveMatches()

	cachedMatches = matches

	if len(matches) == 0 {

		sendTelegram(
			chatID,
			"❌ Tidak ada live match",
			nil,
		)

		return
	}

	var msg strings.Builder

	msg.WriteString(
		"🔥 LIVE MATCHES\n\n",
	)

	limit := 10

	if len(matches) < limit {
		limit = len(matches)
	}

	for i := 0; i < limit; i++ {

		m := matches[i]

		msg.WriteString(
			fmt.Sprintf(
				"%d. 🌍 %s\n🏆 %s\n⚽ %s %d - %d %s\n⏱ %s\n\n",
				i+1,
				m.Tournament.Category.Name,
				m.Tournament.Name,
				m.HomeTeam.Name,
				m.HomeScore.Current,
				m.AwayScore.Current,
				m.AwayTeam.Name,
				m.Status.Description,
			),
		)
	}

	msg.WriteString(
		"Reply:\n/WATCH 1",
	)

	sendTelegram(
		chatID,
		msg.String(),
		nil,
	)
}

func watchByNumber(
	chatID int64,
	number int,
) {

	if number <= 0 ||
		number > len(cachedMatches) {

		sendTelegram(
			chatID,
			"❌ Match tidak ditemukan",
			nil,
		)

		return
	}

	match :=
		cachedMatches[number-1]

	userSessions[chatID] =
		&UserSession{
			Watching: true,
			MatchID:  match.ID,
		}

	sendMatchDetail(
		chatID,
		match.ID,
	)
}

func watchRandom(chatID int64) {

	matches :=
		fetchLiveMatches()

	if len(matches) == 0 {

		sendTelegram(
			chatID,
			"❌ Tidak ada live match",
			nil,
		)

		return
	}

	random :=
		matches[rand.Intn(len(matches))]

	userSessions[chatID] =
		&UserSession{
			Watching: true,
			MatchID:  random.ID,
		}

	sendMatchDetail(
		chatID,
		random.ID,
	)
}

func refreshMatch(chatID int64) {

	session :=
		userSessions[chatID]

	if session == nil {

		sendTelegram(
			chatID,
			"❌ Tidak sedang watch",
			nil,
		)

		return
	}

	sendMatchDetail(
		chatID,
		session.MatchID,
	)
}

func stopWatch(chatID int64) {

	delete(
		userSessions,
		chatID,
	)

	sendTelegram(
		chatID,
		"🛑 Watch stopped",
		nil,
	)
}

func sendMatchDetail(
	chatID int64,
	matchID int,
) {

	match :=
		findMatch(matchID)

	if match == nil {

		sendTelegram(
			chatID,
			"❌ Match selesai",
			nil,
		)

		return
	}

	incidents :=
		fetchIncidents(matchID)

	goals := []string{}
	yellows := []string{}
	reds := []string{}

	for _, i := range incidents {

		switch i.Type {

		case "goal":

			goals =
				append(
					goals,
					fmt.Sprintf(
						"%s %d'",
						i.Player.Name,
						i.Time,
					),
				)

		case "yellowCard":

			yellows =
				append(
					yellows,
					i.Player.Name,
				)

		case "redCard":

			reds =
				append(
					reds,
					i.Player.Name,
				)
		}
	}

	if len(goals) == 0 {
		goals = append(goals, "-")
	}

	if len(yellows) == 0 {
		yellows = append(yellows, "-")
	}

	if len(reds) == 0 {
		reds = append(reds, "-")
	}

	msg :=
		fmt.Sprintf(
			`🌍 %s
🏆 %s

⚽ %s %d - %d %s

⏱ %s

⚽ Goal:
%s

🟨 Yellow:
%s

🟥 Red:
%s`,
			match.Tournament.Category.Name,
			match.Tournament.Name,

			match.HomeTeam.Name,
			match.HomeScore.Current,
			match.AwayScore.Current,
			match.AwayTeam.Name,

			match.Status.Description,

			strings.Join(goals, "\n"),
			strings.Join(yellows, "\n"),
			strings.Join(reds, "\n"),
		)

	sendTelegram(
		chatID,
		msg,
		nil,
	)
}

func fetchIncidents(
	matchID int,
) []Incident {

	url :=
		fmt.Sprintf(
			"https://www.sofascore.com/api/v1/event/%d/incidents",
			matchID,
		)

	req, _ :=
		http.NewRequest(
			"GET",
			url,
			nil,
		)

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0",
	)

	client := &http.Client{}

	resp, err :=
		client.Do(req)

	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	body, _ :=
		io.ReadAll(resp.Body)

	var result IncidentResponse

	json.Unmarshal(body, &result)

	return result.Incidents
}

func findMatch(
	matchID int,
) *Event {

	matches :=
		fetchLiveMatches()

	for _, m := range matches {

		if m.ID == matchID {
			return &m
		}
	}

	return nil
}

func sendBigMatches(chatID int64) {

	sendTelegram(
		chatID,
		"🔥 BIG MATCH coming soon",
		nil,
	)
}

func sendTelegram(
	chatID int64,
	text string,
	keyboard interface{},
) {

	data :=
		url.Values{}

	data.Set(
		"chat_id",
		fmt.Sprintf("%d", chatID),
	)

	data.Set(
		"text",
		text,
	)

	if keyboard != nil {

		kb, _ :=
			json.Marshal(keyboard)

		data.Set(
			"reply_markup",
			string(kb),
		)
	}

	http.Post(
		TELEGRAM_API+"/sendMessage",
		"application/x-www-form-urlencoded",
		bytes.NewBufferString(
			data.Encode(),
		),
	)
}
