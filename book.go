package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Book is a book on Goodreads.
type Book struct {
	ID          string `json:"id"`
	GoodreadsID string
	Title       string `json:"title"`
	Link        string
	Authors     []string
	ISBN        string
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

func (Book) IsLinkable() {}

// Save inserts or updates a book into the database.
func (p *Book) Save(ctx context.Context) error {
	if p.ID == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		p.ID = uuid.String()
	}

	if p.Created.IsZero() {
		p.Created = time.Now()
	}

	p.Modified = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO books(id, title, goodreads_id, created_at, modified_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
SET (title, goodreads_id, created_at, modified_at) = ($2, $3, $4, $5)
WHERE books.id = $1;
`,
		p.ID,
		p.Title,
		p.GoodreadsID,
		p.Created,
		p.Modified); err != nil {
		return err
	}

	return nil
}

func (b *Book) URI() string {
	return fmt.Sprintf("https://goodreads.com/%s", b.GoodreadsID)
}
