//go:generate gorunpkg github.com/99designs/gqlgen

package writing

import (
	context "context"
)

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreatePost(ctx context.Context, input NewPost) (Post, error) {
	panic("not implemented")
}
func (r *mutationResolver) EditPost(ctx context.Context, Id string, input NewPost) (Post, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateLink(ctx context.Context, input NewLink) (Link, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) AllPosts(ctx context.Context) ([]*Post, error) {
	panic("not implemented")
}
func (r *queryResolver) Posts(ctx context.Context, limit *int, offset *int) ([]*Post, error) {
	panic("not implemented")
}
func (r *queryResolver) Post(ctx context.Context, id string) (*Post, error) {
	panic("not implemented")
}
func (r *queryResolver) AllLinks(ctx context.Context) ([]*Link, error) {
	panic("not implemented")
}
func (r *queryResolver) Links(ctx context.Context, limit *int, offset *int) ([]*Link, error) {
	panic("not implemented")
}
func (r *queryResolver) Link(ctx context.Context, id string) (*Link, error) {
	panic("not implemented")
}
