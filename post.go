package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"math"
	"regexp"
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

// GetPostString gets a post by an ID string.
func GetPostString(ctx context.Context, id string) (*Post, error) {
	match, err := regexp.MatchString("^[0-9]+$", id)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, fmt.Errorf("no post with id %s", id)
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	return GetPost(ctx, i)
}

// GetPost gets a post by ID from the database.
func GetPost(ctx context.Context, id int64) (*Post, error) {
	var post Post
	row := db.QueryRowContext(ctx, "SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE id = $1", id)
	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Datetime, &post.Created, &post.Modified, pq.Array(&post.Tags), &post.Draft)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("error with get: %w", err)
	default:
		return &post, nil
	}
}

// AllTags returns all tags used in all posts.
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
func Drafts(ctx context.Context, limit, offset int) ([]*Post, error) {
	query := `
SELECT id, title, content, date, created_at, modified_at, tags, draft
FROM posts
WHERE draft = true
ORDER BY date DESC
LIMIT $1 OFFSET $2
`
	return postQuery(ctx, query, limit, offset)
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

	if p.Datetime.IsZero() {
		p.Datetime = time.Now()
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

// Comments returns the comments for a post
func (p *Post) Comments(ctx context.Context, input *Limit) ([]*Comment, error) {
	limit := 100
	offset := 0
	if input != nil {
		i := *input
		if i.Limit != nil {
			limit = *i.Limit
		}

		if i.Offset != nil {
			offset = *i.Offset
		}
	}

	return PostComments(ctx, p.ID, limit, offset)
}

// IntID returns this posts ID as an int.
func (p *Post) IntID() int64 {
	i, err := strconv.ParseInt(p.ID, 10, 64)
	if err != nil {
		return 0
	}

	return i
}

// Summary returns the first sentence of a post.
func (p *Post) Summary() string {
	return SummarizeText(p.Content)
}

// HTML returns the post as rendered HTML.
func (p *Post) HTML() template.HTML {
	return Markdown(p.Content)
}

// URI returns an absolute link to this post.
func (p *Post) URI() *URI {
	return NewURI(fmt.Sprintf("https://writing.natwelch.com/post/%s", p.ID))
}

// Next returns the next post chronologically.
func (p *Post) Next(ctx context.Context) (*Post, error) {
	var postID string
	row := db.QueryRowContext(ctx, "SELECT id FROM posts WHERE draft = false AND date > (SELECT date FROM posts WHERE id = $1) ORDER BY date ASC LIMIT 1", p.ID)
	err := row.Scan(&postID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return GetPostString(ctx, postID)
	}
}

// Prev returns the previous post chronologically.
func (p *Post) Prev(ctx context.Context) (*Post, error) {
	var postID string
	row := db.QueryRowContext(ctx, "SELECT id FROM posts WHERE draft = false AND date < (SELECT date FROM posts WHERE id = $1) ORDER BY date DESC LIMIT 1", p.ID)
	err := row.Scan(&postID)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return GetPostString(ctx, postID)
	}
}

// ReadTime calculates the number of seconds it should take to read the post.
func (p *Post) Readtime(_ context.Context) int32 {
	ReadingSpeed := 265.0
	words := len(strings.Split(p.Content, " "))
	seconds := int32(math.Ceil(float64(words) / ReadingSpeed * 60.0))

	return seconds
}

// IsLinkable exists to show that this method implements the Linkable type in
// graphql.
func (p *Post) IsLinkable() {}

// Related returns an array of related posts. It is quite slow in comparison to
// other queries.
func (p *Post) Related(ctx context.Context, input *Limit) ([]*Post, error) {
	limit := 3
	offset := 0
	if input != nil {
		i := *input
		if i.Limit != nil {
			limit = *i.Limit
		}

		if i.Offset != nil {
			offset = *i.Offset
		}
	}

	_, err := db.QueryContext(ctx, "SELECT set_limit(0.6)")
	if err != nil {
		return nil, err
	}

	// From https://www.postgresql.org/docs/9.6/pgtrgm.html
	query := `
  SELECT SIMILARITY($1, title) AS sim, id
  FROM posts
  WHERE id != $2
    AND title % $1
    AND draft = false
  ORDER BY sim DESC
  LIMIT $3 OFFSET $4`

	rows, err := db.QueryContext(ctx, query, p.Title, p.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*Post, 0)
	for rows.Next() {
		var id string
		var sim float64
		err := rows.Scan(&sim, &id)
		if err != nil {
			return nil, err
		}
		post, err := GetPostString(ctx, id)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if len(posts) < limit {
		existing := []int64{}
		for _, p := range posts {
			existing = append(existing, p.IntID())
		}

		addPosts, err := GetRandomPosts(ctx, limit-len(posts), existing)
		if err != nil {
			return nil, err
		}

		posts = append(posts, addPosts...)
	}

	return posts, nil
}

// GetRandomPosts returns a random selection of posts.
func GetRandomPosts(ctx context.Context, limit int, notIn []int64) ([]*Post, error) {
	query := `SELECT id, title, content, date, created_at, modified_at, tags, draft
  FROM posts
  WHERE draft = false
    AND id <> ALL($1)
  ORDER BY random() DESC LIMIT $2`

	return postQuery(ctx, query, pq.Array(notIn), limit)
}

// Posts returns some posts.
func Posts(ctx context.Context, limit int, offset int) ([]*Post, error) {
	query := `
SELECT id, title, content, date, created_at, modified_at, tags, draft
FROM posts
WHERE draft = false
  AND date <= NOW()
ORDER BY date DESC
LIMIT $1 OFFSET $2
`
	return postQuery(ctx, query, limit, offset)
}

// FuturePosts returns some posts that are in the future.
func FuturePosts(ctx context.Context, limit int, offset int) ([]*Post, error) {
	query := `
SELECT id, title, content, date, created_at, modified_at, tags, draft
FROM posts
WHERE draft = false
  AND date > NOW()
ORDER BY date DESC
LIMIT $1 OFFSET $2
`
	return postQuery(ctx, query, limit, offset)
}

// PostsByTag returns all posts with a tag.
func PostsByTag(ctx context.Context, tag string) ([]*Post, error) {
	query := `
SELECT id, title, content, date, created_at, modified_at, tags, draft
FROM posts
WHERE $1 = ANY(tags)
  AND draft = false
ORDER BY date DESC
`

	return postQuery(ctx, query, tag)
}

func postQuery(ctx context.Context, query string, args ...interface{}) ([]*Post, error) {
	rows, err := db.QueryContext(ctx, query, args...)
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
