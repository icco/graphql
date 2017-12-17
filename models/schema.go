package models

import (
	"log"
	"strconv"

	graphql "github.com/neelance/graphql-go"
)

var Schema = `
	schema {
		query: Query
	}
	# The query type, represents all of the entry points into our object graph
	type Query {
		Posts(): [Post]!
		Post(Id: ID!): Post 
	}
	type Post {
    Id: ID!
    Title: String!
    Content: String!
    Datetime: Time!
    Created: Time!
    Modified: Time!
    Draft: Boolean!
	}

  scalar Time
`

type Resolver struct{}

func (r *Resolver) Post(args struct{ ID graphql.ID }) (*Post, error) {
	id, err := strconv.ParseInt(string(args.ID), 10, 64)
	if err != nil {
		log.Printf("Got error trying to convert id: %+v: %+v", args.ID, err)
		return nil, err
	}

	post, err := GetPost(id)
	if err != nil {
		log.Printf("Got error trying to get post %+v: %+v", args.ID, err)
		return nil, err
	}
	return post, nil
}

func (r *Resolver) Posts() []*Post {
	posts, err := AllPosts()
	if err != nil {
		log.Printf("Got error trying to get posts: %+v", err)
		return nil
	}

	return posts
}
