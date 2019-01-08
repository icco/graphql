package graphql

import (
	"context"
	"time"

	"github.com/lib/pq"
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
	FavoriteCount int64     `json:"favorite_count"`
	RetweetCount  int64     `json:"retweet_count"`
	Posted        time.Time `json:"posted"`
}

// Save inserts or updates a tweet into the database.
func (t *Tweet) Save(ctx context.Context) error {
	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO tweets(id, text, hashtags, symbols, user_mentions, urls, user, favorites, retweets, posted, created_at, modified_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
ON CONFLICT (id) DO UPDATE
SET (text, hashtags, symbols, user_mentions, urls, user, favorites, retweets, posted, modified_at) = ($2, $3, $4, $5, $6, $7, $8, $9, $10)
WHERE tweets.id = $1;
`,
		t.ID,
		t.Text,
		pq.Array(t.Hashtags),
		pq.Array(t.Symbols),
		pq.Array(t.UserMentions),
		pq.Array(t.Urls),
		t.User,
		t.FavoriteCount,
		t.RetweetCount,
		t.Posted,
		time.Now(),
	); err != nil {
		return err
	}

	return nil
}

func GetTweets(ctx context.Context, limit *int, offset *int) ([]*Tweet, error) {
	return []*Tweet{}, nil
}
