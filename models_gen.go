// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graphql

import (
	time "time"
)

type Comment struct {
	ID string `json:"id"`
}
type Link struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URI         string    `json:"uri"`
	Created     time.Time `json:"created"`
	Description string    `json:"description"`
	Screenshot  string    `json:"screenshot"`
	Tags        []string  `json:"tags"`
}
type NewLink struct {
	Title       string    `json:"title"`
	URI         string    `json:"uri"`
	Description string    `json:"description"`
	Tags        []*string `json:"tags"`
	Created     time.Time `json:"created"`
}
type NewPost struct {
	Content  string    `json:"content"`
	Title    string    `json:"title"`
	Datetime time.Time `json:"datetime"`
	Draft    bool      `json:"draft"`
}
type Stat struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
