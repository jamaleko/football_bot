package main

import (
	"fmt"
	"os"
	"football_bot/internal/footballdata"
)

func main() {

	client := footballdata.New(os.Getenv("TOKEN"))

	match, err := client.Match(552096)
	if err != nil {
		panic(err)
	}

	fmt.Println(match.ID)
	fmt.Println(match.Status)
	fmt.Println(match.HomeTeam.Name)
	fmt.Println(match.AwayTeam.Name)
}
