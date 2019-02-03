//go:generate go run ./scripts/gqlgen.go

package graphql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/icco/cacophony/models"
)

type key int8

const (
	userCtxKey key = 0
)

// ForContext finds the user from the context. This is usually inserted by
// WithUser.
func ForContext(ctx context.Context) *User {
	raw, _ := ctx.Value(userCtxKey).(*User)
	return raw
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
		u := ForContext(ctx)
		if u == nil || Role(u.Role) != role {
			// block calling the next resolver
			return nil, fmt.Errorf("forbidden")
		}

		// or let it pass through
		return next(ctx)
	}

	c.Directives.LoggedIn = func(ctx context.Context, _ interface{}, next graphql.Resolver) (interface{}, error) {
		u := ForContext(ctx)
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

func (r *mutationResolver) CreatePost(ctx context.Context, input NewPost) (Post, error) {
	p := &Post{}
	maxID, err := GetMaxID(ctx)
	if err != nil {
		return Post{}, err
	}
	id := maxID + 1

	p.ID = strconv.FormatInt(id, 10)

	if input.Title != nil {
		p.Title = *input.Title
	}

	if input.Content != nil {
		p.Content = *input.Content
	}

	if input.Datetime != nil {
		p.Datetime = *input.Datetime
	} else {
		p.Datetime = time.Now()
	}

	if input.Draft != nil {
		p.Draft = *input.Draft
	} else {
		p.Draft = true
	}

	p.Created = time.Now()

	err = p.Save(ctx)
	if err != nil {
		return Post{}, err
	}

	post, err := GetPost(ctx, id)
	if err != nil {
		return Post{}, err
	}

	return *post, nil
}

func (r *mutationResolver) UpsertBook(ctx context.Context, input EditBook) (Book, error) {
	b := &Book{}

	if input.ID != nil {
		b.ID = *input.ID
	}

	if input.Title != nil {
		b.Title = *input.Title
	}

	b.GoodreadsID = input.GoodreadsID

	err := b.Save(ctx)
	return *b, err
}

func (r *mutationResolver) EditPost(ctx context.Context, id string, input EditedPost) (Post, error) {
	p := &Post{}
	p.ID = id
	p.Title = input.Title
	p.Content = input.Content
	p.Datetime = input.Datetime
	p.Draft = input.Draft

	err := p.Save(ctx)
	if err != nil {
		return Post{}, err
	}

	i, err := strconv.ParseInt(p.ID, 10, 64)
	if err != nil {
		return *p, err
	}

	post, err := GetPost(ctx, i)
	if err != nil {
		return Post{}, err
	}

	return *post, nil
}

func (r *mutationResolver) UpsertLink(ctx context.Context, input NewLink) (Link, error) {
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
		return Link{}, err
	}

	link, err := GetLinkByURI(ctx, l.URI)
	if err != nil {
		return Link{}, err
	}

	return *link, nil
}

func (r *mutationResolver) UpsertStat(ctx context.Context, input NewStat) (Stat, error) {
	return Stat{}, fmt.Errorf("not implemented")
}

func (r *mutationResolver) InsertLog(ctx context.Context, input NewLog) (*Log, error) {
	l := &Log{}
	l.Code = input.Code
	l.Project = input.Project

	if input.Description != nil {
		l.Description = *input.Description
	}

	if input.Location != nil {
		l.Location = &Geo{
			Lat:  input.Location.Lat,
			Long: input.Location.Long,
		}
	}

	err := l.Save(ctx)
	return l, err
}

func (r *mutationResolver) UpsertTweet(ctx context.Context, input NewTweet) (Tweet, error) {
	t := &Tweet{
		FavoriteCount: input.FavoriteCount,
		Hashtags:      input.Hashtags,
		ID:            input.ID,
		Posted:        input.Posted,
		RetweetCount:  input.RetweetCount,
		Symbols:       input.Symbols,
		Text:          input.Text,
		Urls:          input.Urls,
		ScreenName:    input.ScreenName,
		UserMentions:  input.UserMentions,
	}

	err := t.Save(ctx)
	if err != nil {
		return Tweet{}, err
	}

	return *t, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Drafts(ctx context.Context, limit *int, offset *int) ([]*Post, error) {
	return Drafts(ctx)
}

func (r *queryResolver) Posts(ctx context.Context, limit *int, offset *int) ([]*Post, error) {
	return Posts(ctx, limit, offset)
}

func (r *queryResolver) Post(ctx context.Context, id string) (*Post, error) {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	return GetPost(ctx, i)
}

func (r *queryResolver) NextPost(ctx context.Context, id string) (*Post, error) {
	var postID string
	row := db.QueryRowContext(ctx, "SELECT id FROM posts WHERE draft = false AND date > (SELECT date FROM posts WHERE id = $1) ORDER BY date ASC LIMIT 1", id)
	err := row.Scan(&postID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		i, err := strconv.ParseInt(postID, 10, 64)
		if err != nil {
			return nil, err
		}
		return GetPost(ctx, i)
	}
}

func (r *queryResolver) PrevPost(ctx context.Context, id string) (*Post, error) {
	var postID string
	row := db.QueryRowContext(ctx, "SELECT id FROM posts WHERE draft = false AND date < (SELECT date FROM posts WHERE id = $1) ORDER BY date DESC LIMIT 1", id)
	err := row.Scan(&postID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		i, err := strconv.ParseInt(postID, 10, 64)
		if err != nil {
			return nil, err
		}
		return GetPost(ctx, i)
	}
}

func (r *queryResolver) Links(ctx context.Context, limit *int, offset *int) ([]*Link, error) {
	return GetLinks(ctx, limit, offset)
}

func (r *queryResolver) Link(ctx context.Context, id *string, url *string) (*Link, error) {
	if id != nil && url != nil {
		return nil, fmt.Errorf("do not specify an ID and a URI in input")
	}

	if id != nil {
		return GetLinkByID(ctx, *id)
	}

	if url != nil {
		return GetLinkByURI(ctx, *url)
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
		"stats",
		"links",
		"posts",
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
	return ForContext(ctx), nil
}

func (r *queryResolver) Tweets(ctx context.Context, limit *int, offset *int) ([]*Tweet, error) {
	return GetTweets(ctx, limit, offset)
}

func (r *queryResolver) Tweet(ctx context.Context, id string) (*Tweet, error) {
	return GetTweet(ctx, id)
}

func (r *queryResolver) TweetsByScreenName(ctx context.Context, screenName string, limit *int, offset *int) ([]*Tweet, error) {
	return GetTweetsByScreenName(ctx, screenName, limit, offset)
}

func (r *queryResolver) HomeTimelineURLs(ctx context.Context, limitIn *int) ([]*models.SavedURL, error) {
	urls := []*models.SavedURL{}
	limit := 100
	if limitIn != nil {
		limit = *limitIn
	}

	url := fmt.Sprintf("https://cacophony.natwelch.com/?count=%d", limit)
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

type twitterURLResolver struct{ *Resolver }

func (r *twitterURLResolver) Tweets(ctx context.Context, obj *models.SavedURL) ([]*Tweet, error) {
	tweets := make([]*Tweet, len(obj.TweetIDs))
	for i, id := range obj.TweetIDs {
		t, _ := GetTweet(ctx, id)
		tweets[i] = t
	}

	return tweets, nil
}
