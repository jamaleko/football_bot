package main

import (
	"fmt"

	"football_bot/internal/footballdata"
)

func main() {

	client := footballdata.New("40d8ae2308d148afa834e6253fab62fd")

	matches, err := client.WorldCupMatches()
	if err != nil {
		panic(err)
	}

	for _, m := range matches {
	    fmt.Printf(
	        "%d | %s vs %s\n",
	        m.ID,
	        m.HomeTeam.Name,
	        m.AwayTeam.Name,
	    )
	}
}
