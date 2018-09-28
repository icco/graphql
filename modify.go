package graphql

import (
	"context"
	"time"
)

func GetMaxId(ctx context.Context) (int64, error) {
	row := db.QueryRowContext(ctx, "SELECT MAX(id) from posts")
	var id int64
	if err := row.Scan(&id); err != nil {
		return -1, err
	}

	return id, nil
}

func CreatePost(ctx context.Context, input Post) (Post, error) {

	maxId, err := GetMaxId(ctx)
	if err != nil {
		return Post{}, err
	}
	id := maxId + 1

	_, err = db.ExecContext(ctx, "INSERT INTO posts(id, title, content, date, draft, created_at, modified_at) VALUES ($1, $2, $3, $4, $5, $6, $6)",
		id,
		input.Title,
		input.Content,
		input.Datetime,
		input.Draft,
		time.Now(),
	)
	if err != nil {
		return Post{}, err
	}

	post, err := GetPost(ctx, id)
	if err != nil {
		return Post{}, err
	}

	return *post, nil
}
