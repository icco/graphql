package graphql

import (
	"context"
	"fmt"
)

func (s *Stat) Save(ctx context.Context) error {
	return fmt.Errorf("not implemented")
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
