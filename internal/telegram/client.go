package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	Token string
}

func New(token string) *Client {
	return &Client{
		Token: token,
	}
}

func (c *Client) Send(chatID int64, text string) error {

	u := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		c.Token,
	)

	resp, err := http.PostForm(
		u,
		url.Values{
			"chat_id": {fmt.Sprintf("%d", chatID)},
			"text":    {text},
		},
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

type UpdateResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	UpdateID int `json:"update_id"`

	Message struct {
		MessageID int `json:"message_id"`

		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`

		Text string `json:"text"`
	} `json:"message"`
}

func (c *Client) Updates(offset int) (*UpdateResponse, error) {

	u := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?offset=%d",
		c.Token,
		offset,
	)

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result UpdateResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
