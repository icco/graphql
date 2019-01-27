package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Post is our representation of a post in the database.
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

// GetMaxID returns the greatest post ID in the database.
func GetMaxID(ctx context.Context) (int64, error) {
	row := db.QueryRowContext(ctx, "SELECT MAX(id) from posts")
	var id int64
	if err := row.Scan(&id); err != nil {
		return -1, err
	}

	return id, nil
}

// GetPost gets a post by ID from the database.
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

// AllPosts returns all posts from the database.
func AllPosts(ctx context.Context, isDraft bool) ([]*Post, error) {
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

func AllTags(ctx context.Context) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT UNNEST(tags) AS tag, COUNT(*) AS cnt FROM posts GROUP BY tag ORDER BY cnt DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]string, 0)
	for rows.Next() {
		var tag string
		var cnt int
		err := rows.Scan(&tag, &cnt)
		if err != nil {
			return tags, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return tags, err
	}

	return tags, nil
}

// Drafts is a simple wrapper around Posts that does return drafts.
func Drafts(ctx context.Context) ([]*Post, error) {
	return AllPosts(ctx, true)
}

var tagAliases = map[string]string{
	"hackerschool": "recursecenter",
}

// ParseTags returns a list of all hashtags currently in a post.
func ParseTags(text string) ([]string, error) {
	// http://golang.org/pkg/regexp/#Regexp.FindAllStringSubmatch
	finds := HashtagRegex.FindAllStringSubmatch(text, -1)
	tagMap := map[string]int{}
	for _, v := range finds {
		if len(v) > 2 {
			tag := strings.ToLower(v[2])

			if alias, ok := tagAliases[tag]; ok {
				tagMap[alias]++
			}
			tagMap[tag]++
		}
	}

	ret := []string{}
	for k := range tagMap {
		ret = append(ret, k)
	}

	sort.Strings(ret)

	return ret, nil
}

// Save inserts or updates a post into the database.
func (p *Post) Save(ctx context.Context) error {
	if p.ID == "" {
		maxID, err := GetMaxID(ctx)
		if err != nil {
			return err
		}

		p.ID = fmt.Sprintf("%d", maxID+1)
	}

	tags, err := ParseTags(p.Content)
	if err != nil {
		return err
	}
	p.Tags = tags

	if p.Title == "" {
		p.Title = fmt.Sprintf("Untitled #%s", p.ID)
	}

	if p.Created.IsZero() {
		p.Created = time.Now()
	}

	p.Modified = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO posts(id, title, content, date, draft, created_at, modified_at, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE
SET (title, content, date, draft, created_at, modified_at, tags) = ($2, $3, $4, $5, $6, $7, $8)
WHERE posts.id = $1;
`,
		p.ID,
		p.Title,
		p.Content,
		p.Datetime,
		p.Draft,
		p.Created,
		p.Modified,
		pq.Array(p.Tags)); err != nil {
		return err
	}

	_, err = strconv.ParseInt(p.ID, 10, 64)
	if err != nil {
		return err
	}

	return nil
}

// Summary returns the first sentence of a post.
func (p *Post) Summary() string {
	return SummarizeText(p.Content)
}

// HTML returns the post as rendered HTML.
func (p *Post) HTML() template.HTML {
	return Markdown(p.Content)
}

// ReadTime calculates the number of seconds it should take to read the post.
func (p *Post) ReadTime() int32 {
	ReadingSpeed := 265.0
	words := len(strings.Split(p.Content, " "))
	seconds := int32(math.Ceil(float64(words) / ReadingSpeed * 60.0))

	return seconds
}

// Posts returns some posts.
func Posts(ctx context.Context, limit *int, offset *int) ([]*Post, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE draft = false ORDER BY date DESC LIMIT $1 OFFSET $2", limit, offset)
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

// PostsByTag returns all posts with a tag.
func PostsByTag(ctx context.Context, tag string) ([]*Post, error) {
	query := "SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE $1 = ANY(tags) and draft = false ORDER BY date DESC"
	rows, err := db.QueryContext(ctx, query, tag)
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
