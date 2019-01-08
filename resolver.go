//go:generate go run ./scripts/gqlgen.go

package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

type key int

const (
	// UserCtxKey is a constant context key
	UserCtxKey = iota
)

// ForContext finds the user from the context. Requires
// server.ContextMiddleware to have run.
func ForContext(ctx context.Context) *User {
	raw, _ := ctx.Value(UserCtxKey).(*User)
	return raw
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
			return nil, fmt.Errorf("Forbidden")
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

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreatePost(ctx context.Context, input NewPost) (Post, error) {
	p := &Post{}
	maxID, err := GetMaxID(ctx)
	if err != nil {
		return Post{}, err
	}
	id := maxID + 1

	p.ID = strconv.FormatInt(id, 10)
	p.Title = input.Title
	p.Content = input.Content
	p.Datetime = input.Datetime
	p.Draft = input.Draft
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
		return nil, fmt.Errorf("Error running get query: %+v", err)
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
		return nil, fmt.Errorf("Error running get query: %+v", err)
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
		return nil, fmt.Errorf("Please don't specify an ID and a URI in input.")
	}

	if id != nil {
		return GetLinkByID(ctx, *id)
	}

	if url != nil {
		return GetLinkByURI(ctx, *url)
	}

	return nil, fmt.Errorf("Not valid input.")
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
