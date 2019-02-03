package graphql

import (
	"context"
	"time"

	"github.com/google/uuid"
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
		l.Modified); err != nil {
		return err
	}

	return nil
}

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

func UserLogs(ctx context.Context, u *User) ([]*Log, error) {
	rows, err := db.QueryContext(
		ctx,
		"SELECT id, code, datetime, description, location, project, user_id, created_at, modified_at FROM logs WHERE user_id = $1 ORDER BY date DESC",
		u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*Log, 0)
	for rows.Next() {
		l := &Log{}
		err := rows.Scan(&l.ID, &l.Code, &l.Datetime, &l.Description, &l.Location, &l.Project, &l.User.ID, &l.Created, &l.Modified)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}
