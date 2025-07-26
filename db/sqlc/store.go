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
	DeleteUserTx(ctx context.Context, id int64) error
	DeleteUserWithTransferTx(ctx context.Context, arg DeleteUserWithTransferTxParams) error
	UpdateUserTx(ctx context.Context, arg UpdateUserTxParams) (UpdateUserTxResult, error)
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

func (store *SQLStore) DeleteUserTx(ctx context.Context, id int64) error {
	err := store.execTx(ctx, func(q *Queries) error {
		err := q.DeleteUserSessions(ctx, id)
		if err != nil {
			return err
		}

		err = q.DeleteUserPostsByUserID(ctx, id)
		if err != nil {
			return err
		}

		err = q.DeletePostsByUserID(ctx, id)
		if err != nil {
			return err
		}

		err = q.DeleteUser(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

type UpdateUserTxParams struct {
	UpdateUserParams
	CheckUniqueness bool
}

type UpdateUserTxResult struct {
	User User `json:"user"`
}

func (store *SQLStore) UpdateUserTx(ctx context.Context, arg UpdateUserTxParams) (UpdateUserTxResult, error) {
	var result UpdateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		if arg.CheckUniqueness {
			existingUser, err := q.GetUserByUsername(ctx, arg.Username)
			if err == nil && existingUser.ID != arg.ID {
				return fmt.Errorf("username '%s' already exists", arg.Username)
			}

			existingUser, err = q.GetUserByEmail(ctx, arg.Email)
			if err == nil && existingUser.ID != arg.ID {
				return fmt.Errorf("email '%s' already exists", arg.Email)
			}
		}

		result.User, err = q.UpdateUser(ctx, arg.UpdateUserParams)
		if err != nil {
			return err
		}

		if arg.Username != "" {
			err = q.UpdatePostsUsername(ctx, UpdatePostsUsernameParams{
				UserID:   arg.ID,
				Username: arg.Username,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

type DeleteUserWithTransferTxParams struct {
	UserID       int64
	TransferToID int64
}

func (store *SQLStore) DeleteUserWithTransferTx(ctx context.Context, arg DeleteUserWithTransferTxParams) error {
	err := store.execTx(ctx, func(q *Queries) error {
		err := q.TransferPostsToAdmin(ctx, TransferPostsToAdminParams{
			UserID:   arg.UserID,
			UserID_2: arg.TransferToID,
		})
		if err != nil {
			return err
		}

		err = q.UpdateUserPostsOwnership(ctx, UpdateUserPostsOwnershipParams{
			UserID:   arg.UserID,
			UserID_2: arg.TransferToID,
		})
		if err != nil {
			return err
		}

		err = q.DeleteUserSessions(ctx, arg.UserID)
		if err != nil {
			return err
		}

		err = q.DeleteUser(ctx, arg.UserID)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
