package main

import (
	"fmt"

	"football_bot/internal/footballdata"
)

func main() {

	client := footballdata.New("40d8ae2308d148afa834e6253fab62fd")

	wc, err := client.WorldCupMatches()
	if err != nil {
		panic(err)
	}
	
	cl, err := client.ChampionsLeagueMatches()
	if err != nil {
		panic(err)
	}
	
	matches := append(cl, wc...)
	
	for i, m := range matches {
		fmt.Printf(
			"%d | %s vs %s\n",
			i+1,
			m.HomeTeam.Name,
			m.AwayTeam.Name,
		)
	}
}
