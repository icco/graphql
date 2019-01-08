package graphql

import (
	"context"
	"time"
)

// A Tweet is an archived tweet.
type Tweet struct {
	ID            string    `json:"id"`
	Text          string    `json:"text"`
	Hashtags      []string  `json:"hashtags"`
	Symbols       []string  `json:"symbols"`
	UserMentions  []string  `json:"user_mentions"`
	Urls          []string  `json:"urls"`
	User          string    `json:"user"`
	FavoriteCount int       `json:"favorite_count"`
	RetweetCount  int       `json:"retweet_count"`
	Posted        time.Time `json:"posted"`
}

func GetTweets(ctx context.Context, limit *int, offset *int) ([]*Tweet, error) {
	return []*Tweet{}, nil
}
