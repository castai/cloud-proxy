package gcpauth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/serviceusage/v1"
)

func NewTokenSource(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
	if len(scopes) == 0 {
		scopes = []string{serviceusage.CloudPlatformScope}
	}

	return google.DefaultTokenSource(ctx, scopes...)
}
