package graphql

import (
	"context"
	"time"

	"github.com/lib/pq"
)

// Link is a link I have save on pinboard or a link in a post.
type Link struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URI         string    `json:"uri"`
	Created     time.Time `json:"created"`
	Description string    `json:"description"`
	Screenshot  string    `json:"screenshot"`
	Tags        []string  `json:"tags"`
}

// Save inserts or updates a link into the database.
func (l *Link) Save(ctx context.Context) error {
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
