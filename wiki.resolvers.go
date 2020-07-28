package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
)

func (r *mutationResolver) InsertLog(ctx context.Context, input NewLog) (*Log, error) {
	l := &Log{}
	l.Code = input.Code
	l.Project = input.Project

	u := GetUserFromContext(ctx)
	if u != nil {
		l.User = *u
	}

	if input.Description != nil {
		l.Description = *input.Description
	}

	if input.Location != nil {
		l.Location = &Geo{
			Lat:  input.Location.Lat,
			Long: input.Location.Long,
		}
	}

	if input.Duration != nil {
		l.Duration = ParseDurationFromString(*input.Duration)
	}

	if err := l.Save(ctx); err != nil {
		return nil, err
	}

	return l, nil
}

func (r *mutationResolver) UpsertPage(ctx context.Context, input EditPage) (*Page, error) {
	var err error
	p := &Page{}

	if input.ID != nil {
		p, err = GetPageByID(ctx, *input.ID)
		if err != nil {
			return nil, err
		}
	}

	p.Content = input.Content
	p.Title = input.Title

	u := GetUserFromContext(ctx)
	if u != nil {
		p.User = *u
	}

	if input.Slug != nil {
		p.Slug = *input.Slug
	}

	if input.Category != nil {
		p.Category = *input.Category
	}

	err = p.Save(ctx)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *queryResolver) Logs(ctx context.Context, input *Limit) ([]*Log, error) {
	u := GetUserFromContext(ctx)
	limit, offset := ParseLimit(input, 25, 0)

	return UserLogs(ctx, u, limit, offset)
}

func (r *queryResolver) Log(ctx context.Context, id string) (*Log, error) {
	return GetLog(ctx, id)
}

func (r *queryResolver) GetPageByID(ctx context.Context, id string) (*Page, error) {
	return GetPageByID(ctx, id)
}

func (r *queryResolver) GetPageBySlug(ctx context.Context, slug string) (*Page, error) {
	return GetPageBySlug(ctx, slug)
}

func (r *queryResolver) GetPages(ctx context.Context) ([]*Page, error) {
	return GetPages(ctx)
}

func (r *queryResolver) Photos(ctx context.Context, input *Limit) ([]*Photo, error) {
	u := GetUserFromContext(ctx)
	limit, offset := ParseLimit(input, 25, 0)

	return UserPhotos(ctx, u, limit, offset)
}
