package gcpauth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
)

func NewCredentialsSource(scopes ...string) *CredentialsSource {
	if len(scopes) == 0 {
		scopes = []string{"https://www.googleapis.com/auth/cloud-platform"}
	}
	return &CredentialsSource{
		scopes: scopes,
	}
}

type CredentialsSource struct {
	scopes []string
}

// TODO: check if we should be doing it constantly; cache them; cache the token or something else

func (src *CredentialsSource) getDefaultCredentials() (*google.Credentials, error) {
	defaultCreds, err := google.FindDefaultCredentials(context.Background(), src.scopes...)
	if err != nil {
		return nil, fmt.Errorf("could not load default credentials: %w", err)
	}
	return defaultCreds, nil
}
func (src *CredentialsSource) GetToken() (string, error) {
	credentials, err := src.getDefaultCredentials()
	if err != nil {
		return "", fmt.Errorf("cannot load GCP credentials: %w", err)
	}

	token, err := credentials.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("cannot get access token from src (%T): %w", credentials.TokenSource, err)
	}

	return token.AccessToken, nil
}
