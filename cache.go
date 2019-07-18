package graphql

import (
	"time"
)

type Cache struct {
}

const apqPrefix = "apq:"

func NewCache() (*Cache, error) {
	return &Cache{}, nil
}

func (c *Cache) Add(ctx context.Context, hash string, query string) error {
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

	return err
}

func (c *Cache) Get(ctx context.Context, hash string) (string, bool) {
	var value string
	row := db.QueryRowContext(ctx, "SELECT value FROM cache WHERE key = $1", hash)
	err := row.Scan(&value)

	if err != nil {
		return "", false
	}
	return value, true
}
