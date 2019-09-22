//go:generate go run github.com/99designs/gqlgen -v

package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/icco/cacophony/models"
)

type key int8

const (
	userCtxKey key = 0
)

// GetUserFromContext finds the user from the context. This is usually inserted
// by WithUser.
func GetUserFromContext(ctx context.Context) *User {
	u, ok := ctx.Value(userCtxKey).(*User)
	if !ok {
		return nil
	}

	return u
}

// ParseLimit turns a limit and applies defaults into a pair of ints.
func ParseLimit(lim *Limit, defaultLimit, defaultOffset int) (int, int) {
	limit := defaultLimit
	offset := defaultOffset

	if lim != nil {
		i := *lim
		if i.Limit != nil {
			limit = *i.Limit
		}

		if i.Offset != nil {
			offset = *i.Offset
		}
	}

	return limit, offset
}

// WithUser puts a user in the context.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// Resolver is the type that gqlgen expects to exist
type Resolver struct{}

// New returns a Config that has all of the proper settings for this graphql
// server.
func New() Config {
	c := Config{
		Resolvers: &Resolver{},
	}

	c.Directives.HasRole = func(ctx context.Context, _ interface{}, next graphql.Resolver, role Role) (interface{}, error) {
		u := GetUserFromContext(ctx)
		if u == nil || Role(u.Role) != role {
			// block calling the next resolver
			return nil, fmt.Errorf("forbidden")
		}

		// or let it pass through
		return next(ctx)
	}

	c.Directives.LoggedIn = func(ctx context.Context, _ interface{}, next graphql.Resolver) (interface{}, error) {
		u := GetUserFromContext(ctx)
		if u == nil {
			// block calling the next resolver
			return nil, fmt.Errorf("forbidden")
		}

		// or let it pass through
		return next(ctx)
	}

	return c
}

// Mutation returns the resolver for Mutations.
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query returns the resolver for Queries.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// TwitterURL is a resolver factory to wrap the external twitter url type.
func (r *Resolver) TwitterURL() TwitterURLResolver {
	return &twitterURLResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreatePost(ctx context.Context, input EditPost) (*Post, error) {
	return r.EditPost(ctx, input)
}

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

func (r *mutationResolver) EditPost(ctx context.Context, input EditPost) (*Post, error) {
	var err error
	p := &Post{}

	// We do this so the defaults in save don't overwrite stuff on upsert.
	if input.ID != nil {
		p, err = GetPostString(ctx, *input.ID)
		if err != nil {
			return nil, err
		}

		if p == nil {
			return nil, fmt.Errorf("cannot edit post that does not exist")
		}
	}

	if input.Title != nil {
		p.Title = *input.Title
	}

	if input.Content != nil {
		p.Content = *input.Content
	}

	if input.Datetime != nil {
		p.Datetime = *input.Datetime
	}

	if input.Draft != nil {
		p.Draft = *input.Draft
	} else {
		p.Draft = true
	}

	err = p.Save(ctx)
	if err != nil {
		return nil, err
	}

	return GetPostString(ctx, p.ID)
}

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

func (r *mutationResolver) UpsertStat(ctx context.Context, input NewStat) (*Stat, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *mutationResolver) InsertLog(ctx context.Context, input NewLog) (*Log, error) {
	l := &Log{}
	l.Code = input.Code
	l.Project = input.Project

	u := GetUserFromContext(ctx)
	if u != nil {
		l.User = *u
	}

	if input.Description != nil {
		l.Description = *input.Description
	}

	if input.Location != nil {
		l.Location = &Geo{
			Lat:  input.Location.Lat,
			Long: input.Location.Long,
		}
	}

	if input.Duration != nil {
		l.Duration = ParseDurationFromString(*input.Duration)
	}

	err := l.Save(ctx)
	return l, err
}

func (r *mutationResolver) UpsertPage(ctx context.Context, input EditPage) (*Page, error) {
	var err error
	p := &Page{}

	if input.ID != nil {
		p, err = GetPageByID(ctx, *input.ID)
		if err != nil {
			return nil, err
		}
	}

	p.Content = input.Content
	p.Title = input.Title

	u := GetUserFromContext(ctx)
	if u != nil {
		p.User = *u
	}

	if input.Slug != nil {
		p.Slug = *input.Slug
	}

	if input.Category != nil {
		p.Category = *input.Category
	}

	err = p.Save(ctx)
	if err != nil {
		return nil, err
	}

	return p, nil
}

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

func (r *mutationResolver) AddComment(ctx context.Context, input AddComment) (*Comment, error) {
	c := &Comment{}
	c.Content = input.Content
	c.User = GetUserFromContext(ctx)

	post, err := GetPostString(ctx, input.PostID)
	if err != nil {
		return nil, err
	}
	c.Post = post

	err = c.Save(ctx)
	if err != nil {
		return nil, err
	}

	return GetComment(ctx, c.ID)
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Drafts(ctx context.Context, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return Drafts(ctx, limit, offset)
}

func (r *queryResolver) FuturePosts(ctx context.Context, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return FuturePosts(ctx, limit, offset)
}

func (r *queryResolver) Posts(ctx context.Context, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return Posts(ctx, limit, offset)
}

func (r *queryResolver) Post(ctx context.Context, id string) (*Post, error) {
	return GetPostString(ctx, id)
}

func (r *queryResolver) NextPost(ctx context.Context, id string) (*Post, error) {
	p, err := GetPostString(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.Next(ctx)
}

func (r *queryResolver) PrevPost(ctx context.Context, id string) (*Post, error) {
	p, err := GetPostString(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.Prev(ctx)
}

func (r *queryResolver) Links(ctx context.Context, input *Limit) ([]*Link, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return GetLinks(ctx, limit, offset)
}

func (r *queryResolver) Books(ctx context.Context, input *Limit) ([]*Book, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return GetBooks(ctx, limit, offset)
}

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

func (r *queryResolver) Stats(ctx context.Context, count *int) ([]*Stat, error) {
	limit := 6
	if count != nil {
		limit = *count
		if limit <= 0 {
			limit = 6
		}
	}

	rows, err := db.QueryContext(ctx, "SELECT key, value FROM stats ORDER BY modified_at DESC LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]*Stat, 0)
	for rows.Next() {
		stat := new(Stat)
		err := rows.Scan(&stat.Key, &stat.Value)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *queryResolver) PostsByTag(ctx context.Context, tag string) ([]*Post, error) {
	return PostsByTag(ctx, tag)
}

func (r *queryResolver) Counts(ctx context.Context) ([]*Stat, error) {
	stats := make([]*Stat, 0)
	for _, table := range []string{
		"books",
		"links",
		"logs",
		"photos",
		"posts",
		"stats",
		"tweets",
	} {
		stat := new(Stat)
		stat.Key = table
		err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT count(*) FROM %s", table)).Scan(&stat.Value)
		if err != nil {
			return stats, err
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *queryResolver) Whoami(ctx context.Context) (*User, error) {
	return GetUserFromContext(ctx), nil
}

func (r *queryResolver) Tweets(ctx context.Context, input *Limit) ([]*Tweet, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return GetTweets(ctx, limit, offset)
}

func (r *queryResolver) Tweet(ctx context.Context, id string) (*Tweet, error) {
	return GetTweet(ctx, id)
}

func (r *queryResolver) TweetsByScreenName(ctx context.Context, screenName string, input *Limit) ([]*Tweet, error) {
	limit, offset := ParseLimit(input, 10, 0)
	return GetTweetsByScreenName(ctx, screenName, limit, offset)
}

func (r *queryResolver) HomeTimelineURLs(ctx context.Context, input *Limit) ([]*models.SavedURL, error) {
	urls := []*models.SavedURL{}
	limit, offset := ParseLimit(input, 100, 0)

	url := fmt.Sprintf("https://cacophony.natwelch.com/?count=%d&offset=%d", limit, offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return urls, err
	}

	req = req.WithContext(ctx)
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return urls, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return urls, err
	}

	err = json.Unmarshal(body, &urls)
	return urls, err
}

func (r *queryResolver) Tags(ctx context.Context) ([]string, error) {
	return AllTags(ctx)
}

func (r *queryResolver) Log(ctx context.Context, id string) (*Log, error) {
	return GetLog(ctx, id)
}

func (r *queryResolver) Logs(ctx context.Context, input *Limit) ([]*Log, error) {
	u := GetUserFromContext(ctx)
	limit, offset := ParseLimit(input, 25, 0)

	return UserLogs(ctx, u, limit, offset)
}

func (r *queryResolver) Photos(ctx context.Context, input *Limit) ([]*Photo, error) {
	u := GetUserFromContext(ctx)
	limit, offset := ParseLimit(input, 25, 0)

	return UserPhotos(ctx, u, limit, offset)
}

func (r *queryResolver) Time(ctx context.Context) (*time.Time, error) {
	now := time.Now()
	return &now, nil
}

func (r *queryResolver) GetPageByID(ctx context.Context, id string) (*Page, error) {
	return GetPageByID(ctx, id)
}

func (r *queryResolver) GetPageBySlug(ctx context.Context, slug string) (*Page, error) {
	return GetPageBySlug(ctx, slug)
}

func (r *queryResolver) GetPages(ctx context.Context) ([]*Page, error) {
	return GetPages(ctx)
}

type twitterURLResolver struct{ *Resolver }

func (r *twitterURLResolver) Link(ctx context.Context, obj *models.SavedURL) (*URI, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *twitterURLResolver) Tweets(ctx context.Context, obj *models.SavedURL) ([]*Tweet, error) {
	tweets := make([]*Tweet, len(obj.TweetIDs))
	for i, id := range obj.TweetIDs {
		t, _ := GetTweet(ctx, id)
		tweets[i] = t
	}

	return tweets, nil
}
