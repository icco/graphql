package graphql

import (
	"context"
	"fmt"
	"time"
)

func (s *Stat) Save(ctx context.Context) error {
	if s.Key == "" {
		return fmt.Errorf("Empty key not allowed")
	}

	s.Modified = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO stats(key, value, modified_at)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE
SET (value, modified_at) = ($2, $3)
WHERE stats.key = $1;
`,
		s.Key,
		s.Value,
		s.Modified,
	); err != nil {
		return err
	}

	return nil
}

func GetStats(ctx context.Context, limit int) ([]*Stat, error) {
	rows, err := db.QueryContext(ctx, "SELECT key, value, modified_at FROM stats ORDER BY modified_at DESC LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*Stat
	for rows.Next() {
		stat := new(Stat)
		if err := rows.Scan(&stat.Key, &stat.Value, &stat.Modified); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}
