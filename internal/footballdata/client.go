package footballdata

import (
	"encoding/json"
	"net/http"
)

type Client struct {
	token string
	http  *http.Client
}

func New(token string) *Client {
	return &Client{
		token: token,
		http:  &http.Client{},
	}
}

func (c *Client) PremierLeagueMatches() ([]Match, error) {

	req, err := http.NewRequest(
		"GET",
		"https://api.football-data.org/v4/competitions/PL/matches",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result MatchesResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Matches, nil
}
