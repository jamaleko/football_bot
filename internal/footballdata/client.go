package footballdata

type MatchesResponse struct {
	Matches []Match `json:"matches"`
}

type Match struct {
	ID       int  `json:"id"`
	Status   string `json:"status"`

	HomeTeam Team `json:"homeTeam"`
	AwayTeam Team `json:"awayTeam"`

	Score Score `json:"score"`
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Score struct {
	FullTime FullTime `json:"fullTime"`
}

type FullTime struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}
