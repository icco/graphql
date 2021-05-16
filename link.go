package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// Link is a link I have save on pinboard or a link in a post.
type Link struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URI         URI       `json:"uri"`
	Created     time.Time `json:"created"`
	Description string    `json:"description"`
	Screenshot  URI       `json:"screenshot"`
	Tags        []string  `json:"tags"`
	Modified    time.Time `json:"modified"`
}

// Save inserts or updates a link into the database.
func (l *Link) Save(ctx context.Context) error {
	if l.Created.IsZero() {
		l.Created = time.Now()
	}

	l.Modified = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO links(title, uri, description, created, created_at, modified_at, tags)
VALUES ($1, $2, $3, $4, $6, $6, $5)
ON CONFLICT (uri) DO UPDATE
SET (title, description, created, modified_at, tags) = ($1, $3, $4, $6, $5)
WHERE links.uri = $2;
`,
		l.Title,
		l.URI,
		l.Description,
		l.Created,
		pq.Array(l.Tags),
		time.Now(),
	); err != nil {
		return err
	}

	return nil
}

// IsLinkable exists to show that this method implements the Linkable type in
// graphql.
func (l *Link) IsLinkable() {}

// GetLinkByURI gets a link by uri from the database.
func GetLinkByURI(ctx context.Context, uri string) (*Link, error) {
	var link Link
	row := db.QueryRowContext(ctx, "SELECT id, title, uri, description, created, modified_at, tags FROM links WHERE uri = $1", uri)
	err := row.Scan(&link.ID, &link.Title, &link.URI, &link.Description, &link.Created, &link.Modified, pq.Array(&link.Tags))
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("no link %q", uri)
	case err != nil:
		return nil, fmt.Errorf("error with get: %w", err)
	default:
		return &link, nil
	}
}

// GetLinkByID gets a link by id from the database.
func GetLinkByID(ctx context.Context, id string) (*Link, error) {
	var link Link
	row := db.QueryRowContext(ctx, "SELECT id, title, uri, description, created, modified_at, tags FROM links WHERE id = $1", id)
	err := row.Scan(&link.ID, &link.Title, &link.URI, &link.Description, &link.Created, &link.Modified, pq.Array(&link.Tags))
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("no link %q", id)
	case err != nil:
		return nil, fmt.Errorf("error with get: %w", err)
	default:
		return &link, nil
	}
}

// GetLinks returns all links from the database.
func GetLinks(ctx context.Context, limit int, offset int) ([]*Link, error) {
	rows, err := db.QueryContext(ctx, `
  SELECT id, title, uri, description, created, modified_at, tags
  FROM links
  ORDER BY created DESC
  LIMIT $1 OFFSET $2`,
		limit,
		offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]*Link, 0)
	for rows.Next() {
		link := new(Link)
		err := rows.Scan(&link.ID, &link.Title, &link.URI, &link.Description, &link.Created, &link.Modified, pq.Array(&link.Tags))
		if err != nil {
			return nil, err
		}

		if link.Created.IsZero() {
			link.Created = time.Now()
		}

		links = append(links, link)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
