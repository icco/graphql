package graphql

import (
	"context"
	"net/url"
	"time"
)

// GenerateSocialImage reuses the blog's self-hosted Open Graph renderer.
func GenerateSocialImage(_ context.Context, title string, when time.Time) (*URI, error) {
	params := url.Values{"title": {title}, "date": {when.Format("January 2, 2006")}}
	return NewURI("https://writing.natwelch.com/api/og?" + params.Encode()), nil
}
