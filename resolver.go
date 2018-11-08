//go:generate gorunpkg github.com/99designs/gqlgen

package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/lib/pq"
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

	return *l, nil
}

func (r *mutationResolver) UpsertStat(ctx context.Context, input NewStat) (Stat, error) {
	return Stat{}, fmt.Errorf("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Posts(ctx context.Context, limit *int, offset *int) ([]*Post, error) {
	return Posts(ctx, limit, offset)
}

func (r *queryResolver) PostsByTag(ctx context.Context, tag string) ([]*Post, error) {
	return PostsByTag(ctx, tag)
}

func (r *queryResolver) Post(ctx context.Context, id string) (*Post, error) {
	var post Post
	row := db.QueryRowContext(ctx, "SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE id = $1", id)
	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Datetime, &post.Created, &post.Modified, pq.Array(&post.Tags), &post.Draft)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("No post with id %s", id)
	case err != nil:
		return nil, fmt.Errorf("Error running get query: %+v", err)
	default:
		return &post, nil
	}
}

func (r *queryResolver) NextPost(ctx context.Context, id string) (*Post, error) {
	var postID string
	row := db.QueryRowContext(ctx, "SELECT id FROM posts WHERE draft = false AND date > (SELECT date FROM posts WHERE id = $1) ORDER BY date ASC LIMIT 1", id)
	err := row.Scan(&postID)
	switch {
	case err == sql.ErrNoRows:
		return nil, sql.ErrNoRows
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
		return nil, sql.ErrNoRows
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

func (r *queryResolver) Drafts(ctx context.Context, limit *int, offset *int) ([]*Post, error) {
	panic("not implemented")
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

func (r *queryResolver) Links(ctx context.Context, limit *int, offset *int) ([]*Link, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *queryResolver) Link(ctx context.Context, id string) (*Link, error) {
	return nil, fmt.Errorf("not implemented")
}
