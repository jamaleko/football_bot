package sofascore

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Client struct {
    http *http.Client
}

func New() *Client {
    return &Client{
        http: &http.Client{},
    }
}

func (c *Client) ScheduledEvents(date string) ([]Event, error) {

    url := fmt.Sprintf(
        "https://www.sofascore.com/api/v1/sport/football/scheduled-events/%s",
        date,
    )

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("User-Agent", "Mozilla/5.0")

    resp, err := c.http.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result ScheduledEventsResponse

    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return nil, err
    }
    fmt.Println("events:", len(Events))

    return result.Events, nil
}
