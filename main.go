package main

import (
 "bytes"
 "encoding/json"
 "fmt"
 "io"
 "log"
 "net/http"
 "os"
 "time"

 "github.com/joho/godotenv"
)

const (
 baseURL = "https://v3.football.api-sports.io"
)

var (
 lastGoals = map[int]int{}
)

type FixtureResponse struct {
 Response []struct {

  Fixture struct {
   ID int json:"id"

   Status struct {
    Elapsed int json:"elapsed"
    Short   string json:"short"
   } json:"status"

   Date string json:"date"
  } json:"fixture"

  Teams struct {
   Home struct {
    Name string json:"name"
   } json:"home"

   Away struct {
    Name string json:"name"
   } json:"away"
  } json:"teams"

  Goals struct {
   Home int json:"home"
   Away int json:"away"
  } json:"goals"

  Statistics []struct {

   Team struct {
    Name string json:"name"
   } json:"team"

   Statistics []struct {
    Type  string      json:"type"
    Value interface{} json:"value"
   } json:"statistics"

  } json:"statistics"

 } json:"response"
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

 message := "📅 Match Hari Ini\n\n"

 count := 0

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
   "⚽ %s vs %s\n🕒 %s WIB\n\n",

   match.Teams.Home.Name,
   match.Teams.Away.Name,

   wib.Format("15:04"),
  )

  count++

  if count >= 10 {
   break
  }
 }

 sendTelegram(message)
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

 if len(data.Response) == 0 {

  log.Println("no live match")

  return 0
 }

 for _, match := range data.Response {

  fixtureID := match.Fixture.ID

  totalGoals :=
   match.Goals.Home +
    match.Goals.Away

  oldGoals :=
   lastGoals[fixtureID]

  if totalGoals > oldGoals {

   lastGoals[fixtureID] =
    totalGoals

   sendTelegram(
    fmt.Sprintf(
     "⚽ GOAL!\n\n"+
      "%s %d - %d %s\n\n"+
      "⏱ %d'",

     match.Teams.Home.Name,
     match.Goals.Home,

     match.Goals.Away,
     match.Teams.Away.Name,

     match.Fixture.Status.Elapsed,
    ),
   )
  }

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

  log.Printf(
   "%s %d - %d %s | Possession %s vs %s",

   match.Teams.Home.Name,
   match.Goals.Home,

   match.Goals.Away,
   match.Teams.Away.Name,

   homePossession,
   awayPossession,
  )
 }

 return len(data.Response)
}

func main() {

 godotenv.Load()

 sendTelegram(
  "⚽ Football Bot Started",
 )
sendTodayMatches()

 for {

  liveCount :=
   getLiveMatches()

  if liveCount == 0 {

   time.Sleep(
    30 * time.Minute,
   )

  } else {

   time.Sleep(
    5 * time.Minute,
   )
  }
 }
}
