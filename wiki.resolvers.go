package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
)

func (r *mutationResolver) UpsertPage(ctx context.Context, input EditPage) (*Page, error) {
	u := GetUserFromContext(ctx)
	p, err := GetPageBySlug(ctx, u, input.Slug)
	if err != nil {
		return nil, err
	}

	p.User = u
	p.Content = input.Content

	if err := p.Save(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *queryResolver) Photos(ctx context.Context, input *Limit) ([]*Photo, error) {
	u := GetUserFromContext(ctx)
	limit, offset := ParseLimit(input, 25, 0)

	return UserPhotos(ctx, u, limit, offset)
}

func (r *queryResolver) Page(ctx context.Context, slug string) (*Page, error) {
	u := GetUserFromContext(ctx)
	return GetPageBySlug(ctx, u, slug)
}

func (r *queryResolver) Pages(ctx context.Context, input *Limit) ([]*Page, error) {
	u := GetUserFromContext(ctx)
	limit, offset := ParseLimit(input, 25, 0)

	return GetPages(ctx, u, limit, offset)
}
