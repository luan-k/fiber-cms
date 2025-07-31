package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/go-live-cms/go-live-cms/db/sqlc"
)

type CreateMediaRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description" binding:"required,min=5,max=500"`
	Alt         string `json:"alt" binding:"required,min=2,max=255"`
	MediaPath   string `json:"media_path" binding:"required,min=1,max=500"`
	PostID      *int64 `json:"post_id" binding:"omitempty"`
	Order       *int32 `json:"order" binding:"omitempty,min=0"`
}

type UpdateMediaRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=255"`
	Description string `json:"description" binding:"omitempty,min=5,max=500"`
	Alt         string `json:"alt" binding:"omitempty,min=2,max=255"`
	MediaPath   string `json:"media_path" binding:"omitempty,min=1,max=500"`
}

type MediaResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Alt         string    `json:"alt"`
	MediaPath   string    `json:"media_path"`
	UserID      int64     `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	ChangedAt   time.Time `json:"changed_at"`
	PostCount   *int64    `json:"post_count,omitempty"`
}

type PopularMediaResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Alt         string    `json:"alt"`
	MediaPath   string    `json:"media_path"`
	UserID      int64     `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	ChangedAt   time.Time `json:"changed_at"`
	PostCount   int64     `json:"post_count"`
}

func toMediaResponse(media db.Medium) MediaResponse {
	return MediaResponse{
		ID:          media.ID,
		Name:        media.Name,
		Description: media.Description,
		Alt:         media.Alt,
		MediaPath:   media.MediaPath,
		UserID:      media.UserID,
		CreatedAt:   media.CreatedAt,
		ChangedAt:   media.ChangedAt,
	}
}

func toMediaWithCountResponse(row db.ListMediaWithPostCountRow) MediaResponse {
	return MediaResponse{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Alt:         row.Alt,
		MediaPath:   row.MediaPath,
		UserID:      row.UserID,
		CreatedAt:   row.CreatedAt,
		ChangedAt:   row.ChangedAt,
		PostCount:   &row.PostCount,
	}
}

func toPopularMediaResponse(row db.GetPopularMediaRow) PopularMediaResponse {
	return PopularMediaResponse{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Alt:         row.Alt,
		MediaPath:   row.MediaPath,
		UserID:      row.UserID,
		CreatedAt:   row.CreatedAt,
		ChangedAt:   row.ChangedAt,
		PostCount:   row.PostCount,
	}
}

func (server *Server) createMedia(c *gin.Context) {
	var req CreateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := int64(1)

	if req.PostID != nil {

		var order int32
		if req.Order != nil {
			order = *req.Order
		} else {
			order = 0
		}

		result, err := server.store.CreateMediaAndLinkTx(c.Request.Context(), db.CreateMediaAndLinkTxParams{
			Name:        req.Name,
			Description: req.Description,
			Alt:         req.Alt,
			MediaPath:   req.MediaPath,
			UserID:      userID,
			PostID:      *req.PostID,
			Order:       order,
		})
		if err != nil {
			if containsString(err.Error(), "post not found") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "post not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create media with post link"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"media":      toMediaResponse(result.Media),
			"post_media": result.PostMedia,
		})
	} else {

		media, err := server.store.CreateMedia(c.Request.Context(), db.CreateMediaParams{
			Name:        req.Name,
			Description: req.Description,
			Alt:         req.Alt,
			MediaPath:   req.MediaPath,
			UserID:      userID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create media"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"media": toMediaResponse(media),
		})
	}
}

func (server *Server) getMediaByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	media, err := server.store.GetMedia(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media": toMediaResponse(media),
	})
}

func (server *Server) getMedia(c *gin.Context) {

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	withCounts := c.DefaultQuery("with_counts", "false")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	if withCounts == "true" {

		media, err := server.store.ListMediaWithPostCount(c.Request.Context(), db.ListMediaWithPostCountParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list media"})
			return
		}

		mediaResponses := make([]MediaResponse, len(media))
		for i, m := range media {
			mediaResponses[i] = toMediaWithCountResponse(m)
		}

		c.JSON(http.StatusOK, gin.H{
			"media": mediaResponses,
			"meta": gin.H{
				"limit":       limit,
				"offset":      offset,
				"count":       len(mediaResponses),
				"with_counts": true,
			},
		})
	} else {

		media, err := server.store.ListMedia(c.Request.Context(), db.ListMediaParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list media"})
			return
		}

		mediaResponses := make([]MediaResponse, len(media))
		for i, m := range media {
			mediaResponses[i] = toMediaResponse(m)
		}

		c.JSON(http.StatusOK, gin.H{
			"media": mediaResponses,
			"meta": gin.H{
				"limit":       limit,
				"offset":      offset,
				"count":       len(mediaResponses),
				"with_counts": false,
			},
		})
	}
}

func (server *Server) getPopularMedia(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}
	if limit > 50 {
		limit = 50
	}

	media, err := server.store.GetPopularMedia(c.Request.Context(), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get popular media"})
		return
	}

	mediaResponses := make([]PopularMediaResponse, len(media))
	for i, m := range media {
		mediaResponses[i] = toPopularMediaResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{
		"media": mediaResponses,
		"meta": gin.H{
			"limit": limit,
			"count": len(mediaResponses),
		},
	})
}

func (server *Server) searchMedia(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	media, err := server.store.SearchMediaByName(c.Request.Context(), db.SearchMediaByNameParams{
		Column1: sql.NullString{String: query, Valid: true},
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search media"})
		return
	}

	mediaResponses := make([]MediaResponse, len(media))
	for i, m := range media {
		mediaResponses[i] = toMediaResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{
		"media": mediaResponses,
		"meta": gin.H{
			"query":  query,
			"limit":  limit,
			"offset": offset,
			"count":  len(mediaResponses),
		},
	})
}

func (server *Server) getMediaByUser(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	_, err = server.store.GetUser(c.Request.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	media, err := server.store.GetMediaByUser(c.Request.Context(), db.GetMediaByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user media"})
		return
	}

	mediaResponses := make([]MediaResponse, len(media))
	for i, m := range media {
		mediaResponses[i] = toMediaResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{
		"media": mediaResponses,
		"meta": gin.H{
			"user_id": userID,
			"limit":   limit,
			"offset":  offset,
			"count":   len(mediaResponses),
		},
	})
}

func (server *Server) getMediaByPost(c *gin.Context) {
	postIDParam := c.Param("id")
	postID, err := strconv.ParseInt(postIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	post, err := server.store.GetPost(c.Request.Context(), postID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get post"})
		return
	}

	media, err := server.store.GetMediaByPost(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get post media"})
		return
	}

	mediaResponses := make([]MediaResponse, len(media))
	for i, m := range media {
		mediaResponses[i] = toMediaResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{
		"post":  toPostResponse(post),
		"media": mediaResponses,
		"meta": gin.H{
			"post_id": postID,
			"count":   len(mediaResponses),
		},
	})
}

func (server *Server) updateMedia(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	var req UpdateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingMedia, err := server.store.GetMedia(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get media"})
		return
	}

	updateParams := db.UpdateMediaParams{
		ID:          id,
		Name:        existingMedia.Name,
		Description: existingMedia.Description,
		Alt:         existingMedia.Alt,
		MediaPath:   existingMedia.MediaPath,
	}

	if req.Name != "" {
		updateParams.Name = req.Name
	}
	if req.Description != "" {
		updateParams.Description = req.Description
	}
	if req.Alt != "" {
		updateParams.Alt = req.Alt
	}
	if req.MediaPath != "" {
		updateParams.MediaPath = req.MediaPath
	}

	updatedMedia, err := server.store.UpdateMedia(c.Request.Context(), updateParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media": toMediaResponse(updatedMedia),
	})
}

func (server *Server) deleteMedia(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	_, err = server.store.GetMedia(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get media"})
		return
	}

	userID := int64(1)

	err = server.store.DeleteMediaTx(c.Request.Context(), db.DeleteMediaTxParams{
		MediaID: id,
		UserID:  userID,
	})
	if err != nil {
		if containsString(err.Error(), "permission denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own media"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "media deleted successfully",
	})
}
