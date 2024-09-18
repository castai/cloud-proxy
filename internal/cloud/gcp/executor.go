//go:generate mockgen -destination ./mock/gcp.go -package=mock_gcp cloud-proxy/internal/cloud/gcp Credentials
package gcp

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type Credentials interface {
	GetToken() (string, error)
}
type Client struct {
	httpClient *http.Client
}

func New(tokenSource oauth2.TokenSource) *Client {
	client := oauth2.NewClient(context.Background(), tokenSource)
	return &Client{httpClient: client}
}

func (c *Client) DoHTTPRequest(request *http.Request) (*http.Response, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: request %+v error: %w", request, err)
	}

	return resp, nil
}
