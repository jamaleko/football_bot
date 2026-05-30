package main

import (
    "fmt"
    "time"

    "football_bot/internal/sofascore"
)

func main() {

    client := sofascore.New()

    today := time.Now().Format("2006-01-02")

    events, err := client.ScheduledEvents(today)
    if err != nil {
        panic(err)
    }

    for _, e := range events {

        fmt.Printf(
            "%s\n%s vs %s\nEventID: %d\n\n",
            e.Tournament.Name,
            e.HomeTeam.Name,
            e.AwayTeam.Name,
            e.ID,
        )
    }
}
