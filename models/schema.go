package models

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
