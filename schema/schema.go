package schema

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	graphql "github.com/neelance/graphql-go"
)

var Schema = `
	schema {
		query: Query
	}
	# The query type, represents all of the entry points into our object graph
	type Query {
		posts(): [Post]!
		Post(id: ID!): Post 
	}
	type Post {
    id: ID!
    title: String!
    content: String!
    datetime: Time!
    created: Time!
    modified: Time!
    draft: Bool!
	}

  scalar Time
`

type Post struct {
	Id       graphql.ID `json:"-"`
	PostId   int64      `json:"id"`
	Title    string     `json:"title"` // optional
	Content  string     `json:"text"`  // Markdown
	Datetime time.Time  `json:"date"`
	Created  time.Time  `json:"created"`
	Modified time.Time  `json:"modified"`
	Tags     []string   `json:"tags"`
	Longform string     `json:"-"`
	Draft    bool       `json:"-"`
}
