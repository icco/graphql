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
