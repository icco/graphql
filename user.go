package graphql

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID       string
	Role     string
	Created  time.Time
	Modified time.Time
}

func (u *User) Save() error {
	_, err := db.Exec(
		`
    INSERT INTO users (id, role, created_at, modified_at)
    VALUES ($1, $2, $3, $4)
    ON CONFLICT (id) DO UPDATE
    SET (role, modified_at) = ($2, $4)
    WHERE users.id = $1;`,
		u.ID,
		u.Role,
		u.Created,
		time.Now())

	return err
}

func GetUser(id string) (*User, error) {
	user := GenerateUser(id)
	row := db.QueryRow("SELECT id, role, created_at, modified_at FROM users WHERE id = $1", id)
	err := row.Scan(user.ID, user.Role, user.Created, user.Modified)

	switch {
	case err == sql.ErrNoRows:
		user.Role = "normal"
		user.Created = time.Now()
		user.Modified = time.Now()
		return user, user.Save()
	case err != nil:
		return nil, fmt.Errorf("Error running get query: %+v", err)
	default:
		return user, nil
	}
}

func GenerateUser(id string) *User {
	e := new(User)

	e.ID = id

	// Computer generated content
	e.Created = time.Now()
	e.Modified = time.Now()

	return e
}
