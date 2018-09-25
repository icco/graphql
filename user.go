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

func GetUser(id string) (*User, error) {
	user := GenerateUser(id)
	row := db.QueryRow("SELECT id, role, created_at, modified_at FROM users WHERE id = $1", id)
	err := row.Scan(user.ID, user.Role, user.Created, user.Modified)
	switch {
	case err == sql.ErrNoRows:
		// TODO: Maybe just return nil
		return nil, fmt.Errorf("No user with id %d", id)
	case err != nil:
		return nil, fmt.Errorf("Error running get query: %+v", err)
	default:
		return user, nil
	}
}

func GenerateUser(id string) *User {
	e := new(User)

	e.ID = id
	e.Role = "normal"

	// Computer generated content
	e.Created = time.Now()
	e.Modified = time.Now()

	return e
}
