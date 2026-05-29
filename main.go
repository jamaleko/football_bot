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
	"sort"
)

const BOT_TOKEN = "8727181698:AAGAehZUWyQmh7CBiWNt_x6Xsevh2hrcuCE"

var TELEGRAM_API =
	"https://api.telegram.org/bot" + BOT_TOKEN

// =========================
// STRUCT
// =========================

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

		DateEvent string `json:"dateEvent"`

		StrTime string `json:"strTime"`
	} `json:"events"`
}

type Match struct {
	ID int

	Country string
	League string

	Home string
	Away string

	HomeScore int
	AwayScore int

	Status string

	Date string
	Time string
}

type UserSession struct {
	Watching bool
	Index int
}

// =========================
// GLOBAL
// =========================

var (
	lastUpdateID = 0

	cachedMatches []Match

	userSessions =
		map[int64]*UserSession{}
)

// BIG LEAGUES
var bigLeagues = []string{

 "4328", // EPL
 "4335", // La Liga
 "4331", // Bundesliga
 "4332", // Serie A
 "4334", // Ligue 1

 "4480", // UCL
 "4481", // Europa
 "4482", // Conference

 "4346", // MLS

 "4351", // Brazil Serie A
 "4500", // Argentina
 "4337", // Eredivisie
 "4344", // Portugal
 "4339", // Turkey
 "4668", // Saudi League

 "4429", // World Cup
 "4481", // Euro
 "4790", // Indonesia

 "4562", // International Friendly
}

// =========================
// MAIN
// =========================

func main() {

	rand.Seed(
		time.Now().UnixNano(),
	)

	fmt.Println("⚽ BOT RUNNING")

	for {

		getUpdates()

		time.Sleep(
			2 * time.Second,
		)
	}
}

// =========================
// TELEGRAM
// =========================

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

// =========================
// KEYBOARD
// =========================

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

// =========================
// LIVE MATCH
// =========================
func fetchLiveMatches() []Match {

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

 var matches []Match

 for _, e := range result.Events {

  id := 0

  fmt.Sscanf(
   e.IDEvent,
   "%d",
   &id,
  )

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

  match :=
   Match{

    ID: id,

    Country:
     e.StrCountry,

    League:
     e.StrLeague,

    Home:
     e.StrHomeTeam,

    Away:
     e.StrAwayTeam,

    HomeScore:
     homeScore,

    AwayScore:
     awayScore,

    Status:
     e.StrStatus,

    Date:
     e.DateEvent,

    Time:
     e.StrTime,
   }

  matches =
   append(
    matches,
    match,
   )
 }

 return matches
}
func toWIB(date string, clock string) string {

 layout :=
  "2006-01-02 15:04:05"

 t, err :=
  time.Parse(
   layout,
   date+" "+clock,
  )

 if err != nil {

  return clock
 }

 wib :=
  t.Add(7 * time.Hour)

 return wib.Format(
  "02 Jan 15:04",
 )
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

 status := m.Status

 if status == "NS" {

  if len(m.Time) >= 5 {

   status =
 "🕒 " +
  toWIB(
   m.Date,
   m.Time,
  ) +
  " WIB"

  } else {

   status = "🕒 Soon"
  }

 } else if status == "FT" {

  status = "✅ FT"

 } else {

  status = "🔴 LIVE " + status
 }

 msg.WriteString(
  fmt.Sprintf(
   "%d. 🌍 %s\n🏆 %s\n⚽ %s %d - %d %s\n⏱ %s\n\n",

   i+1,

   m.Country,
   m.League,

   m.Home,
   m.HomeScore,

   m.AwayScore,
   m.Away,

   status,
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

// =========================
// BIG MATCH
// =========================

func fetchBigMatches() []Match {

	var matches []Match

	for _, leagueID := range bigLeagues {

		link :=
			fmt.Sprintf(
				"https://www.thesportsdb.com/api/v1/json/123/eventsnextleague.php?id=%s",
				leagueID,
			)

		resp, err :=
			http.Get(link)

		if err != nil {
			continue
		}

		body, _ :=
			io.ReadAll(resp.Body)

		resp.Body.Close()

		var result SportsDBResponse

		json.Unmarshal(
			body,
			&result,
		)

		for _, e := range result.Events {

			id := 0

			fmt.Sscanf(
				e.IDEvent,
				"%d",
				&id,
			)

			match :=
				Match{

					ID: id,

					Country:
						e.StrCountry,

					League:
						e.StrLeague,

					Home:
						e.StrHomeTeam,

					Away:
						e.StrAwayTeam,

					Date:
						e.DateEvent,

					Time:
						e.StrTime,
				}

			matches =
				append(
					matches,
					match,
				)
		}
	}
	sort.Slice(
	 matches,
	 func(i, j int) bool {
	
	  a :=
	   matches[i].Date + " " +
	    matches[i].Time
	
	  b :=
	   matches[j].Date + " " +
	    matches[j].Time
	
	  return a < b
	 },
	)
	return matches
}

func sendBigMatches(
	chatID int64,
) {

	matches :=
		fetchBigMatches()

	if len(matches) == 0 {

		sendTelegram(
			chatID,
			"❌ Tidak ada big match",
			nil,
		)

		return
	}

	var msg strings.Builder

	msg.WriteString(
		"🔥 BIG MATCHES\n\n",
	)

	limit := 10

	if len(matches) < limit {
		limit = len(matches)
	}

	for i := 0; i < limit; i++ {

		m := matches[i]

		msg.WriteString(
 fmt.Sprintf(
  "%d. 🌍 %s\n🏆 %s\n⚽ %s vs %s\n🕒 %s WIB\n\n",

  i+1,

  m.Country,
  m.League,

  m.Home,
  m.Away,

  toWIB(
   m.Date,
   m.Time,
  ),
 ),
)
	}

	sendTelegram(
		chatID,
		msg.String(),
		nil,
	)
}

// =========================
// WATCH
// =========================

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
			Index: number - 1,
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
			Index: index,
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
		session.Index,
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

	m :=
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

			m.Country,
			m.League,

			m.Home,
			m.HomeScore,

			m.AwayScore,
			m.Away,

			m.Status,
		)

	sendTelegram(
		chatID,
		msg,
		nil,
	)
}

// =========================
// SEND TELEGRAM
// =========================

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
