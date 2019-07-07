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
	URI      URI       `json:"uri"`
}

func (Comment) IsLinkable() {}
