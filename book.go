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

// IsLinkable exists to show that this method implements the Linkable type in
// graphql.
func (Book) IsLinkable() {}

// Save inserts or updates a book into the database.
func (b *Book) Save(ctx context.Context) error {
	if b.ID == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		b.ID = uuid.String()
	}

	if b.Created.IsZero() {
		b.Created = time.Now()
	}

	b.Modified = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO books(id, title, goodreads_id, created_at, modified_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
SET (title, goodreads_id, created_at, modified_at) = ($2, $3, $4, $5)
WHERE books.id = $1;
`,
		b.ID,
		b.Title,
		b.GoodreadsID,
		b.Created,
		b.Modified); err != nil {
		return err
	}

	return nil
}

// URI returns an absolute link to this book.
func (b *Book) URI() URI {
	return URI{fmt.Sprintf("https://www.goodreads.com/book/show/%s", b.GoodreadsID)}
}
