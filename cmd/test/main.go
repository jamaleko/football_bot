package main

import (
 "encoding/json"
 "fmt"
 "net/http"
 "os"
)

func main() {

 req, err := http.NewRequest(
  "GET",
  "https://api.football-data.org/v4/matches/552096",
  nil,
 )
 if err != nil {
  panic(err)
 }

 req.Header.Set(
  "X-Auth-Token",
  os.Getenv("TOKEN"),
 )

 resp, err := http.DefaultClient.Do(req)
 if err != nil {
  panic(err)
 }
 defer resp.Body.Close()

 var result map[string]any

 err = json.NewDecoder(resp.Body).Decode(&result)
 if err != nil {
  panic(err)
 }

 b, _ := json.MarshalIndent(
  result,
  "",
  "  ",
 )

 fmt.Println(string(b))
}
