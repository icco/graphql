package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Comment is a comment on a post.
type Comment struct {
	ID       string    `json:"id"`
	Post     *Post     `json:"post"`
	User     *User     `json:"author"`
	Content  string    `json:"content"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// IsLinkable tells gqlgen this model has a URI function.
func (Comment) IsLinkable() {}

// URI returns the URI for this comment.
func (c *Comment) URI() *URI {
	return NewURI("https://writing.natwelch.com/comment/" + c.ID)
}

// Save adds the comment to the database and checks that no data is missing.
func (c *Comment) Save(ctx context.Context) error {
	if c.ID == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		c.ID = uuid.String()
	}

	if c.Created.IsZero() {
		c.Created = time.Now()
	}

	c.Modified = time.Now()

	if c.Post == nil {
		return fmt.Errorf("post cannot be nil")
	}

	if c.User == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO comments(id, post_id, user_id, content, created_at, modified_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
SET (post_id, user_id, content, created_at, modified_at) = ($2, $3, $4, $5, $6)
WHERE comments.id = $1;
`,
		c.ID,
		c.Post.ID,
		c.User.ID,
		c.Content,
		c.Created,
		c.Modified); err != nil {
		return err
	}

	return nil
}
