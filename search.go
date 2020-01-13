package graphql

import (
	"context"
)

// Search searches for posts that have matching titles, content or tags
func Search(ctx context.Context, searchQuery string, limit int, offset int) ([]*Post, error) {
	query := `
SELECT id, title, content, date, created_at, modified_at, tags, draft
FROM posts
WHERE id in (
  SELECT id
  FROM posts, plainto_tsquery($1) query, to_tsvector(title || ' ' || content || array_to_tsvector(tags)) textsearch
  WHERE draft = false
    AND date <= NOW()
    AND query @@ textsearch
  ORDER BY ts_rank_cd(textsearch, query) DESC
)
LIMIT $2 OFFSET $3
`
	return postQuery(ctx, query, searchQuery, limit, offset)
}
