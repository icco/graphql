package graphql

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
)

type key int8

const (
	userCtxKey key = 0
)

// GetUserFromContext finds the user from the context. This is usually inserted
// by WithUser.
func GetUserFromContext(ctx context.Context) *User {
	u, ok := ctx.Value(userCtxKey).(*User)
	if !ok {
		return nil
	}

	return u
}

// ParseLimit turns a limit and applies defaults into a pair of ints.
func ParseLimit(lim *Limit, defaultLimit, defaultOffset int) (int, int) {
	limit := defaultLimit
	offset := defaultOffset

	if lim != nil {
		i := *lim
		if i.Limit != nil {
			limit = *i.Limit
		}

		if i.Offset != nil {
			offset = *i.Offset
		}
	}

	return limit, offset
}

// WithUser puts a user in the context.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// Resolver is the type that gqlgen expects to exist
type Resolver struct{}

// New returns a Config that has all of the proper settings for this graphql
// server.
func New() Config {
	c := Config{
		Resolvers: &Resolver{},
	}

	c.Directives.HasRole = func(ctx context.Context, _ interface{}, next graphql.Resolver, role Role) (interface{}, error) {
		u := GetUserFromContext(ctx)
		if u == nil || Role(u.Role) != role {
			// block calling the next resolver
			return nil, fmt.Errorf("forbidden")
		}

		// or let it pass through
		return next(ctx)
	}

	c.Directives.LoggedIn = func(ctx context.Context, _ interface{}, next graphql.Resolver) (interface{}, error) {
		u := GetUserFromContext(ctx)
		if u == nil {
			// block calling the next resolver
			return nil, fmt.Errorf("forbidden")
		}

		// or let it pass through
		return next(ctx)
	}

	return c
}
