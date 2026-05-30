package sofascore

import (
    //"encoding/json"
    "fmt"
    "net/http"
    "io"
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
    fmt.Println("status:", resp.Status)
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    
    fmt.Println(string(body))
    
    return nil, nil
    if err != nil {
        return nil, err
    }

    return result.Events, nil
}
