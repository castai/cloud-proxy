package gcpauth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func NewTokenSource(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
	if len(scopes) == 0 {
		scopes = []string{"https://www.googleapis.com/auth/cloud-platform"}
	}

	return google.DefaultTokenSource(ctx, scopes...)
}
