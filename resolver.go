//go:generate gorunpkg github.com/99designs/gqlgen

package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Resolver struct{}

func New() Config {
	c := Config{
		Resolvers: &Resolver{},
	}
	return c
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func GetMaxId() (int64, error) {
	row := db.QueryRow("SELECT MAX(id) from posts")
	var id int64
	if err := row.Scan(&id); err != nil {
		return -1, err
	}

	return id, nil
}

func (r *mutationResolver) CreatePost(ctx context.Context, input NewPost) (Post, error) {

	maxId, err := GetMaxId()
	if err != nil {
		return Post{}, err
	}
	id := maxId + 1

	_, err = db.Exec("INSERT INTO posts(id, title, content, date, draft, created_at, modified_at) VALUES ($1, $2, $3, $4, $5, $6, $6)",
		id,
		input.Title,
		input.Content,
		input.Datetime,
		input.Draft,
		time.Now(),
	)
	if err != nil {
		return Post{}, err
	}

	post, err := GetPost(id)
	if err != nil {
		return Post{}, err
	}

	return *post, nil
}

func (r *mutationResolver) EditPost(ctx context.Context, id string, input NewPost) (Post, error) {
	panic("not implemented")
}

func (r *mutationResolver) CreateLink(ctx context.Context, input NewLink) (Link, error) {
	panic("not implemented")
}

func (r *mutationResolver) UpsertStat(ctx context.Context, input NewStat) (Stat, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) AllPosts(ctx context.Context) ([]*Post, error) {
	rows, err := db.Query("SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE draft = false ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*Post, 0)
	for rows.Next() {
		post := new(Post)
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Datetime, &post.Created, &post.Modified, pq.Array(&post.Tags), &post.Draft)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *queryResolver) Posts(ctx context.Context, limit *int, offset *int) ([]*Post, error) {
	rows, err := db.Query("SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE draft = false ORDER BY date DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*Post, 0)
	for rows.Next() {
		post := new(Post)
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Datetime, &post.Created, &post.Modified, pq.Array(&post.Tags), &post.Draft)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *queryResolver) Post(ctx context.Context, id string) (*Post, error) {
	var post Post
	row := db.QueryRow("SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE id = $1", id)
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

func (r *queryResolver) NextPost(ctx context.Context, id string) (*string, error) {
	var postId string
	row := db.QueryRow("SELECT id FROM posts WHERE id = $1", id)
	err := row.Scan(&postId)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("No post with id %s", id)
	case err != nil:
		return nil, fmt.Errorf("Error running get query: %+v", err)
	default:
		return &postId, nil
	}
}

func (r *queryResolver) PrevPost(ctx context.Context, id string) (*string, error) {
	var postId string
	row := db.QueryRow("SELECT id FROM posts WHERE id = $1", id)
	err := row.Scan(&postId)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("No post with id %s", id)
	case err != nil:
		return nil, fmt.Errorf("Error running get query: %+v", err)
	default:
		return &postId, nil
	}
}

func (r *queryResolver) Stats(ctx context.Context, count *int) ([]*Stat, error) {
	limit := 6
	if count != nil {
		limit = *count
		if limit <= 0 {
			limit = 6
		}
	}

	rows, err := db.Query("SELECT key, value FROM stats ORDER BY modified_at DESC LIMIT $1", limit)
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

func (r *queryResolver) AllLinks(ctx context.Context) ([]*Link, error) {
	panic("not implemented")
}
func (r *queryResolver) Links(ctx context.Context, limit *int, offset *int) ([]*Link, error) {
	panic("not implemented")
}
func (r *queryResolver) Link(ctx context.Context, id string) (*Link, error) {
	panic("not implemented")
}
