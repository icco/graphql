package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

// A Log is a journal entry by an individual.
type Log struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Datetime    time.Time `json:"datetime"`
	Description string    `json:"description"`
	Location    *Geo      `json:"location"`
	Project     string    `json:"project"`
	User        User      `json:"user"`
	Created     time.Time
	Modified    time.Time
}

// Save inserts or updates a post into the database.
func (l *Log) Save(ctx context.Context) error {

	if p.Created.IsZero() {
		l.Created = time.Now()
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

	return nil
}
