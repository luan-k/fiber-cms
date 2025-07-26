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

	CreatePostWithTaxonomiesTx(ctx context.Context, arg CreatePostWithTaxonomiesTxParams) (CreatePostWithTaxonomiesTxResult, error)
	DeleteTaxonomyTx(ctx context.Context, id int64) error
	UpdatePostTaxonomiesTx(ctx context.Context, arg UpdatePostTaxonomiesTxParams) error
	CreateTaxonomyAndLinkTx(ctx context.Context, arg CreateTaxonomyAndLinkTxParams) (CreateTaxonomyAndLinkTxResult, error)

	CreatePostWithImagesTx(ctx context.Context, arg CreatePostWithImagesTxParams) (CreatePostWithImagesTxResult, error)
	DeleteImageTx(ctx context.Context, arg DeleteImageTxParams) error
	UpdatePostImagesTx(ctx context.Context, arg UpdatePostImagesTxParams) error
	CreateImageAndLinkTx(ctx context.Context, arg CreateImageAndLinkTxParams) (CreateImageAndLinkTxResult, error)

	ExecTx(ctx context.Context, fn func(*Queries) error) error
}

func (store *SQLStore) ExecTx(ctx context.Context, fn func(*Queries) error) error {
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

		err = q.DeleteImagesByUserID(ctx, id)
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

		err = q.TransferImagesToUser(ctx, TransferImagesToUserParams{
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

type CreatePostWithTaxonomiesTxParams struct {
	CreatePostsParams
	AuthorIDs   []int64
	TaxonomyIDs []int64
}

type CreatePostWithTaxonomiesTxResult struct {
	Post           Post            `json:"post"`
	UserPosts      []UserPost      `json:"user_posts"`
	PostTaxonomies []PostsTaxonomy `json:"post_taxonomies"`
}

func (store *SQLStore) CreatePostWithTaxonomiesTx(ctx context.Context, arg CreatePostWithTaxonomiesTxParams) (CreatePostWithTaxonomiesTxResult, error) {
	var result CreatePostWithTaxonomiesTxResult

	err := store.ExecTx(ctx, func(q *Queries) error {
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

		for _, taxonomyID := range arg.TaxonomyIDs {

			_, err := q.GetTaxonomy(ctx, taxonomyID)
			if err != nil {
				return fmt.Errorf("taxonomy %d not found: %w", taxonomyID, err)
			}

			postTaxonomy, err := q.CreatePostTaxonomy(ctx, CreatePostTaxonomyParams{
				PostID:     result.Post.ID,
				TaxonomyID: taxonomyID,
			})
			if err != nil {
				return err
			}
			result.PostTaxonomies = append(result.PostTaxonomies, postTaxonomy)
		}

		return nil
	})

	return result, err
}

func (store *SQLStore) DeleteTaxonomyTx(ctx context.Context, id int64) error {
	err := store.ExecTx(ctx, func(q *Queries) error {

		err := q.DeleteTaxonomyPosts(ctx, id)
		if err != nil {
			return err
		}

		err = q.DeleteTaxonomy(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

type UpdatePostTaxonomiesTxParams struct {
	PostID      int64
	TaxonomyIDs []int64
}

func (store *SQLStore) UpdatePostTaxonomiesTx(ctx context.Context, arg UpdatePostTaxonomiesTxParams) error {
	err := store.ExecTx(ctx, func(q *Queries) error {

		err := q.DeletePostTaxonomies(ctx, arg.PostID)
		if err != nil {
			return err
		}

		for _, taxonomyID := range arg.TaxonomyIDs {

			_, err := q.GetTaxonomy(ctx, taxonomyID)
			if err != nil {
				return fmt.Errorf("taxonomy %d not found: %w", taxonomyID, err)
			}

			_, err = q.CreatePostTaxonomy(ctx, CreatePostTaxonomyParams{
				PostID:     arg.PostID,
				TaxonomyID: taxonomyID,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

type CreateTaxonomyAndLinkTxParams struct {
	Name        string
	Description string
	PostID      int64
}

type CreateTaxonomyAndLinkTxResult struct {
	Taxonomy     Taxonomy      `json:"taxonomy"`
	PostTaxonomy PostsTaxonomy `json:"post_taxonomy"`
}

func (store *SQLStore) CreateTaxonomyAndLinkTx(ctx context.Context, arg CreateTaxonomyAndLinkTxParams) (CreateTaxonomyAndLinkTxResult, error) {
	var result CreateTaxonomyAndLinkTxResult

	err := store.ExecTx(ctx, func(q *Queries) error {
		var err error

		existingTaxonomy, err := q.GetTaxonomyByName(ctx, arg.Name)
		if err == nil {
			result.Taxonomy = existingTaxonomy
		} else {
			result.Taxonomy, err = q.CreateTaxonomy(ctx, CreateTaxonomyParams{
				Name:        arg.Name,
				Description: arg.Description,
			})
			if err != nil {
				return err
			}
		}

		_, err = q.GetPost(ctx, arg.PostID)
		if err != nil {
			return fmt.Errorf("post %d not found: %w", arg.PostID, err)
		}

		existing, _ := q.GetPostTaxonomies(ctx, arg.PostID)
		for _, t := range existing {
			if t.ID == result.Taxonomy.ID {
				result.PostTaxonomy = PostsTaxonomy{
					PostID:     arg.PostID,
					TaxonomyID: result.Taxonomy.ID,
				}
				return nil
			}
		}

		result.PostTaxonomy, err = q.CreatePostTaxonomy(ctx, CreatePostTaxonomyParams{
			PostID:     arg.PostID,
			TaxonomyID: result.Taxonomy.ID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

type CreatePostWithImagesTxParams struct {
	CreatePostsParams
	AuthorIDs []int64
	ImageIDs  []int64
}

type CreatePostWithImagesTxResult struct {
	Post       Post        `json:"post"`
	UserPosts  []UserPost  `json:"user_posts"`
	PostImages []PostImage `json:"post_images"`
}

type DeleteImageTxParams struct {
	ImageID int64
	UserID  int64
}

type UpdatePostImagesTxParams struct {
	PostID   int64
	ImageIDs []int64
}

type CreateImageAndLinkTxParams struct {
	Name        string
	Description string
	Alt         string
	ImagePath   string
	UserID      int64
	PostID      int64
	Order       int32
}

type CreateImageAndLinkTxResult struct {
	Image     Image     `json:"image"`
	PostImage PostImage `json:"post_image"`
}

func (store *SQLStore) CreatePostWithImagesTx(ctx context.Context, arg CreatePostWithImagesTxParams) (CreatePostWithImagesTxResult, error) {
	var result CreatePostWithImagesTxResult

	err := store.ExecTx(ctx, func(q *Queries) error {
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

		for i, imageID := range arg.ImageIDs {

			_, err := q.GetImage(ctx, imageID)
			if err != nil {
				return fmt.Errorf("image %d not found: %w", imageID, err)
			}

			postImage, err := q.CreatePostImage(ctx, CreatePostImageParams{
				PostID:  result.Post.ID,
				ImageID: imageID,
				Order:   int32(i),
			})
			if err != nil {
				return err
			}
			result.PostImages = append(result.PostImages, postImage)
		}

		return nil
	})

	return result, err
}

func (store *SQLStore) DeleteImageTx(ctx context.Context, arg DeleteImageTxParams) error {
	err := store.ExecTx(ctx, func(q *Queries) error {

		image, err := q.GetImage(ctx, arg.ImageID)
		if err != nil {
			return err
		}

		if image.UserID != arg.UserID {
			return fmt.Errorf("user %d does not own image %d", arg.UserID, arg.ImageID)
		}

		err = q.DeleteImagePosts(ctx, arg.ImageID)
		if err != nil {
			return err
		}

		err = q.DeleteImage(ctx, arg.ImageID)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (store *SQLStore) UpdatePostImagesTx(ctx context.Context, arg UpdatePostImagesTxParams) error {
	err := store.ExecTx(ctx, func(q *Queries) error {

		_, err := q.GetPost(ctx, arg.PostID)
		if err != nil {
			return fmt.Errorf("post %d not found: %w", arg.PostID, err)
		}

		err = q.DeletePostImages(ctx, arg.PostID)
		if err != nil {
			return err
		}

		for i, imageID := range arg.ImageIDs {

			_, err := q.GetImage(ctx, imageID)
			if err != nil {
				return fmt.Errorf("image %d not found: %w", imageID, err)
			}

			_, err = q.CreatePostImage(ctx, CreatePostImageParams{
				PostID:  arg.PostID,
				ImageID: imageID,
				Order:   int32(i),
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (store *SQLStore) CreateImageAndLinkTx(ctx context.Context, arg CreateImageAndLinkTxParams) (CreateImageAndLinkTxResult, error) {
	var result CreateImageAndLinkTxResult

	err := store.ExecTx(ctx, func(q *Queries) error {
		var err error

		result.Image, err = q.CreateImage(ctx, CreateImageParams{
			Name:        arg.Name,
			Description: arg.Description,
			Alt:         arg.Alt,
			ImagePath:   arg.ImagePath,
			UserID:      arg.UserID,
		})
		if err != nil {
			return err
		}

		_, err = q.GetPost(ctx, arg.PostID)
		if err != nil {
			return fmt.Errorf("post %d not found: %w", arg.PostID, err)
		}

		result.PostImage, err = q.CreatePostImage(ctx, CreatePostImageParams{
			PostID:  arg.PostID,
			ImageID: result.Image.ID,
			Order:   arg.Order,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
