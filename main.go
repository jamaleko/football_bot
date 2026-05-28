package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	BOT_TOKEN = os.Getenv("BOT_TOKEN")
)

var (
	userWatch = map[int]int{}

	bigLeagues = []int{
		39,  // Premier League
		140, // La Liga
		135, // Serie A
		78,  // Bundesliga
		61,  // Ligue 1
		2,   // Champions League
		3,   // Europa League
		848, // Conference League
		1,   // World Cup
		4,   // Euro Cup
		71,  // Brasileirão
		94,  // Liga 1 Indonesia
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
)

func main() {
	sendBotMenu()

	offset := 0

	for {
		updates := getUpdates(offset)

		for _, update := range updates {

			offset = update.UpdateID + 1

			if update.Message.Text != "" {
				handleMessage(update.Message)
			}

			if update.CallbackQuery.ID != "" {
				handleCallback(update.CallbackQuery)
			}
		}

		time.Sleep(2 * time.Second)
	}
}

type UpdateResponse struct {
	Result []Update `json:"result"`
}

type Update struct {
	UpdateID      int           `json:"update_id"`
	Message       Message       `json:"message"`
	CallbackQuery CallbackQuery `json:"callback_query"`
}

type CallbackQuery struct {
	ID      string  `json:"id"`
	From    User    `json:"from"`
	Message Message `json:"message"`
	Data    string  `json:"data"`
}

type User struct {
	ID int `json:"id"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ID int `json:"id"`
}

func getUpdates(offset int) []Update {

	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?timeout=30&offset=%d",
		BOT_TOKEN,
		offset,
	)

	resp, _ := http.Get(url)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data UpdateResponse
	json.Unmarshal(body, &data)

	return data.Result
}

func handleMessage(msg Message) {

	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)

	if text == "/start" {

		sendBotMenuTo(chatID)
		return
	}
}

func handleCallback(cb CallbackQuery) {

	chatID := cb.Message.Chat.ID
	data := cb.Data

	answerCallbackQuery(cb.ID)

	switch data {

	case "live_now":

		matches := getLiveMatches(false)

		if len(matches) == 0 {

			sendTelegram(chatID,
				"❌ Tidak ada pertandingan live sekarang.")
			return
		}

		sendLiveMatches(chatID, matches)

	case "live_big":

		matches := getLiveMatches(true)

		if len(matches) == 0 {

			sendTelegram(chatID,
				"❌ Tidak ada big match live sekarang.")
			return
		}

		sendLiveMatches(chatID, matches)

	case "random_match":

		matches := getLiveMatches(false)

		if len(matches) == 0 {

			sendTelegram(chatID,
				"❌ Tidak ada live match.")
			return
		}

		match := matches[0]

		fixtureID := int(match["fixture"].(map[string]interface{})["id"].(float64))

		userWatch[chatID] = fixtureID

		sendStats(chatID, fixtureID)

	case "refresh_stats":

		fixtureID := userWatch[chatID]

		if fixtureID == 0 {

			sendTelegram(chatID,
				"❌ Tidak ada match yang sedang ditonton.")
			return
		}

		sendStats(chatID, fixtureID)

	case "stop_watch":

		delete(userWatch, chatID)

		sendTelegram(chatID,
			"🛑 Watch stopped.")

	case "fixtures":

		sendFixtures(chatID)
	}
}

func sendBotMenu() {
	fmt.Println("⚽ Football Bot Started")
}

func sendBotMenuTo(chatID int) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{
				{
					"text": "📺 Live Now",
					"callback_data": "live_now",
				},
				{
					"text": "🔥 Live Big Match",
					"callback_data": "live_big",
				},
			},
			{
				{
					"text": "📅 Big Fixtures",
					"callback_data": "fixtures",
				},
				{
					"text": "🎲 Random Match",
					"callback_data": "random_match",
				},
			},
		},
	}

	sendPayload(chatID,
		"⚽ FOOTBALL BOT",
		keyboard)
}

func sendFixtures(chatID int) {

	date := time.Now().Format("2006-01-02")

	url := fmt.Sprintf(
		"https://v3.football.api-sports.io/fixtures?date=%s",
		date,
	)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("x-apisports-key",
		os.Getenv("API_KEY"))

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	response := result["response"].([]interface{})

	text := "📅 BIG MATCHES\n\n"

	count := 0

	for _, item := range response {

		match := item.(map[string]interface{})

		league := match["league"].(map[string]interface{})
		leagueID := int(league["id"].(float64))

		if !contains(bigLeagues, leagueID) {
			continue
		}

		teams := match["teams"].(map[string]interface{})

		home := teams["home"].(map[string]interface{})["name"]
		away := teams["away"].(map[string]interface{})["name"]

		country := league["country"]
		leagueName := league["name"]

		fixture := match["fixture"].(map[string]interface{})
		dateUTC := fixture["date"].(string)

		t, _ := time.Parse(time.RFC3339, dateUTC)

		wib := t.Add(7 * time.Hour)

		count++

		text += fmt.Sprintf(
			"%d. 🌍 %s\n🏆 %s\n⚽ %s vs %s\n🕒 %s WIB\n\n",
			count,
			country,
			leagueName,
			home,
			away,
			wib.Format("02 Jan 15:04"),
		)

		if count >= 10 {
			break
		}
	}

	if count == 0 {

		text = "❌ Tidak ada big match hari ini."
	}

	sendTelegram(chatID, text)
}

func sendLiveMatches(chatID int, matches []interface{}) {

	text := "🔥 LIVE MATCHES\n\n"

	count := 0

	for _, item := range matches {

		match := item.(map[string]interface{})

		league := match["league"].(map[string]interface{})
		country := league["country"]
		leagueName := league["name"]

		teams := match["teams"].(map[string]interface{})

		home := teams["home"].(map[string]interface{})["name"]
		away := teams["away"].(map[string]interface{})["name"]

		goals := match["goals"].(map[string]interface{})

		homeGoal := int(goals["home"].(float64))
		awayGoal := int(goals["away"].(float64))

		fixture := match["fixture"].(map[string]interface{})
		status := fixture["status"].(map[string]interface{})

		minute := int(status["elapsed"].(float64))

		count++

		text += fmt.Sprintf(
			"%d. 🌍 %s\n🏆 %s\n⚽ %s %d - %d %s\n⏱ %d'\n\n",
			count,
			country,
			leagueName,
			home,
			homeGoal,
			awayGoal,
			away,
			minute,
		)

		if count >= 10 {
			break
		}
	}

	sendTelegram(chatID, text)
}

func getLiveMatches(bigOnly bool) []interface{} {

	url := "https://v3.football.api-sports.io/fixtures?live=all"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("x-apisports-key",
		os.Getenv("API_KEY"))

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	response := result["response"].([]interface{})

	final := []interface{}{}

	for _, item := range response {

		match := item.(map[string]interface{})

		league := match["league"].(map[string]interface{})
		leagueID := int(league["id"].(float64))

		if bigOnly {

			if contains(bigLeagues, leagueID) {
				final = append(final, item)
			}

		} else {

			final = append(final, item)
		}
	}

	return final
}

func sendStats(chatID int, fixtureID int) {

	url := fmt.Sprintf(
		"https://v3.football.api-sports.io/fixtures/statistics?fixture=%d",
		fixtureID,
	)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("x-apisports-key",
		os.Getenv("API_KEY"))

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	response := result["response"].([]interface{})

	if len(response) < 2 {

		sendTelegram(chatID,
			"❌ Statistik tidak tersedia.")
		return
	}

	home := response[0].(map[string]interface{})
	away := response[1].(map[string]interface{})

	homeTeam := home["team"].(map[string]interface{})["name"]
	awayTeam := away["team"].(map[string]interface{})["name"]

	text := fmt.Sprintf(
		"⚽ %s vs %s\n\n",
		homeTeam,
		awayTeam,
	)

	for _, stat := range home["statistics"].([]interface{}) {

		s := stat.(map[string]interface{})

		statType := s["type"].(string)

		homeVal := fmt.Sprintf("%v", s["value"])

		awayVal := "-"

		for _, a := range away["statistics"].([]interface{}) {

			as := a.(map[string]interface{})

			if as["type"] == statType {

				awayVal = fmt.Sprintf("%v", as["value"])
			}
		}

		text += fmt.Sprintf(
			"%s\n%s vs %s\n\n",
			statType,
			homeVal,
			awayVal,
		)
	}

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{
				{
					"text": "🔄 Refresh",
					"callback_data": "refresh_stats",
				},
				{
					"text": "🛑 Stop Watch",
					"callback_data": "stop_watch",
				},
			},
			{
				{
					"text": "📺 Live Matches",
					"callback_data": "live_now",
				},
			},
		},
	}

	sendPayload(chatID, text, keyboard)
}

func sendTelegram(chatID int, text string) {

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text": text,
	}

	send(payload)
}

func sendPayload(
	chatID int,
	text string,
	keyboard map[string]interface{},
) {

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text": text,
		"reply_markup": keyboard,
	}

	send(payload)
}

func send(payload map[string]interface{}) {

	jsonData, _ := json.Marshal(payload)

	http.Post(
		fmt.Sprintf(
			"https://api.telegram.org/bot%s/sendMessage",
			BOT_TOKEN,
		),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
}

func answerCallbackQuery(id string) {

	payload := map[string]interface{}{
		"callback_query_id": id,
	}

	jsonData, _ := json.Marshal(payload)

	http.Post(
		fmt.Sprintf(
			"https://api.telegram.org/bot%s/answerCallbackQuery",
			BOT_TOKEN,
		),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
}

func contains(arr []int, value int) bool {

	for _, v := range arr {

		if v == value {
			return true
		}
	}

	return false
}

func init() {

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Paste API KEY: ")

	key, _ := reader.ReadString('\n')

	os.Setenv("API_KEY",
		strings.TrimSpace(key))

	fmt.Println("✅ API Ready")
}
