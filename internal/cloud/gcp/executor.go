//go:generate mockgen -destination ./mock/gcp.go -package=mock_gcp cloud-proxy/internal/cloud/gcp Credentials
package gcp

import (
	"cloud-proxy/internal/cloud/gcp/gcpauth"
	"fmt"
	"net/http"
)

type Credentials interface {
	GetToken() (string, error)
}
type Client struct {
	credentials Credentials
	httpClient  *http.Client
}

func New(credentials *gcpauth.CredentialsSource, client *http.Client) *Client {
	return &Client{credentials: credentials, httpClient: client}
}

func (c *Client) DoHTTPRequest(request *http.Request) (*http.Response, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}

	token, err := c.credentials.GetToken()
	if err != nil {
		return nil, fmt.Errorf("credentialsSrc.GetToken: error: %w", err)
	}
	// Set the authorize header manually since we can't rely on mothership auth.
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: request %+v error: %w", request, err)
	}

	return resp, nil
}
