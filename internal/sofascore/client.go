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
        "https://www.sofascore.com/api/v1/event/12437786",
        date,
    )
    fmt.Println("url:", url)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Referer", "https://www.sofascore.com/")

    resp, err := c.http.Do(req)
    if err != nil {
        return nil, err
    }

    fmt.Println("status:", resp.Status)

    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    fmt.Println(string(body))

    return nil, nil
}
