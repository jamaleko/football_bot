package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const BOT_TOKEN = "8727181698:AAGAehZUWyQmh7CBiWNt_x6Xsevh2hrcuCE"

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

type Event struct {
	ID int

	League string
	Country string

	HomeTeam string
	AwayTeam string

	HomeScore int
	AwayScore int

	Status string
}

type UserSession struct {
	Watching bool
	MatchIndex int
}

type SportsDBResponse struct {
	Events []struct {

		IDEvent string `json:"idEvent"`

		StrLeague string `json:"strLeague"`

		StrCountry string `json:"strCountry"`

		StrHomeTeam string `json:"strHomeTeam"`

		StrAwayTeam string `json:"strAwayTeam"`

		IntHomeScore string `json:"intHomeScore"`

		IntAwayScore string `json:"intAwayScore"`

		StrStatus string `json:"strStatus"`
	} `json:"events"`
}

var (
	lastUpdateID = 0

	cachedMatches []Event

	userSessions =
		map[int64]*UserSession{}
)

func main() {

	rand.Seed(
		time.Now().UnixNano(),
	)

	fmt.Println("⚽ BOT RUNNING...")

	for {

		getUpdates()

		time.Sleep(
			2 * time.Second,
		)
	}
}

func getUpdates() {

	endpoint :=
		fmt.Sprintf(
			"%s/getUpdates?offset=%d",
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

	json.Unmarshal(
		body,
		&result,
	)

	for _, update := range result.Result {

		lastUpdateID =
			update.UpdateID

		text :=
			strings.TrimSpace(
				strings.ToUpper(
					update.Message.Text,
				),
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

	switch {

	case text == "/START":

		sendMainMenu(chatID)

	case text == "/LIVE" ||
		text == "LIVE":

		sendLiveMatches(chatID)

	case text == "/BIG" ||
		text == "BIG":

		sendBigMatches(chatID)

	case text == "/RANDOM" ||
		text == "RANDOM":

		watchRandom(chatID)

	case text == "/REFRESH" ||
		text == "REFRESH":

		refreshMatch(chatID)

	case text == "/STOP" ||
		text == "STOP":

		stopWatch(chatID)

	case strings.HasPrefix(
		text,
		"/WATCH ",
	):

		parts :=
			strings.Split(
				text,
				" ",
			)

		if len(parts) < 2 {
			return
		}

		number, err :=
			strconv.Atoi(
				parts[1],
			)

		if err != nil {
			return
		}

		watchByNumber(
			chatID,
			number,
		)
	}
}

func mainKeyboard() map[string]interface{} {

	return map[string]interface{}{
		"keyboard": [][]map[string]string{

			{
				{
					"text": "LIVE",
				},
				{
					"text": "BIG",
				},
			},

			{
				{
					"text": "RANDOM",
				},
				{
					"text": "REFRESH",
				},
			},

			{
				{
					"text": "STOP",
				},
			},
		},

		"resize_keyboard": true,
	}
}

func sendMainMenu(
	chatID int64,
) {

	msg :=
		`⚽ FOOTBALL BOT

LIVE
BIG
RANDOM
REFRESH
STOP`

	sendTelegram(
		chatID,
		msg,
		mainKeyboard(),
	)
}

func fetchLiveMatches() []Event {

	now :=
		time.Now().Format(
			"2006-01-02",
		)

	link :=
		fmt.Sprintf(
			"https://www.thesportsdb.com/api/v1/json/123/eventsday.php?d=%s&s=Soccer",
			now,
		)

	resp, err :=
		http.Get(link)

	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	body, _ :=
		io.ReadAll(resp.Body)

	var result SportsDBResponse

	json.Unmarshal(
		body,
		&result,
	)

	var matches []Event

	for _, e := range result.Events {

		if e.StrStatus != "Live" &&
			e.StrStatus != "1H" &&
			e.StrStatus != "2H" {

			continue
		}

		status :=
			e.StrStatus

		if status == "1H" ||
			status == "2H" ||
			status == "Live" {

			status = "LIVE"
		}

		homeScore := 0
		awayScore := 0

		fmt.Sscanf(
			e.IntHomeScore,
			"%d",
			&homeScore,
		)

		fmt.Sscanf(
			e.IntAwayScore,
			"%d",
			&awayScore,
		)

		id := 0

		fmt.Sscanf(
			e.IDEvent,
			"%d",
			&id,
		)

		matches =
			append(
				matches,
				Event{

					ID: id,

					League:
						e.StrLeague,

					Country:
						e.StrCountry,

					HomeTeam:
						e.StrHomeTeam,

					AwayTeam:
						e.StrAwayTeam,

					HomeScore:
						homeScore,

					AwayScore:
						awayScore,

					Status:
						status,
				},
			)
	}

	return matches
}

func sendLiveMatches(
	chatID int64,
) {

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

				m.Country,
				m.League,

				m.HomeTeam,
				m.HomeScore,

				m.AwayScore,
				m.AwayTeam,

				m.Status,
			),
		)
	}

	msg.WriteString(
		"Watch:\n/WATCH 1",
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

	userSessions[chatID] =
		&UserSession{
			Watching: true,
			MatchIndex: number - 1,
		}

	sendMatchDetail(
		chatID,
		number-1,
	)
}

func watchRandom(
	chatID int64,
) {

	if len(cachedMatches) == 0 {

		cachedMatches =
			fetchLiveMatches()
	}

	if len(cachedMatches) == 0 {

		sendTelegram(
			chatID,
			"❌ Tidak ada live match",
			nil,
		)

		return
	}

	index :=
		rand.Intn(
			len(cachedMatches),
		)

	userSessions[chatID] =
		&UserSession{
			Watching: true,
			MatchIndex: index,
		}

	sendMatchDetail(
		chatID,
		index,
	)
}

func refreshMatch(
	chatID int64,
) {

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
		session.MatchIndex,
	)
}

func stopWatch(
	chatID int64,
) {

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
	index int,
) {

	if index < 0 ||
		index >= len(cachedMatches) {

		sendTelegram(
			chatID,
			"❌ Match tidak ditemukan",
			nil,
		)

		return
	}

	match :=
		cachedMatches[index]

	msg :=
		fmt.Sprintf(
			`🌍 %s
🏆 %s

⚽ %s %d - %d %s

⏱ %s

⚽ Goal:
-

🟨 Yellow:
-

🟥 Red:
-`,

			match.Country,
			match.League,

			match.HomeTeam,
			match.HomeScore,

			match.AwayScore,
			match.AwayTeam,

			match.Status,
		)

	sendTelegram(
		chatID,
		msg,
		nil,
	)
}

func sendBigMatches(
	chatID int64,
) {

	sendTelegram(
		chatID,
		"🔥 BIG MATCH COMING SOON",
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
			json.Marshal(
				keyboard,
			)

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
