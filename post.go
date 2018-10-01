package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Readtime int       `json:"readtime"`
	Datetime time.Time `json:"datetime"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	Draft    bool      `json:"draft"`
	Tags     []string  `json:"tags"`
	Links    []*Link   `json:"links"`
}

func GeneratePost(title string, content string, datetime time.Time, tags []string) *Post {
	e := new(Post)

	// User supplied content
	e.Title = title
	e.Content = content
	e.Datetime = datetime
	e.Tags = tags

	// Computer generated content
	e.Created = time.Now()
	e.Modified = time.Now()
	e.Draft = false

	return e
}

func GetMaxId(ctx context.Context) (int64, error) {
	row := db.QueryRowContext(ctx, "SELECT MAX(id) from posts")
	var id int64
	if err := row.Scan(&id); err != nil {
		return -1, err
	}

	return id, nil
}

func CreatePost(ctx context.Context, input Post) (Post, error) {
	maxId, err := GetMaxId(ctx)
	if err != nil {
		return Post{}, err
	}
	id := maxId + 1

	_, err = db.ExecContext(ctx, "INSERT INTO posts(id, title, content, date, draft, created_at, modified_at) VALUES ($1, $2, $3, $4, $5, $6, $6)",
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

	post, err := GetPost(ctx, id)
	if err != nil {
		return Post{}, err
	}

	return *post, nil
}

func GetPost(ctx context.Context, id int64) (*Post, error) {
	var post Post
	row := db.QueryRowContext(ctx, "SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE id = $1", id)
	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Datetime, &post.Created, &post.Modified, pq.Array(&post.Tags), &post.Draft)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("No post with id %d", id)
	case err != nil:
		return nil, fmt.Errorf("Error running get query: %+v", err)
	default:
		return &post, nil
	}
}

func Posts(ctx context.Context, isDraft bool) ([]*Post, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE draft = $1 ORDER BY date DESC", isDraft)
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

func AllPosts(ctx context.Context) ([]*Post, error) {
	return Posts(ctx, false)
}

func Drafts(ctx context.Context) ([]*Post, error) {
	return Posts(ctx, true)
}

func ParseTags(text string) ([]string, error) {
	// http://golang.org/pkg/regexp/#Regexp.FindAllStringSubmatch
	finds := HashtagRegex.FindAllStringSubmatch(text, -1)
	ret := make([]string, 0)
	for _, v := range finds {
		if len(v) > 2 {
			ret = append(ret, strings.ToLower(v[2]))
		}
	}

	return ret, nil
}

func (p *Post) Save(ctx context.Context) error {
	_, err := db.ExecContext(ctx, "INSERT INTO posts(id, title, content, date, draft, created_at, modified_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		p.ID,
		p.Title,
		p.Content,
		p.Datetime,
		p.Draft,
		p.Created,
		time.Now())

	return err
}

func (p *Post) Summary() string {
	return SummarizeText(p.Content)
}

func (e *Post) Html() template.HTML {
	return Markdown(e.Content)
}

func (e *Post) ReadTime() int32 {
	ReadingSpeed := 265.0
	words := len(strings.Split(e.Content, " "))
	seconds := int32(math.Ceil(float64(words) / ReadingSpeed * 60.0))

	return seconds
}
