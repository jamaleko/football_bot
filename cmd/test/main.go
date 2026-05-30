package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {

	req, _ := http.NewRequest(
		"GET",
		"https://api.football-data.org/v4/competitions/PL",
		nil,
	)

	req.Header.Set("X-Auth-Token", "40d8ae2308d148afa834e6253fab62fd")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	fmt.Println("status:", resp.Status)

	body, _ := io.ReadAll(resp.Body)

	fmt.Println(string(body))
}
