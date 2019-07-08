package graphql

import (
	"time"
)

// Comment is a comment on a post.
type Comment struct {
	ID       string    `json:"id"`
	Post     *Post     `json:"post"`
	Author   *User     `json:"author"`
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

// Save adds the photo to the database and checks that no data is missing.
func (p *Photo) Save(ctx context.Context) error {
	if p.ID == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		p.ID = uuid.String()
	}

	if p.Created.IsZero() {
		p.Created = time.Now()
	}

	p.Modified = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO photos(id, year, content_type, user_id, created_at, modified_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
SET (year, content_type, user_id, created_at, modified_at) = ($2, $3, $4, $5, $6)
WHERE photos.id = $1;
`,
		p.ID,
		p.Year,
		p.ContentType,
		p.User.ID,
		p.Created,
		p.Modified); err != nil {
		return err
	}

	return nil
}
