package gcpauth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
)

func NewCredentialsSource(scopes ...string) (*CredentialsSource, error) {
	if len(scopes) == 0 {
		scopes = []string{"https://www.googleapis.com/auth/cloud-platform"}
	}

	creds, err := getDefaultCredentials(scopes...)
	if err != nil {
		return nil, err
	}

	return &CredentialsSource{
		creds: creds,
	}, nil
}

type CredentialsSource struct {
	creds *google.Credentials
}

func (src *CredentialsSource) GetToken() (string, error) {
	token, err := src.creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("cannot get access token from src (%T): %w", src.creds.TokenSource, err)
	}

	return token.AccessToken, nil
}

func getDefaultCredentials(scopes ...string) (*google.Credentials, error) {
	defaultCreds, err := google.FindDefaultCredentials(context.Background(), scopes...)
	if err != nil {
		return nil, fmt.Errorf("could not load default credentials: %w", err)
	}

	return defaultCreds, nil
}
