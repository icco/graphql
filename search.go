package graphql

import (
	"context"
)

func Search(ctx context.Context, searchQuery string, limit int, offset int) ([]*Post, error) {
	query := `
SELECT id, title, content, date, created_at, modified_at, tags, draft
FROM posts
WHERE id in (
  SELECT id
  FROM posts, plainto_tsquery('year') query, to_tsvector(title || ' ' || content) textsearch
  WHERE draft = false
    AND date <= NOW()
    AND query @@ textsearch
  ORDER BY ts_rank_cd(textsearch, query) DESC
)
LIMIT $2 OFFSET $3
`
	return postQuery(ctx, query, searchQuery, limit, offset)
}
