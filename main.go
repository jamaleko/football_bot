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
	baseURL = "https://www.thesportsdb.com/api/v1/json/3"
)

type UserSession struct {
	SelectedEvent string
	IsWatching    bool
	LastHomeScore int
	LastAwayScore int
}

type MatchItem struct {
	ID    string
	Title string
}

var (
	lastUpdateID int

	userSessions = map[int64]*UserSession{}
	userMatches  = map[int64][]MatchItem{}
)

var bigLeagueIDs = []string{
	"4328", // EPL
	"4335", // Bundesliga
	"4331", // Serie A
	"4332", // La Liga
	"4334", // Ligue 1
	"4480", // UCL
}

type LiveResponse struct {
	Event []struct {

		IDEvent string `json:"idEvent"`

		StrLeague string `json:"strLeague"`

		StrHomeTeam string `json:"strHomeTeam"`
		StrAwayTeam string `json:"strAwayTeam"`

		IntHomeScore string `json:"intHomeScore"`
		IntAwayScore string `json:"intAwayScore"`

		StrProgress string `json:"strProgress"`

		StrStatus string `json:"strStatus"`

		StrCountry string `json:"strCountry"`

	} `json:"event"`
}

type EventResponse struct {
	Events []struct {

		IDEvent string `json:"idEvent"`

		StrLeague string `json:"strLeague"`

		StrHomeTeam string `json:"strHomeTeam"`
		StrAwayTeam string `json:"strAwayTeam"`

		IntHomeScore string `json:"intHomeScore"`
		IntAwayScore string `json:"intAwayScore"`

		StrProgress string `json:"strProgress"`

		StrCountry string `json:"strCountry"`

		StrTimestamp string `json:"strTimestamp"`

		StrStatus string `json:"strStatus"`

		StrHomeGoalDetails string `json:"strHomeGoalDetails"`
		StrAwayGoalDetails string `json:"strAwayGoalDetails"`

		StrHomeYellowCards string `json:"strHomeYellowCards"`
		StrAwayYellowCards string `json:"strAwayYellowCards"`

		StrHomeRedCards string `json:"strHomeRedCards"`
		StrAwayRedCards string `json:"strAwayRedCards"`

	} `json:"events"`
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

	} `json:"result"`
}

func apiRequest(
	url string,
) ([]byte, error) {

	resp, err :=
		http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func sendTelegram(
	chatID int64,
	text string,
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
		"text":    text,
	}

	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}

	body, _ :=
		json.Marshal(payload)

	http.Post(
		url,
		"application/json",
		bytes.NewBuffer(body),
	)
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
		"is_persistent":  true,
	}
}

func sendMainMenu(
	chatID int64,
) {

	sendTelegram(
		chatID,

		"⚽ FOOTBALL BOT\n\n"+
			"LIVE = live match\n"+
			"BIG = big upcoming match\n"+
			"RANDOM = random live match\n"+
			"REFRESH = refresh stats\n"+
			"STOP = stop watching",

		mainKeyboard(),
	)
}

func sendLiveMatches(
	chatID int64,
) {

	userMatches[chatID] = nil

	url :=
		baseURL +
			"/livescore.php?s=Soccer"

	body, err :=
		apiRequest(url)

	if err != nil {
		return
	}

	var data LiveResponse

	json.Unmarshal(body, &data)

	if len(data.Event) == 0 {

		sendTelegram(
			chatID,
			"❌ Tidak ada live match",
			nil,
		)

		return
	}

	text :=
		"🔥 LIVE MATCHES\n\n"

	index := 1

	for _, match :=
		range data.Event {

		text += fmt.Sprintf(

			"%d. 🌍 %s\n"+
				"🏆 %s\n"+
				"⚽ %s %s - %s %s\n"+
				"⏱ %s\n\n",

			index,

			match.StrCountry,
			match.StrLeague,

			match.StrHomeTeam,
			match.IntHomeScore,

			match.IntAwayScore,
			match.StrAwayTeam,

			match.StrProgress,
		)

		userMatches[chatID] =
			append(
				userMatches[chatID],
				MatchItem{
					ID: match.IDEvent,
				},
			)

		index++

		if index > 10 {
			break
		}
	}

	text +=
		"Watch match:\n"+
			"WATCH 1"

	sendTelegram(
		chatID,
		text,
		nil,
	)
}

func sendBigMatches(
	chatID int64,
) {

	text :=
		"📅 BIG LEAGUES\n\n"

	text +=
		"Premier League\n"+
			"Serie A\n"+
			"La Liga\n"+
			"Bundesliga\n"+
			"Ligue 1\n"+
			"Champions League\n\n"

	text +=
		"Gunakan:\n"+
			"LIVE"

	sendTelegram(
		chatID,
		text,
		nil,
	)
}

func watchRandomMatch(
	chatID int64,
) {

	matches :=
		userMatches[chatID]

	if len(matches) == 0 {

		sendTelegram(
			chatID,
			"❌ Jalankan LIVE dulu",
			nil,
		)

		return
	}

	random :=
		matches[rand.Intn(len(matches))]

	watchMatch(
		chatID,
		random.ID,
	)
}

func watchMatch(
	chatID int64,
	eventID string,
) {

	session :=
		&UserSession{
			SelectedEvent: eventID,
			IsWatching:    true,
		}

	userSessions[chatID] =
		session

	refreshStats(chatID)
}

func refreshStats(
	chatID int64,
) {

	session :=
		userSessions[chatID]

	if session == nil {

		sendTelegram(
			chatID,
			"❌ Tidak ada match",
			nil,
		)

		return
	}

	url :=
		fmt.Sprintf(
			"%s/lookupevent.php?id=%s",
			baseURL,
			session.SelectedEvent,
		)

	body, err :=
		apiRequest(url)

	if err != nil {
		return
	}

	var data EventResponse

	json.Unmarshal(body, &data)

	if len(data.Events) == 0 {

		sendTelegram(
			chatID,
			"❌ Event tidak ditemukan",
			nil,
		)

		return
	}

	match := data.Events[0]

	goals :=
		formatGoals(
			match.StrHomeGoalDetails,
			match.StrAwayGoalDetails,
		)

	yellows :=
		formatCards(
			match.StrHomeYellowCards,
			match.StrAwayYellowCards,
		)

	reds :=
		formatCards(
			match.StrHomeRedCards,
			match.StrAwayRedCards,
		)

	text := fmt.Sprintf(

		"🌍 %s\n"+
			"🏆 %s\n\n"+

			"⚽ %s %s - %s %s\n\n"+

			"⏱ %s\n\n"+

			"⚽ Goal:\n%s\n\n"+

			"🟨 Yellow:\n%s\n\n"+

			"🟥 Red:\n%s",

		match.StrCountry,
		match.StrLeague,

		match.StrHomeTeam,
		match.IntHomeScore,

		match.IntAwayScore,
		match.StrAwayTeam,

		match.StrProgress,

		goals,
		yellows,
		reds,
	)

	sendTelegram(
		chatID,
		text,
		nil,
	)
}

func formatGoals(
	home string,
	away string,
) string {

	result := ""

	if home != "" {
		result += home + "\n"
	}

	if away != "" {
		result += away
	}

	if result == "" {
		result = "-"
	}

	return result
}

func formatCards(
	home string,
	away string,
) string {

	result := ""

	if home != "" {
		result += home + "\n"
	}

	if away != "" {
		result += away
	}

	if result == "" {
		result = "-"
	}

	return result
}

func getTelegramUpdates() {

	token := os.Getenv("BOT_TOKEN")

	for {

		url := fmt.Sprintf(
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

			if update.Message.Text != "" {

				chatID :=
					update.Message.Chat.ID

				text :=
					strings.TrimSpace(
						strings.ToUpper(
							update.Message.Text,
						),
					)

				switch {

				case text == "/START":

					sendMainMenu(chatID)

				case text == "LIVE":

					sendLiveMatches(chatID)

				case text == "BIG":

					sendBigMatches(chatID)

				case text == "RANDOM":

					watchRandomMatch(chatID)

				case text == "REFRESH":

					refreshStats(chatID)

				case text == "STOP":

					session :=
						userSessions[chatID]

					if session != nil {

						session.IsWatching =
							false
					}

					sendTelegram(
						chatID,
						"🛑 Watch stopped",
						nil,
					)

				case strings.HasPrefix(
					text,
					"WATCH ",
				):

					indexText :=
						strings.TrimPrefix(
							text,
							"WATCH ",
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

						sendTelegram(
							chatID,
							"❌ Match tidak valid",
							nil,
						)

						continue
					}

					watchMatch(
						chatID,
						matches[index-1].ID,
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
