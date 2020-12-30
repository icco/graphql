package graphql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/icco/graphql/time/hexdate"
	"github.com/icco/graphql/time/neralie"
)

// Page is a wiki page.
type Page struct {
	Slug     string      `json:"slug"`
	Content  string      `json:"content"`
	User     *User       `json:"user"`
	Created  time.Time   `json:"created"`
	Modified time.Time   `json:"modified"`
	Meta     []*PageMeta `json:"meta"`
}

type PageMeta struct {
	Key    string `json:"key"`
	Record string `json:"record"`
}

func (a PageMeta) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *PageMeta) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// IsLinkable exists to show that this method implements the Linkable type in
// graphql.
func (p *Page) IsLinkable() {}

// URI returns an absolute link to this post.
func (p *Page) URI() *URI {
	return NewURI(fmt.Sprintf("https://etu.natwelch.com/page/%s", p.Slug))
}

// Save inserts or updates a page into the database.
func (p *Page) Save(ctx context.Context) error {
	if p.Slug == "" {
		p.Slug = fmt.Sprintf("%s/%s", hexdate.Now().String(), neralie.Now().String())
	}

	if p.Created.IsZero() {
		p.Created = time.Now()
	}

	p.Modified = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO pages(slug, content, user_id, created_at, modified_at, meta)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (slug, user_id) DO UPDATE
SET (content, modified_at, meta) = ($2, $5, $6)
WHERE pages.slug = $1 AND pages.user_id = $3;
`,
		p.Slug,
		p.Content,
		p.User.ID,
		p.Created,
		p.Modified,
		p.Meta); err != nil {
		return err
	}

	return nil
}

// GetPageBySlug gets a page by ID from the database.
func GetPageBySlug(ctx context.Context, u *User, slug string) (*Page, error) {
	var p Page
	var userID string

	row := db.QueryRowContext(ctx,
		`SELECT slug, content, meta, user_id, created_at, modified_at
     FROM pages
     WHERE slug = $1 AND user_id = $2`,
		slug,
		u.ID)
	err := row.Scan(&p.Slug, &p.Content, &p.Meta, &userID, &p.Created, &p.Modified)
	switch {
	case err == sql.ErrNoRows:
		u, err := GetUser(ctx, userID)
		if err != nil {
			return nil, err
		}

		return &Page{User: u, Content: "Empty!", Slug: slug}, nil
	case err != nil:
		return nil, fmt.Errorf("error with get: %w", err)
	default:
		u, err := GetUser(ctx, userID)
		if err != nil {
			return nil, err
		}

		if u != nil {
			p.User = u
		}
		return &p, nil
	}
}

// GetPages returns an array of all pages that exist.
func GetPages(ctx context.Context, u *User, limit int, offset int) ([]*Page, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT slug, content, meta, user_id, created_at, modified_at
    FROM pages
    WHERE user_id = $1
    ORDER BY modified_at DESC
    LIMIT $2 OFFSET $3`,
		u.ID,
		limit,
		offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []*Page
	for rows.Next() {
		var p Page
		var userID string
		err := rows.Scan(&p.Slug, &p.Content, &p.Meta, &userID, &p.Created, &p.Modified)
		if err != nil {
			return nil, err
		}

		u, err := GetUser(ctx, userID)
		if err != nil {
			return nil, err
		}

		if u != nil {
			p.User = u
		}

		pages = append(pages, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}
