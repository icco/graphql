package graphql

import (
	"context"
	"time"
)

// Cache is a basic type as defined by gqlgen.
type Cache struct {
}

// NewCache creates a new cache.
func NewCache() (*Cache, error) {
	return &Cache{}, nil
}

// Add inserts a key value pair into the database.
func (c *Cache) Add(ctx context.Context, hash string, query interface{}) {
	_, err := db.ExecContext(
		ctx,
		`
INSERT INTO cache(key, value, modified_at)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE
SET (value, modified_at) = ($2, $3)
WHERE cache.key = $1;
`,
		hash,
		query,
		time.Now())

	if err != nil {
		log.WithError(err).Error("could not insert key")
	}
}

// Get retrieves a value by a key.
func (c *Cache) Get(ctx context.Context, hash string) (interface{}, bool) {
	var value string
	row := db.QueryRowContext(ctx, "SELECT value FROM cache WHERE key = $1", hash)
	err := row.Scan(&value)

	if err != nil {
		return "", false
	}
	return value, true
}
