package models

import (
	"database/sql"
	"fmt"
	"html/template"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/russross/blackfriday"
)

type Post struct {
	Id       int64     `json:"id"`
	Title    string    `json:"title"` // optional
	Content  string    `json:"text"`  // Markdown
	Datetime time.Time `json:"date"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	Tags     []string  `json:"tags"`
	Longform string    `json:"-"`
	Draft    bool      `json:"-"`
}

func NewPost(title string, content string, datetime time.Time, tags []string) *Post {
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

func GetPost(id int64) (*Post, error) {
	var post Post
	row := db.QueryRow("SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE id = $1", id)
	err := row.Scan(&post.Id, &post.Title, &post.Content, &post.Datetime, &post.Created, &post.Modified, pq.Array(&post.Tags), &post.Draft)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("No post with id %d", id)
	case err != nil:
		return nil, fmt.Errorf("Error running get query: %+v", err)
	default:
		return &post, nil
	}
}

func Posts(isDraft bool) ([]*Post, error) {
	rows, err := db.Query("SELECT id, title, content, date, created_at, modified_at, tags, draft FROM posts WHERE draft = $1 ORDER BY date DESC", isDraft)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*Post, 0)
	for rows.Next() {
		post := new(Post)
		err := rows.Scan(&post.Id, &post.Title, &post.Content, &post.Datetime, &post.Created, &post.Modified, pq.Array(&post.Tags), &post.Draft)
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

func AllPosts() ([]*Post, error) {
	return Posts(false)
}

func Drafts() ([]*Post, error) {
	return Posts(true)
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

func (e *Post) Save() error {

	return nil
}

var HashtagRegex *regexp.Regexp = regexp.MustCompile(`(\s)#(\w+)`)
var TwitterHandleRegex *regexp.Regexp = regexp.MustCompile(`(\s)@([_A-Za-z0-9]+)`)

// Markdown generator.
func Markdown(str string) template.HTML {
	inc := []byte(str)
	inc = twitterHandleToMarkdown(inc)
	inc = hashTagsToMarkdown(inc)
	s := blackfriday.MarkdownCommon(inc)
	return template.HTML(s)
}

// Takes a chunk of markdown and just returns the first paragraph.
func SummarizeText(str string) string {
	out := strings.Split(str, "\n")
	return out[0]
}

func twitterHandleToMarkdown(in []byte) []byte {
	return TwitterHandleRegex.ReplaceAll(in, []byte("$1[@$2](http://twitter.com/$2)"))
}

func hashTagsToMarkdown(in []byte) []byte {
	return HashtagRegex.ReplaceAll(in, []byte("$1[#$2](/tags/$2)"))
}

type MarkdownHandlerData struct {
	Text template.HTML
}

func (e *Post) Html() template.HTML {
	return Markdown(e.Content)
}

func (e *Post) ReadTime() int {
	ReadingSpeed := 265.0
	words := len(strings.Split(e.Content, " "))
	seconds := int(math.Ceil(float64(words) / ReadingSpeed * 60.0))

	return seconds
}

func (e *Post) Summary() string {
	return ""
}
