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
	return NewURI(fmt.Sprintf("https://www.goodreads.com/book/show/%s", b.GoodreadsID))
}

// GetLinks returns all links from the database.
func GetBooks(ctx context.Context, limit int, offset int) ([]*Book, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, title, goodreads_id, created_at, modified_at FROM books LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]*Book, 0)
	for rows.Next() {
		book := new(Book)
		err := rows.Scan(&book.ID, &book.Title, &book.GoodreadsID, &book.Created, &book.Modified)
		if err != nil {
			return nil, err
		}

		if book.Created.IsZero() {
			book.Created = time.Now()
		}

		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}
