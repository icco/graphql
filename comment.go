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

func GetComment(ctx context.Context, id string) (*Comment, error) {
	c := &Comment{}
	row := db.QueryRowContext(
		ctx, `
    SELECT id, post_id, user_id, content, created_at, modified_at
    FROM comments
    WHERE id = $1
    `, id)

	var postID, userID string
	err := row.Scan(
		&c.ID,
		&postID,
		&userID,
		&c.Content,
		&c.Created,
		&c.Modified,
	)
	if err != nil {
		return nil, err
	}

	c.User, err = GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	c.Post, err = GetPostString(ctx, postID)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func PostComments(ctx context.Context, p *Post, limit int, offset int) ([]*Comment, error) {
	if p == nil {
		return nil, fmt.Errorf("no post specified")
	}

	rows, err := db.QueryContext(
		ctx, `
    SELECT id, post_id, user_id, content, created_at, modified_at
    FROM comments
    WHERE post_id = $1
    ORDER BY created_at ASC
    LIMIT $2 OFFSET $3
    `, p.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*Comment, 0)
	for rows.Next() {
		c := &Comment{}
		var postID, userID string
		err := rows.Scan(
			&c.ID,
			&postID,
			&userID,
			&c.Content,
			&c.Created,
			&c.Modified,
		)
		if err != nil {
			return nil, err
		}

		c.User, err = GetUser(ctx, userID)
		if err != nil {
			return nil, err
		}

		c.Post, err = GetPostString(ctx, postID)
		if err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
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
