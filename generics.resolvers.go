package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// UpsertBook is the resolver for the upsertBook field.
func (r *mutationResolver) UpsertBook(ctx context.Context, input EditBook) (*Book, error) {
	b := &Book{}

	if input.ID != nil {
		b.ID = *input.ID
	}

	if input.Title != nil {
		b.Title = *input.Title
	}

	b.GoodreadsID = input.GoodreadsID

	err := b.Save(ctx)
	return b, err
}

// UpsertLink is the resolver for the upsertLink field.
func (r *mutationResolver) UpsertLink(ctx context.Context, input NewLink) (*Link, error) {
	l := &Link{}
	l.Title = input.Title
	l.Description = input.Description
	l.URI = input.URI
	l.Tags = input.Tags

	if input.Created != nil {
		l.Created = *input.Created
	} else {
		now := time.Now()
		input.Created = &now
	}

	err := l.Save(ctx)
	if err != nil {
		return nil, err
	}

	return GetLinkByURI(ctx, l.URI.String())
}

// UpsertStat is the resolver for the upsertStat field.
func (r *mutationResolver) UpsertStat(ctx context.Context, input NewStat) (*Stat, error) {
	s := &Stat{
		Key:   input.Key,
		Value: input.Value,
	}

	if err := s.Save(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

// UpsertTweet is the resolver for the upsertTweet field.
func (r *mutationResolver) UpsertTweet(ctx context.Context, input NewTweet) (*Tweet, error) {
	t := &Tweet{
		FavoriteCount: input.FavoriteCount,
		Hashtags:      input.Hashtags,
		ID:            input.ID,
		Posted:        input.Posted,
		RetweetCount:  input.RetweetCount,
		Symbols:       input.Symbols,
		Text:          input.Text,
		ScreenName:    input.ScreenName,
		UserMentions:  input.UserMentions,
		Urls:          input.Urls,
	}

	err := t.Save(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Books is the resolver for the books field.
func (r *queryResolver) Books(ctx context.Context, input *Limit) ([]*Book, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return GetBooks(ctx, limit, offset)
}

// Links is the resolver for the links field.
func (r *queryResolver) Links(ctx context.Context, input *Limit) ([]*Link, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return GetLinks(ctx, limit, offset)
}

// Link is the resolver for the link field.
func (r *queryResolver) Link(ctx context.Context, id *string, url *URI) (*Link, error) {
	if id != nil && url != nil {
		return nil, fmt.Errorf("do not specify an ID and a URI in input")
	}

	if id != nil {
		return GetLinkByID(ctx, *id)
	}

	if url != nil {
		return GetLinkByURI(ctx, url.String())
	}

	return nil, fmt.Errorf("not valid input")
}

// Stats is the resolver for the stats field.
func (r *queryResolver) Stats(ctx context.Context, count *int) ([]*Stat, error) {
	limit := 6
	if count != nil {
		limit = *count
		if limit <= 0 {
			limit = 6
		}
	}

	return GetStats(ctx, limit)
}

// Stat is the resolver for the stat field.
func (r *queryResolver) Stat(ctx context.Context, key string, input *Limit) ([]*Stat, error) {
	limit, offset := ParseLimit(input, 10, 0)
	return GetStat(ctx, key, limit, offset)
}

// Counts is the resolver for the counts field.
func (r *queryResolver) Counts(ctx context.Context) ([]*Stat, error) {
	var stats []*Stat
	for _, table := range []string{
		"books",
		"links",
		"logs",
		"photos",
		"posts",
	} {
		stat := new(Stat)
		stat.Key = table
		if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT count(*) FROM %s", table)).Scan(&stat.Value); err != nil {
			return stats, err
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// Whoami is the resolver for the whoami field.
func (r *queryResolver) Whoami(ctx context.Context) (*User, error) {
	return GetUserFromContext(ctx), nil
}

// Tweets is the resolver for the tweets field.
func (r *queryResolver) Tweets(ctx context.Context, input *Limit) ([]*Tweet, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return GetTweets(ctx, limit, offset)
}

// Tweet is the resolver for the tweet field.
func (r *queryResolver) Tweet(ctx context.Context, id string) (*Tweet, error) {
	return GetTweet(ctx, id)
}

// TweetsByScreenName is the resolver for the tweetsByScreenName field.
func (r *queryResolver) TweetsByScreenName(ctx context.Context, screenName string, input *Limit) ([]*Tweet, error) {
	limit, offset := ParseLimit(input, 10, 0)
	return GetTweetsByScreenName(ctx, screenName, limit, offset)
}

// HomeTimelineURLs is the resolver for the homeTimelineURLs field.
func (r *queryResolver) HomeTimelineURLs(ctx context.Context, input *Limit) ([]*TwitterURL, error) {
	limit, offset := ParseLimit(input, 100, 0)
	url := fmt.Sprintf("https://cacophony.natwelch.com/?count=%d&offset=%d", limit, offset)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get from cacophony: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read cacophony body: %w", err)
	}

	var urls []*TwitterURL
	if err := json.Unmarshal(body, &urls); err != nil {
		return nil, err
	}

	return urls, nil
}

// Time is the resolver for the time field.
func (r *queryResolver) Time(ctx context.Context) (*time.Time, error) {
	now := time.Now()
	return &now, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
