package sofascore

type ScheduledEventsResponse struct {
    Events []Event `json:"events"`
}

type Event struct {
    ID int64 `json:"id"`

    Tournament struct {
        Name string `json:"name"`
    } `json:"tournament"`

    HomeTeam struct {
        Name string `json:"name"`
    } `json:"homeTeam"`

    AwayTeam struct {
        Name string `json:"name"`
    } `json:"awayTeam"`
}
