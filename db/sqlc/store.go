package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	CreatePostTx(ctx context.Context, arg CreatePostTxParams) (CreatePostTxResult, error)
	DeletePostTx(ctx context.Context, id int64) error
}

// SQLStore gives us the functions to interact with the database
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type CreatePostTxParams struct {
	CreatePostsParams
	AuthorIDs []int64
}

type CreatePostTxResult struct {
	Post      Post       `json:"post"`
	UserPosts []UserPost `json:"user_posts"`
}

func (store *SQLStore) CreatePostTx(ctx context.Context, arg CreatePostTxParams) (CreatePostTxResult, error) {
	var result CreatePostTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Post, err = q.CreatePosts(ctx, arg.CreatePostsParams)
		if err != nil {
			return err
		}

		userPost, err := q.CreateUserPost(ctx, CreateUserPostParams{
			PostID: result.Post.ID,
			UserID: arg.UserID,
			Order:  0,
		})
		if err != nil {
			return err
		}
		result.UserPosts = append(result.UserPosts, userPost)

		for i, authorID := range arg.AuthorIDs {
			if authorID != arg.UserID {
				userPost, err := q.CreateUserPost(ctx, CreateUserPostParams{
					PostID: result.Post.ID,
					UserID: authorID,
					Order:  int32(i + 1),
				})
				if err != nil {
					return err
				}
				result.UserPosts = append(result.UserPosts, userPost)
			}
		}

		return nil
	})

	return result, err
}

func (store *SQLStore) DeletePostTx(ctx context.Context, id int64) error {
	err := store.execTx(ctx, func(q *Queries) error {
		err := q.DeleteUserPost(ctx, id)
		if err != nil {
			return err
		}

		err = q.DeletePost(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
