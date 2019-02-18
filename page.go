package graphql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/lib/pq"
)

// Page is a wiki page.
type Page struct {
	ID       string    `json:"id"`
	Slug     string    `json:"slug"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Category string    `json:"category"`
	Tags     []string  `json:"tags"`
	User     User      `json:"user"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// Save inserts or updates a page into the database.
func (p *Page) Save(ctx context.Context) error {
	if p.ID == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		p.ID = uuid.String()
	}

	if p.Slug == "" {
		p.Slug = Slugify(p.Title)
	}

	tags, err := ParseTags(p.Content)
	if err != nil {
		return err
	}
	p.Tags = tags

	if p.Created.IsZero() {
		p.Created = time.Now()
	}

	p.Modified = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO pages(id, slug, title, content, category, tags, user_id, created_at, modified_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE
SET (slug, title, content, category, tags, user_id, created_at, modified_at) = ($2, $3, $4, $5, $6, $7, $8, $9)
WHERE pages.id = $1;
`,
		p.ID,
		p.Slug,
		p.Title,
		p.Content,
		p.Category,
		pq.Array(p.Tags),
		p.User.ID,
		p.Created,
		p.Modified); err != nil {
		return err
	}

	return nil
}

// Slugify returns a dash seperated string that doesn't have unicode chars.
func Slugify(title string) string {
	return slug.Make(title)
}
