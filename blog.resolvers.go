package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
)

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

func (r *mutationResolver) CreatePost(ctx context.Context, input EditPost) (*Post, error) {
	return r.EditPost(ctx, input)
}

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

func (r *queryResolver) Drafts(ctx context.Context, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return Drafts(ctx, limit, offset)
}

func (r *queryResolver) FuturePosts(ctx context.Context, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return FuturePosts(ctx, limit, offset)
}

func (r *queryResolver) Posts(ctx context.Context, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return Posts(ctx, limit, offset)
}

func (r *queryResolver) Comments(ctx context.Context, input *Limit) ([]*Comment, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Search(ctx context.Context, query string, input *Limit) ([]*Post, error) {
	limit, offset := ParseLimit(input, 10, 0)

	return Search(ctx, query, limit, offset)
}

func (r *queryResolver) Post(ctx context.Context, id string) (*Post, error) {
	return GetPostString(ctx, id)
}

func (r *queryResolver) NextPost(ctx context.Context, id string) (*Post, error) {
	p, err := GetPostString(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.Next(ctx)
}

func (r *queryResolver) PrevPost(ctx context.Context, id string) (*Post, error) {
	p, err := GetPostString(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.Prev(ctx)
}

func (r *queryResolver) PostsByTag(ctx context.Context, id string) ([]*Post, error) {
	return PostsByTag(ctx, id)
}

func (r *queryResolver) Tags(ctx context.Context) ([]string, error) {
	return AllTags(ctx)
}
