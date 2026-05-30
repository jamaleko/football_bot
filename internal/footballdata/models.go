package footballdata

type MatchesResponse struct {
	Matches []Match `json:"matches"`
}

type Match struct {
	ID       int    `json:"id"`
	Status   string `json:"status"`

	HomeTeam Team `json:"homeTeam"`
	AwayTeam Team `json:"awayTeam"`
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
