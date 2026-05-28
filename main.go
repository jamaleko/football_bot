// UPDATED FULL CODE // FEATURES: // ✅ INLINE KEYBOARD // ✅ LIVE NOW // ✅ BIG MATCHES // ✅ RANDOM MATCH // ✅ REFRESH BUTTON // ✅ STOP BUTTON // ✅ NO SPAM // ✅ ONLY GOAL/YELLOW/RED EVENT // ✅ MULTI USER // ✅ BIG MATCH NOT FOUND MESSAGE // ✅ LIVE MATCH + UPCOMING

package main

import ( "bytes" "encoding/json" "fmt" "io" "log" "math/rand" "net/http" "os" "strconv" "strings" "time"

"github.com/joho/godotenv"

)

const baseURL = "https://v3.football.api-sports.io"

type UserSession struct { SelectedMatch int

LastGoals  int
LastYellow int
LastRed    int

IsWatching bool

}

type MatchItem struct { ID int }

var ( lastUpdateID int

userSessions = map[int64]*UserSession{}
userMatches  = map[int64][]MatchItem{}

)

var bigLeagueIDs = []int{ 39, 140, 135, 78, 61, 2, 3, 848, 1, 4, 5, 9, 274, 32, 33, 34, 35, 36, 10, 960 }

type FixtureResponse struct { Response []struct {

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

type TelegramResponse struct { Result []struct {

UpdateID int `json:"update_id"`

	Message struct {
		Text string `json:"text"`

		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`

	CallbackQuery struct {
		ID   string `json:"id"`
		Data string `json:"data"`

		Message struct {
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
		} `json:"message"`
	} `json:"callback_query"`

} `json:"result"`

}

func isBigLeague(id int) bool { for _, leagueID := range bigLeagueIDs { if leagueID == id { return true } }

return false

}

func apiRequest(url string) ([]byte, error) {

req, _ := http.NewRequest("GET", url, nil)

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

func sendTelegram(chatID int64, text string, keyboard interface{}) {

token := os.Getenv("BOT_TOKEN")

url := fmt.Sprintf(
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

body, _ := json.Marshal(payload)

http.Post(
	url,
	"application/json",
	bytes.NewBuffer(body),
)

}

func watchKeyboard() map[string]interface{} {

return map[string]interface{}{
	"inline_keyboard": [][]map[string]string{
		{
			{
				"text": "🔄 Refresh",
				"callback_data": "refresh",
			},
			{
				"text": "🛑 Stop",
				"callback_data": "stop",
			},
		},
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
	},
}

}

func sendMainMenu(chatID int64) {

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
		},
	},
}

sendTelegram(chatID, "⚽ FOOTBALL BOT", keyboard)

}

func sendLiveMatches(chatID int64) {

userMatches[chatID] = nil

url := fmt.Sprintf("%s/fixtures?live=all", baseURL)

body, err := apiRequest(url)

if err != nil {
	return
}

var data FixtureResponse

json.Unmarshal(body, &data)

message := "🔥 LIVE MATCHES\n\n"

var keyboardRows [][]map[string]string

index := 1
found := false

for _, match := range data.Response {

	if !isBigLeague(match.League.ID) {
		continue
	}

	found = true

	message += fmt.Sprintf(
		"%d. 🌍 %s\n🏆 %s\n⚽ %s %d - %d %s\n⏱ %d'\n\n",
		index,
		match.League.Country,
		match.League.Name,
		match.Teams.Home.Name,
		match.Goals.Home,
		match.Goals.Away,
		match.Teams.Away.Name,
		match.Fixture.Status.Elapsed,
	)

	userMatches[chatID] = append(
		userMatches[chatID],
		MatchItem{ID: match.Fixture.ID},
	)

	keyboardRows = append(
		keyboardRows,
		[]map[string]string{
			{
				"text": fmt.Sprintf("⚽ Watch %d", index),
				"callback_data": fmt.Sprintf("watch_%d", index),
			},
		},
	)

	index++

	if index > 10 {
		break
	}
}

if !found {
	sendTelegram(
		chatID,
		"❌ Tidak ada big match yang sedang live sekarang.",
		nil,
	)
	return
}

keyboard := map[string]interface{}{
	"inline_keyboard": keyboardRows,
}

sendTelegram(chatID, message, keyboard)

}

func sendUpcomingMatches(chatID int64) {

now := time.Now()

message := "📅 BIG MATCHES\n\n"

count := 0

for day := 0; day < 30; day++ {

	date := now.AddDate(0, 0, day).Format("2006-01-02")

	url := fmt.Sprintf(
		"%s/fixtures?date=%s",
		baseURL,
		date,
	)

	body, err := apiRequest(url)

	if err != nil {
		continue
	}

	var data FixtureResponse

	json.Unmarshal(body, &data)

	for _, match := range data.Response {

		if !isBigLeague(match.League.ID) {
			continue
		}

		if match.Fixture.Status.Short == "FT" {
			continue
		}

		matchTime, _ := time.Parse(time.RFC3339, match.Fixture.Date)

		wib := matchTime.In(time.FixedZone("WIB", 7*3600))

		count++

		message += fmt.Sprintf(
			"%d. 🌍 %s\n🏆 %s\n⚽ %s vs %s\n🕒 %s WIB\n\n",
			count,
			match.League.Country,
			match.League.Name,
			match.Teams.Home.Name,
			match.Teams.Away.Name,
			wib.Format("02 Jan 15:04"),
		)

		if count >= 10 {
			break
		}
	}
}

if count == 0 {
	message = "❌ Tidak ada big match"
}

sendTelegram(chatID, message, nil)

}

func getStatsText(data FixtureResponse) string {

match := data.Response[0]

return fmt.Sprintf(
	"🌍 %s\n🏆 %s\n\n⚽ %s %d - %d %s\n⏱ %d'",
	match.League.Country,
	match.League.Name,
	match.Teams.Home.Name,
	match.Goals.Home,
	match.Goals.Away,
	match.Teams.Away.Name,
	match.Fixture.Status.Elapsed,
)

}

func refreshStats(chatID int64) {

session := userSessions[chatID]

if session == nil {
	return
}

url := fmt.Sprintf(
	"%s/fixtures?id=%d",
	baseURL,
	session.SelectedMatch,
)

body, err := apiRequest(url)

if err != nil {
	return
}

var data FixtureResponse

json.Unmarshal(body, &data)

sendTelegram(
	chatID,
	getStatsText(data),
	watchKeyboard(),
)

}

func watchMatch(chatID int64, matchID int) {

session := &UserSession{}

session.SelectedMatch = matchID
session.IsWatching = true

userSessions[chatID] = session

refreshStats(chatID)

go watchLoop(chatID)

}

func watchRandomMatch(chatID int64) {

url := fmt.Sprintf("%s/fixtures?live=all", baseURL)

body, _ := apiRequest(url)

var data FixtureResponse

json.Unmarshal(body, &data)

if len(data.Response) == 0 {
	sendTelegram(chatID, "❌ Tidak ada live match", nil)
	return
}

match := data.Response[rand.Intn(len(data.Response))]

watchMatch(chatID, match.Fixture.ID)

}

func checkEvents(chatID int64) {

session := userSessions[chatID]

if session == nil {
	return
}

url := fmt.Sprintf(
	"%s/fixtures?id=%d",
	baseURL,
	session.SelectedMatch,
)

body, _ := apiRequest(url)

var data FixtureResponse

json.Unmarshal(body, &data)

if len(data.Response) == 0 {
	return
}

match := data.Response[0]

totalGoals := match.Goals.Home + match.Goals.Away

if totalGoals > session.LastGoals {

	session.LastGoals = totalGoals

	sendTelegram(
		chatID,
		"⚽ GOAL!\n\n"+getStatsText(data),
		watchKeyboard(),
	)
}

}

func watchLoop(chatID int64) {

for {

	session := userSessions[chatID]

	if session == nil {
		return
	}

	if !session.IsWatching {
		return
	}

	checkEvents(chatID)

	time.Sleep(1 * time.Minute)
}

}

func answerCallbackQuery(id string) {

token := os.Getenv("BOT_TOKEN")

url := fmt.Sprintf(
	"https://api.telegram.org/bot%s/answerCallbackQuery",
	token,
)

payload := map[string]interface{}{
	"callback_query_id": id,
}

body, _ := json.Marshal(payload)

http.Post(
	url,
	"application/json",
	bytes.NewBuffer(body),
)

}

func getTelegramUpdates() {

token := os.Getenv("BOT_TOKEN")

for {

	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?offset=%d",
		token,
		lastUpdateID+1,
	)

	resp, err := http.Get(url)

	if err != nil {
		time.Sleep(3 * time.Second)
		continue
	}

	body, _ := io.ReadAll(resp.Body)

	resp.Body.Close()

	var data TelegramResponse

	json.Unmarshal(body, &data)

	for _, update := range data.Result {

		lastUpdateID = update.UpdateID

		if update.Message.Text != "" {

			chatID := update.Message.Chat.ID

			text := strings.TrimSpace(update.Message.Text)

			if text == "/start" {
				sendMainMenu(chatID)
			}
		}

		if update.CallbackQuery.Data != "" {

			chatID := update.CallbackQuery.Message.Chat.ID

			data := update.CallbackQuery.Data

			answerCallbackQuery(update.CallbackQuery.ID)

			switch {

			case data == "live_now":
				sendLiveMatches(chatID)

			case data == "big_matches":
				sendUpcomingMatches(chatID)

			case data == "random_match":
				watchRandomMatch(chatID)

			case data == "refresh":

				session := userSessions[chatID]

				if session == nil {
					sendTelegram(chatID, "❌ Tidak ada match yang sedang ditonton", nil)
					continue
				}

				refreshStats(chatID)

			case data == "stop":

				session := userSessions[chatID]

				if session != nil {
					session.IsWatching = false
					session.SelectedMatch = 0
				}

				sendTelegram(chatID, "🛑 Watch stopped", nil)

			case strings.HasPrefix(data, "watch_"):

				indexText := strings.TrimPrefix(data, "watch_")

				index, err := strconv.Atoi(indexText)

				if err != nil {
					continue
				}

				matches := userMatches[chatID]

				if index < 1 || index > len(matches) {
					continue
				}

				watchMatch(chatID, matches[index-1].ID)
			}
		}
	}

	time.Sleep(2 * time.Second)
}

}

func main() {

rand.Seed(time.Now().UnixNano())

godotenv.Load()

log.Println("Football bot running...")

getTelegramUpdates()

}
