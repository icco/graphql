package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
)

func (r *mutationResolver) InsertLog(ctx context.Context, input NewLog) (*Log, error) {
	l := &Log{
		Project: input.Project,
		Started: input.Started,
		Stopped: input.Stopped,
		Sector:  input.Sector,
	}

	u := GetUserFromContext(ctx)
	if u != nil {
		l.User = *u
	}

	if input.Description != nil {
		l.Description = *input.Description
	}

	if err := l.Save(ctx); err != nil {
		return nil, err
	}

	return l, nil
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
