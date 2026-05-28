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

const baseURL = "https://v3.football.api-sports.io"

type UserSession struct {
	SelectedMatch int

	LastGoals  int
	LastYellow int
	LastRed    int

	LastHalfTime bool
	LastFullTime bool

	IsWatching bool
}

type MatchItem struct {
	ID    int
	Title string
}

var (
	lastUpdateID int

	userSessions = map[int64]*UserSession{}

	userMatches = map[int64][]MatchItem{}
)

var bigLeagueIDs = []int{

	39,  // Premier League
	140, // La Liga
	135, // Serie A
	78,  // Bundesliga
	61,  // Ligue 1

	274, // Liga 1 Indonesia

	2,   // UCL
	3,   // Europa
	848, // Conference

	1, // World Cup
	4, // Euro
	5, // Nations League
	9, // Copa America

	32, // WC Qual Europe
	33, // WC Qual South America
	34, // WC Qual CONCACAF
	35, // WC Qual Asia
	36, // WC Qual Africa

	10, // Friendly
	960, //Euro Kualifikasi
}

type FixtureResponse struct {
	Response []struct {

		League struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			Country string `json:"country"`
		} `json:"league"`

		Fixture struct {
			ID int `json:"id"`

			Status struct {
				Elapsed int    `json:"elapsed"`
				Short   string `json:"short"`
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

			Statistics []struct {
				Type  string      `json:"type"`
				Value interface{} `json:"value"`
			} `json:"statistics"`

		} `json:"statistics"`

	} `json:"response"`
}

type TelegramResponse struct {
	Result []struct {

		UpdateID int `json:"update_id"`

		Message struct {

			Text string `json:"text"`

			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`

		} `json:"message"`

		CallbackQuery struct {

			ID string `json:"id"`

			Data string `json:"data"`

			Message struct {

				Chat struct {
					ID int64 `json:"id"`
				} `json:"chat"`

			} `json:"message"`

		} `json:"callback_query"`

	} `json:"result"`
}

func isBigLeague(id int) bool {

	for _, leagueID :=
		range bigLeagueIDs {

		if leagueID == id {
			return true
		}
	}

	return false
}

func apiRequest(
	url string,
) ([]byte, error) {

	req, _ :=
		http.NewRequest(
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

func sendTelegram(
	chatID int64,
	message string,
	keyboard interface{},
) {

	token := os.Getenv("BOT_TOKEN")

	url :=
		fmt.Sprintf(
			"https://api.telegram.org/bot%s/sendMessage",
			token,
		)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
	}

	if keyboard != nil {

		payload["reply_markup"] =
			keyboard
	}

	body, _ :=
		json.Marshal(payload)

	http.Post(
		url,
		"application/json",
		bytes.NewBuffer(body),
	)
}

func sendMainMenu(
	chatID int64,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{

			{
				{
					"text": "📺 Live Now",
					"callback_data": "live_now",
				},
				{
					"text": "📅 Big Matches",
					"callback_data": "big_matches",
				},
			},

			{
				{
					"text": "🎲 Random Match",
					"callback_data": "random_match",
				},
				{
					"text": "📊 Refresh",
					"callback_data": "refresh",
				},
			},

			{
				{
					"text": "🛑 Stop",
					"callback_data": "stop",
				},
			},
		},
	}

	sendTelegram(
		chatID,
		"⚽ FOOTBALL BOT",
		keyboard,
	)
}

func sendLiveMatches(
	chatID int64,
) {

	userMatches[chatID] = nil

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
			chatID,
			"❌ Tidak ada live match",
			nil,
		)

		return
	}

	message :=
		"📺 LIVE NOW\n\n"

	var keyboardRows [][]map[string]string

	index := 1

	for _, match :=
		range data.Response {

		if !isBigLeague(
			match.League.ID,
		) {
			continue
		}

		title :=
			fmt.Sprintf(
				"%d. %s vs %s (%d')\n",

				index,

				match.Teams.Home.Name,
				match.Teams.Away.Name,

				match.Fixture.Status.Elapsed,
			)

		message += title

		userMatches[chatID] =
			append(
				userMatches[chatID],
				MatchItem{
					ID: match.Fixture.ID,
				},
			)

		keyboardRows =
			append(
				keyboardRows,

				[]map[string]string{
					{
						"text": fmt.Sprintf(
							"⚽ Watch %d",
							index,
						),

						"callback_data": fmt.Sprintf(
							"watch_%d",
							index,
						),
					},
				},
			)

		index++

		if index > 10 {
			break
		}
	}

	keyboard := map[string]interface{}{
		"inline_keyboard": keyboardRows,
	}

	sendTelegram(
		chatID,
		message,
		keyboard,
	)
}

func sendUpcomingMatches(
	chatID int64,
) {

	userMatches[chatID] = nil

	now := time.Now()

	message :=
		"📅 BIG MATCHES\n\n"

	var keyboardRows [][]map[string]string

	index := 1

	for day := 0; day < 30; day++ {

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

		for _, match :=
			range data.Response {

			if !isBigLeague(
				match.League.ID,
			) {
				continue
			}

			if match.Fixture.Status.Short == "FT" {
				continue
			}

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

			title :=
				fmt.Sprintf(
					"%d. %s vs %s\n"+
						"🏆 %s\n"+
						"🕒 %s WIB\n\n",

					index,

					match.Teams.Home.Name,
					match.Teams.Away.Name,

					match.League.Name,

					wib.Format(
						"02 Jan 15:04",
					),
				)

			message += title

			userMatches[chatID] =
				append(
					userMatches[chatID],
					MatchItem{
						ID: match.Fixture.ID,
					},
				)

			keyboardRows =
				append(
					keyboardRows,

					[]map[string]string{
						{
							"text": fmt.Sprintf(
								"⚽ Watch %d",
								index,
							),

							"callback_data": fmt.Sprintf(
								"watch_%d",
								index,
							),
						},
					},
				)

			index++

			if index > 10 {

				keyboard := map[string]interface{}{
					"inline_keyboard": keyboardRows,
				}

				sendTelegram(
					chatID,
					message,
					keyboard,
				)

				return
			}
		}
	}

	keyboard := map[string]interface{}{
		"inline_keyboard": keyboardRows,
	}

	sendTelegram(
		chatID,
		message,
		keyboard,
	)
}

func resetSession(
	session *UserSession,
) {

	session.LastGoals = 0
	session.LastYellow = 0
	session.LastRed = 0

	session.LastHalfTime = false
	session.LastFullTime = false

	session.IsWatching = true
}

func watchMatch(
	chatID int64,
	matchID int,
) {

	session :=
		userSessions[chatID]

	if session == nil {

		session =
			&UserSession{}

		userSessions[chatID] =
			session
	}

	session.SelectedMatch =
		matchID

	resetSession(session)

	sendTelegram(
		chatID,
		"👀 Watching Match...",
		nil,
	)

	go watchLoop(chatID)
}

func watchRandomMatch(
	chatID int64,
) {

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
			chatID,
			"❌ Tidak ada live match",
			nil,
		)

		return
	}

	randomIndex :=
		rand.Intn(
			len(data.Response),
		)

	match :=
		data.Response[randomIndex]

	watchMatch(
		chatID,
		match.Fixture.ID,
	)
}

func getStatsText(
	data FixtureResponse,
) string {

	match :=
		data.Response[0]

	getStat := func(
		statType string,
	) (string, string) {

		home := "-"
		away := "-"

		for i, teamStats :=
			range match.Statistics {

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

	return fmt.Sprintf(

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
}

func refreshStats(
	chatID int64,
) {

	session :=
		userSessions[chatID]

	if session == nil {
		return
	}

	url :=
		fmt.Sprintf(
			"%s/fixtures?id=%d",
			baseURL,
			session.SelectedMatch,
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

	sendTelegram(
		chatID,
		getStatsText(data),
		nil,
	)
}

func watchLoop(
	chatID int64,
) {

	for {

		session :=
			userSessions[chatID]

		if session == nil {
			return
		}

		if !session.IsWatching {
			return
		}

		checkEvents(chatID)

		time.Sleep(
			1 * time.Minute,
		)
	}
}

func checkEvents(
	chatID int64,
) {

	session :=
		userSessions[chatID]

	if session == nil {
		return
	}

	url :=
		fmt.Sprintf(
			"%s/fixtures?id=%d",
			baseURL,
			session.SelectedMatch,
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

	getStatInt := func(
		statType string,
	) int {

		total := 0

		for _, teamStats :=
			range match.Statistics {

			for _, stat :=
				range teamStats.Statistics {

				if stat.Type == statType {

					value :=
						fmt.Sprintf(
							"%v",
							stat.Value,
						)

					value =
						strings.ReplaceAll(
							value,
							"%",
							"",
						)

					n, _ :=
						strconv.Atoi(
							value,
						)

					total += n
				}
			}
		}

		return total
	}

	totalGoals :=
		match.Goals.Home +
			match.Goals.Away

	totalYellow :=
		getStatInt(
			"Yellow Cards",
		)

	totalRed :=
		getStatInt(
			"Red Cards",
		)

	status :=
		match.Fixture.Status.Short

	if totalGoals >
		session.LastGoals {

		session.LastGoals =
			totalGoals

		sendTelegram(
			chatID,

			"⚽ GOAL!\n\n"+
				getStatsText(data),

			nil,
		)
	}

	if totalYellow >
		session.LastYellow {

		session.LastYellow =
			totalYellow

		sendTelegram(
			chatID,

			"🟨 YELLOW CARD!\n\n"+
				getStatsText(data),

			nil,
		)
	}

	if totalRed >
		session.LastRed {

		session.LastRed =
			totalRed

		sendTelegram(
			chatID,

			"🟥 RED CARD!\n\n"+
				getStatsText(data),

			nil,
		)
	}

	if status == "HT" &&
		!session.LastHalfTime {

		session.LastHalfTime =
			true

		sendTelegram(
			chatID,

			"⏸ HALF TIME\n\n"+
				getStatsText(data),

			nil,
		)
	}

	if status == "FT" &&
		!session.LastFullTime {

		session.LastFullTime =
			true

		session.IsWatching =
			false

		sendTelegram(
			chatID,

			"✅ FULL TIME\n\n"+
				getStatsText(data),

			nil,
		)
	}
}

func answerCallbackQuery(
	callbackID string,
) {

	token := os.Getenv("BOT_TOKEN")

	url :=
		fmt.Sprintf(
			"https://api.telegram.org/bot%s/answerCallbackQuery",
			token,
		)

	payload := map[string]interface{}{
		"callback_query_id": callbackID,
	}

	body, _ :=
		json.Marshal(payload)

	http.Post(
		url,
		"application/json",
		bytes.NewBuffer(body),
	)
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

		for _, update :=
			range data.Result {

			lastUpdateID =
				update.UpdateID

			// COMMAND
			if update.Message.Text != "" {

				chatID :=
					update.Message.Chat.ID

				text :=
					strings.TrimSpace(
						update.Message.Text,
					)

				switch text {

				case "/start":

					sendMainMenu(chatID)
				}
			}

			// BUTTON CLICK
			if update.CallbackQuery.Data != "" {

				chatID :=
					update.CallbackQuery.Message.Chat.ID

				data :=
					update.CallbackQuery.Data

				answerCallbackQuery(
					update.CallbackQuery.ID,
				)

				switch {

				case data == "live_now":

					sendLiveMatches(chatID)

				case data == "big_matches":

					sendUpcomingMatches(chatID)

				case data == "random_match":

					watchRandomMatch(chatID)

				case data == "refresh":

					refreshStats(chatID)

				case data == "stop":

					session :=
						userSessions[chatID]

					if session != nil {

						session.IsWatching =
							false
					}

					sendTelegram(
						chatID,
						"🛑 Bot stopped",
						nil,
					)

				case strings.HasPrefix(
					data,
					"watch_",
				):

					indexText :=
						strings.TrimPrefix(
							data,
							"watch_",
						)

					index, err :=
						strconv.Atoi(
							indexText,
						)

					if err != nil {
						continue
					}

					matches :=
						userMatches[chatID]

					if index < 1 ||
						index > len(matches) {

						continue
					}

					match :=
						matches[index-1]

					watchMatch(
						chatID,
						match.ID,
					)
				}
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

	log.Println(
		"Football bot running...",
	)

	getTelegramUpdates()
}
