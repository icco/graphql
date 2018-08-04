//go:generate gorunpkg github.com/99designs/gqlgen

package writing

import (
	context "context"

	models "github.com/icco/writing/models"
	graph "github.com/icco/writing/server/graph"
)

type Resolver struct{}

func (r *Resolver) Mutation() graph.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() graph.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreatePost(ctx context.Context, input models.NewPost) (models.Post, error) {
	panic("not implemented")
}
func (r *mutationResolver) EditPost(ctx context.Context, Id string, input models.NewPost) (models.Post, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateLink(ctx context.Context, input models.NewLink) (models.Link, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Posts(ctx context.Context) ([]*models.Post, error) {
	panic("not implemented")
}
func (r *queryResolver) Posts(ctx context.Context, limit *int, offset *int) ([]*models.Post, error) {
	panic("not implemented")
}
func (r *queryResolver) Post(ctx context.Context, id string) (*models.Post, error) {
	panic("not implemented")
}
func (r *queryResolver) Links(ctx context.Context) ([]*models.Link, error) {
	panic("not implemented")
}
func (r *queryResolver) Links(ctx context.Context, limit *int, offset *int) ([]*models.Link, error) {
	panic("not implemented")
}
func (r *queryResolver) Link(ctx context.Context, id string) (*models.Link, error) {
	panic("not implemented")
}
