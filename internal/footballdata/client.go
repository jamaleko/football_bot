package footballdata

import (
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
