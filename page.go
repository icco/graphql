package graphql

import (
	"context"
	"time"

	"github.com/google/uuid"
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

	if p.Created.IsZero() {
		p.Created = time.Now()
	}

	p.Modified = time.Now()

	return nil
}
