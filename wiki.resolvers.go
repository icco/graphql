package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
)

func (r *logResolver) Sector(ctx context.Context, obj *Log) (WorkSector, error) {
	panic(fmt.Errorf("not implemented"))
}

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
	u := GetUserFromContext(ctx)
	p, err := GetPageBySlug(ctx, u, input.Slug)
	if err != nil {
		return nil, err
	}

	p.User = u
	p.Content = input.Content
	p.Meta = &PageMetaGrouping{
		Records: input.Meta,
	}

	if err := p.Save(ctx); err != nil {
		return nil, err
	}

	return p, nil
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

func (r *queryResolver) Logs(ctx context.Context, input *Limit) ([]*Log, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Log(ctx context.Context, id string) (*Log, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Photos(ctx context.Context, input *Limit) ([]*Photo, error) {
	u := GetUserFromContext(ctx)
	limit, offset := ParseLimit(input, 25, 0)

	return UserPhotos(ctx, u, limit, offset)
}

// Log returns LogResolver implementation.
func (r *Resolver) Log() LogResolver { return &logResolver{r} }

type logResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *logResolver) Duration(ctx context.Context, obj *Log) (*Duration, error) {
	panic(fmt.Errorf("not implemented"))
}
