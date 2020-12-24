package graphql

import (
	"context"
	"fmt"
	"time"
)

// Save upserts a stat.
func (s *Stat) Save(ctx context.Context) error {
	if s.Key == "" {
		return fmt.Errorf("Empty key not allowed")
	}

	s.When = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO stats(key, value, when) VALUES ($1, $2, $3)`,
		s.Key,
		s.Value,
		s.When,
	); err != nil {
		return err
	}

	return nil
}

// GetStats returns the limit of the most recently updated stats.
func GetStats(ctx context.Context, limit int) ([]*Stat, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT DISTINCT ON (key) key, value, when
    FROM stats
    ORDER by key, when DESC
    LIMIT $1`,
		limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*Stat
	for rows.Next() {
		stat := new(Stat)
		if err := rows.Scan(&stat.Key, &stat.Value, &stat.When); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetStat returns the history of a stat
func GetStat(ctx context.Context, key string, limit int, offset int) ([]*Stat, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT key, value, when
    FROM stats
    WHERE key = $1
    ORDER by when DESC
    LIMIT $2 OFFSET $3`,
		key,
		limit,
		offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*Stat
	for rows.Next() {
		stat := new(Stat)
		if err := rows.Scan(&stat.Key, &stat.Value, &stat.When); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}
