package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paulmach/orb"
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
	Duration    Duration  `json:"duration"`
	Created     time.Time
	Modified    time.Time
}

// IsLinkable exists to show that this method implements the Linkable type in
// graphql.
func (l *Log) IsLinkable() {}

// URI returns the URI for this log.
func (l *Log) URI() *URI {
	url := fmt.Sprintf("https://etu.natwelch.com/log/%s", l.ID)
	return NewURI(url)
}

// Save inserts or updates a log into the database.
func (l *Log) Save(ctx context.Context) error {
	if l.ID == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		l.ID = uuid.String()
	}

	if l.Datetime.IsZero() {
		l.Datetime = time.Now()
	}

	if l.Created.IsZero() {
		l.Created = time.Now()
	}

	l.Modified = time.Now()

	loc, err := GeoConvertValue(l.Location)
	if err != nil {
		return err
	}

	if l.User.Empty() {
		return fmt.Errorf("no user specified")
	}

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO logs(id, code, datetime, description, location, project, user_id, created_at, modified_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE
SET (code, datetime, description, location, project, user_id, created_at, modified_at) = ($2, $3, $4, $5, $6, $7, $8, $9)
WHERE logs.id = $1;
`,
		l.ID,
		l.Code,
		l.Datetime,
		l.Description,
		loc,
		l.Project,
		l.User.ID,
		l.Created,
		l.Modified,
	); err != nil {
		return fmt.Errorf("upsert log: %w", err)
	}

	return nil
}

// SetUser looks up a user by ID and then sets it for this log.
func (l *Log) SetUser(ctx context.Context, id string) error {
	u, err := GetUser(ctx, id)
	if err != nil {
		return err
	}

	if u != nil {
		l.User = *u
	}

	return nil
}

// UserLogs gets all logs for a User.
func UserLogs(ctx context.Context, u *User, limit int, offset int) ([]*Log, error) {
	if u == nil {
		return nil, fmt.Errorf("no user specified")
	}

	rows, err := db.QueryContext(
		ctx, `
    SELECT id, code, datetime, description, ST_AsBinary(location), project, user_id, created_at, modified_at
    FROM logs
    WHERE user_id = $1
    ORDER BY datetime DESC
    LIMIT $2 OFFSET $3
    `,
		u.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*Log, 0)
	for rows.Next() {
		l := &Log{}
		var p orb.Point

		err := rows.Scan(
			&l.ID,
			&l.Code,
			&l.Datetime,
			&l.Description,
			GeoScanner(&p),
			&l.Project,
			&l.User.ID,
			&l.Created,
			&l.Modified,
		)
		if err != nil {
			return nil, err
		}

		l.Location = GeoFromOrb(&p)

		logs = append(logs, l)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

// GetLog gets a single Log by ID.
func GetLog(ctx context.Context, id string) (*Log, error) {
	l := &Log{}
	var p orb.Point

	row := db.QueryRowContext(ctx, `
  SELECT id, code, datetime, description, ST_AsBinary(location), project, user_id, created_at, modified_at
  FROM logs
  WHERE id = $1
  `, id)
	err := row.Scan(
		&l.ID,
		&l.Code,
		&l.Datetime,
		&l.Description,
		GeoScanner(&p),
		&l.Project,
		&l.User.ID,
		&l.Created,
		&l.Modified,
	)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("error with get: %w", err)
	default:
		l.Location = GeoFromOrb(&p)
		return l, nil
	}
}
