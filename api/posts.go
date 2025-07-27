package api

import (
	"github.com/gofiber/fiber/v2"
	db "github.com/luan-k/fiber-cms/db/sqlc"
)

type getPostsRequest struct {
	Limit  int32 `query:"limit"`
	Offset int32 `query:"offset"`
}

func (server *Server) getPosts(c *fiber.Ctx) error {
	var req getPostsRequest
	if err := c.QueryParser(&req); err != nil {
		return err
	}

	posts, err := server.store.ListPosts(c.Context(), db.ListPostsParams{
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		return err
	}
	return c.JSON(posts)
}

type getPostByIDRequest struct {
	ID int64 `param:"id"`
}

func (server *Server) getPostByID(c *fiber.Ctx) error {
	var req getPostByIDRequest
	if err := c.ParamsParser(&req); err != nil {
		return err
	}
	post, err := server.store.GetPost(c.Context(), req.ID)
	if err != nil {
		return err
	}
	return c.JSON(post)
}
