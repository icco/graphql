package graphql

import (
	"context"
	"database/sql"
	"fmt"
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
	ScreenName    string    `json:"screen_name"`
	FavoriteCount int       `json:"favorite_count"`
	RetweetCount  int       `json:"retweet_count"`
	Posted        time.Time `json:"posted"`
}

// Save inserts or updates a tweet into the database.
func (t *Tweet) Save(ctx context.Context) error {
	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO tweets(id, text, hashtags, symbols, user_mentions, urls, screen_name, favorites, retweets, posted, created_at, modified_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
ON CONFLICT (id) DO UPDATE
SET (text, hashtags, symbols, user_mentions, urls, screen_name, favorites, retweets, posted, modified_at) = ($2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
WHERE tweets.id = $1;
`,
		t.ID,
		t.Text,
		pq.Array(t.Hashtags),
		pq.Array(t.Symbols),
		pq.Array(t.UserMentions),
		pq.Array(t.Urls),
		t.ScreenName,
		t.FavoriteCount,
		t.RetweetCount,
		t.Posted,
		time.Now(),
	); err != nil {
		return err
	}

	return nil
}

// GetTweet returns a single tweet by id.
func GetTweet(ctx context.Context, id string) (*Tweet, error) {
	var tweet Tweet
	row := db.QueryRowContext(ctx, "SELECT id, text, hashtags, symbols, user_mentions, urls, screen_name, favorites, retweets, posted FROM tweets WHERE id = $1", id)
	err := row.Scan(&tweet.ID, &tweet.Text, pq.Array(&tweet.Hashtags), pq.Array(&tweet.Symbols), pq.Array(&tweet.UserMentions), pq.Array(&tweet.Urls), &tweet.ScreenName, &tweet.FavoriteCount, &tweet.RetweetCount, &tweet.Posted)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("No tweet %s", id)
	case err != nil:
		return nil, fmt.Errorf("Error running get query: %+v", err)
	default:
		return &tweet, nil
	}
}

// GetTweets returns an array of tweets from the database.
func GetTweets(ctx context.Context, limitIn *int, offsetIn *int) ([]*Tweet, error) {
	limit := 10
	if limitIn != nil {
		limit = *limitIn
	}

	offset := 0
	if offsetIn != nil {
		offset = *offsetIn
	}

	rows, err := db.QueryContext(ctx, "SELECT id, text, hashtags, symbols, user_mentions, urls, screen_name, favorites, retweets, posted FROM tweets ORDER BY posted DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tweets := make([]*Tweet, 0)
	for rows.Next() {
		tweet := new(Tweet)
		err := rows.Scan(&tweet.ID, &tweet.Text, pq.Array(&tweet.Hashtags), pq.Array(&tweet.Symbols), pq.Array(&tweet.UserMentions), pq.Array(&tweet.Urls), &tweet.ScreenName, &tweet.FavoriteCount, &tweet.RetweetCount, &tweet.Posted)
		if err != nil {
			return nil, err
		}

		tweets = append(tweets, tweet)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tweets, nil
}

// GetTweetsByScreenName returns an array of tweets from the database filtered by screenname.
func GetTweetsByScreenName(ctx context.Context, screenName string, limitIn *int, offsetIn *int) ([]*Tweet, error) {
	limit := 10
	if limitIn != nil {
		limit = *limitIn
	}

	offset := 0
	if offsetIn != nil {
		offset = *offsetIn
	}

	rows, err := db.QueryContext(ctx, "SELECT id, text, hashtags, symbols, user_mentions, urls, screen_name, favorites, retweets, posted FROM tweets WHERE screen_name = $3 ORDER BY posted DESC LIMIT $1 OFFSET $2", limit, offset, screenName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tweets := make([]*Tweet, 0)
	for rows.Next() {
		tweet := new(Tweet)
		err := rows.Scan(&tweet.ID, &tweet.Text, pq.Array(&tweet.Hashtags), pq.Array(&tweet.Symbols), pq.Array(&tweet.UserMentions), pq.Array(&tweet.Urls), &tweet.ScreenName, &tweet.FavoriteCount, &tweet.RetweetCount, &tweet.Posted)
		if err != nil {
			return nil, err
		}

		tweets = append(tweets, tweet)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tweets, nil
}
