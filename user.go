package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// User is a database object based off of what we get back from Google OAuth.
type User struct {
	ID       string
	Name     string
	Role     string
	APIKey   string
	Created  time.Time
	Modified time.Time
}

// Empty tells us if the user is real.
func (u User) Empty() bool {
	return u.ID == ""
}

// Save is an upsert based operation for User.
func (u *User) Save(ctx context.Context) error {
	_, err := db.ExecContext(ctx,
		`
    INSERT INTO users (id, role, created_at, modified_at, name)
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (id) DO UPDATE
    SET (role, modified_at, name) = ($2, $4, $5)
    WHERE users.id = $1;`,
		u.ID,
		u.Role,
		u.Created,
		time.Now(),
		u.Name)

	return err
}

// GetUser returns a user from the database. If the User does not exist, we
// create it.
func GetUser(ctx context.Context, id string) (*User, error) {
	var user User
	row := db.QueryRowContext(ctx, "SELECT id, role, apikey, name, created_at, modified_at FROM users WHERE id = $1", id)
	err := row.Scan(&user.ID, &user.Role, &user.APIKey, &user.Name, &user.Created, &user.Modified)

	switch {
	case err == sql.ErrNoRows:
		user.ID = id
		user.Role = "normal"
		user.Name = "anonymous"
		user.Created = time.Now()
		user.Modified = time.Now()
		return &user, (&user).Save(ctx)
	case err != nil:
		return nil, fmt.Errorf("error with get: %w", err)
	default:
		return &user, (&user).Save(ctx)
	}
}

// GetUserByAPIKey returns a user from the database.
func GetUserByAPIKey(ctx context.Context, apikey string) (*User, error) {
	var user User
	row := db.QueryRowContext(ctx, "SELECT id, role, apikey, name, created_at, modified_at FROM users WHERE apikey = $1", apikey)
	err := row.Scan(&user.ID, &user.Role, &user.APIKey, &user.Name, &user.Created, &user.Modified)
	if err != nil {
		return nil, fmt.Errorf("error with get: %w", err)
	}

	return &user, (&user).Save(ctx)
}
