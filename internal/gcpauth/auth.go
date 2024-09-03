package gcpauth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
)

type GCPCredentialsSource struct {
}

// TODO: check if we should be doing it constantly; cache them; cache the token or something else

func (src *GCPCredentialsSource) GetDefaultCredentials() (*google.Credentials, error) {
	defaultCreds, err := google.FindDefaultCredentials(context.Background(), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("could not load default credentials: %w", err)
	}
	return defaultCreds, nil
}
