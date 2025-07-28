package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	db "github.com/luan-k/fiber-cms/db/sqlc"
)

type getPostsRequest struct {
	Limit  int32 `form:"limit" binding:"required,min=1,max=100"`
	Offset int32 `form:"offset" binding:"required,min=0"`
}

func (server *Server) getPosts(c *gin.Context) {
	var req getPostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	posts, err := server.store.ListPosts(c.Request.Context(), db.ListPostsParams{
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"meta": gin.H{
			"limit":  req.Limit,
			"offset": req.Offset,
			"count":  len(posts),
		},
	})
}

func (server *Server) getPostByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	post, err := server.store.GetPost(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"post": post,
	})
}
