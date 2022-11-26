package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
)

// AddComment is the resolver for the addComment field.
func (r *mutationResolver) AddComment(ctx context.Context, input AddComment) (*Comment, error) {
	c := &Comment{}
	c.Content = input.Content
	c.User = GetUserFromContext(ctx)

	post, err := GetPostString(ctx, input.PostID)
	if err != nil {
		return nil, err
	}
	c.Post = post

	err = c.Save(ctx)
	if err != nil {
		return nil, err
	}

	return GetComment(ctx, c.ID)
}

// CreatePost is the resolver for the createPost field.
func (r *mutationResolver) CreatePost(ctx context.Context, input EditPost) (*Post, error) {
	return r.EditPost(ctx, input)
}

// EditPost is the resolver for the editPost field.
func (r *mutationResolver) EditPost(ctx context.Context, input EditPost) (*Post, error) {
	var err error
	p := &Post{}

	// We do this so the defaults in save don't overwrite stuff on upsert.
	if input.ID != nil {
		p, err = GetPostString(ctx, *input.ID)
		if err != nil {
			return nil, err
		}

		if p == nil {
			return nil, fmt.Errorf("cannot edit post that does not exist")
		}
	}

	if input.Title != nil {
		p.Title = *input.Title
	}

	if input.Content != nil {
		p.Content = *input.Content
	}

	if input.Datetime != nil {
		p.Datetime = *input.Datetime
	}

	if input.Draft != nil {
		p.Draft = *input.Draft
	} else {
		p.Draft = true
	}

	err = p.Save(ctx)
	if err != nil {
		return nil, err
	}

	return GetPostString(ctx, p.ID)
}

// Drafts is the resolver for the drafts field.
func (r *queryResolver) Drafts(ctx context.Context, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return Drafts(ctx, limit, offset)
}

// FuturePosts is the resolver for the futurePosts field.
func (r *queryResolver) FuturePosts(ctx context.Context, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return FuturePosts(ctx, limit, offset)
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return Posts(ctx, limit, offset)
}

// Comments is the resolver for the comments field.
func (r *queryResolver) Comments(ctx context.Context, input *Limit) ([]*Comment, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return AllComments(ctx, limit, offset)
}

// Search is the resolver for the search field.
func (r *queryResolver) Search(ctx context.Context, query string, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return Search(ctx, query, limit, offset)
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*Post, error) {
	return GetPostString(ctx, id)
}

// NextPost is the resolver for the nextPost field.
func (r *queryResolver) NextPost(ctx context.Context, id string) (*Post, error) {
	p, err := GetPostString(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.Next(ctx)
}

// PrevPost is the resolver for the prevPost field.
func (r *queryResolver) PrevPost(ctx context.Context, id string) (*Post, error) {
	p, err := GetPostString(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.Prev(ctx)
}

// PostsByTag is the resolver for the postsByTag field.
func (r *queryResolver) PostsByTag(ctx context.Context, id string) ([]*Post, error) {
	return PostsByTag(ctx, id)
}

// Tags is the resolver for the tags field.
func (r *queryResolver) Tags(ctx context.Context) ([]string, error) {
	return AllTags(ctx)
}
