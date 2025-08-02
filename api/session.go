package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/go-live-cms/go-live-cms/db/sqlc"
	"github.com/go-live-cms/go-live-cms/token"
	"github.com/go-live-cms/go-live-cms/util"
	"github.com/google/uuid"
)

type LoginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  UserResponse `json:"user"`
}

type RenewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RenewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

type BlockSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

type SessionResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	UserAgent string    `json:"user_agent"`
	ClientIP  string    `json:"client_ip"`
	IsBlocked bool      `json:"is_blocked"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func toSessionResponse(session db.Session) SessionResponse {
	return SessionResponse{
		ID:        session.ID,
		UserID:    session.UserID,
		Username:  session.Username,
		UserAgent: session.UserAgent,
		ClientIP:  session.ClientIp,
		IsBlocked: session.IsBlocked,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
	}
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req LoginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := server.store.GetUserByUsername(ctx.Request.Context(), req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(
		user.ID,
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	refreshToken, err := server.tokenMaker.CreateRefreshToken(
		user.ID,
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}

	userAgent := ctx.GetHeader("User-Agent")
	clientIP := ctx.ClientIP()

	sessionID := uuid.New()
	session, err := server.store.CreateSession(ctx.Request.Context(), db.CreateSessionParams{
		ID:           sessionID,
		UserID:       user.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIp:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(server.config.RefreshTokenDuration),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	rsp := LoginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  time.Now().Add(server.config.AccessTokenDuration),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: session.ExpiresAt,
		User:                  toUserResponse(user),
	}

	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req RenewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	if refreshPayload.TokenType != "refresh" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
		return
	}

	sessions, err := server.store.ListSessionsByUsername(ctx.Request.Context(), refreshPayload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sessions"})
		return
	}

	var session db.Session
	var sessionFound bool
	for _, s := range sessions {
		if s.RefreshToken == req.RefreshToken {
			session = s
			sessionFound = true
			break
		}
	}

	if !sessionFound {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "session not found"})
		return
	}

	if session.IsBlocked {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "session is blocked"})
		return
	}

	if time.Now().After(session.ExpiresAt) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
		return
	}

	currentUserAgent := ctx.GetHeader("User-Agent")
	ctx.ClientIP()

	if currentUserAgent != session.UserAgent {

		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "suspicious activity detected"})
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(
		refreshPayload.UserID,
		refreshPayload.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	rsp := RenewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: time.Now().Add(server.config.AccessTokenDuration),
	}

	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) getUserSessions(ctx *gin.Context) {

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	sessions, err := server.store.ListSessionsByUser(ctx.Request.Context(), authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sessions"})
		return
	}

	sessionResponses := make([]SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = toSessionResponse(session)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"sessions": sessionResponses,
		"count":    len(sessionResponses),
	})
}

func (server *Server) blockSession(ctx *gin.Context) {
	var req BlockSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	session, err := server.store.GetSession(ctx.Request.Context(), req.SessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get session"})
		return
	}

	if session.UserID != authPayload.UserID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not authorized to block this session"})
		return
	}

	err = server.store.BlockSession(ctx.Request.Context(), req.SessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block session"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "session blocked successfully",
		"session_id": req.SessionID,
	})
}

func (server *Server) logoutUser(ctx *gin.Context) {

	var refreshToken string

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := ctx.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		refreshToken = req.RefreshToken
	} else {

		authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

		sessions, err := server.store.ListSessionsByUser(ctx.Request.Context(), authPayload.UserID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sessions"})
			return
		}

		currentUserAgent := ctx.GetHeader("User-Agent")
		currentClientIP := ctx.ClientIP()

		for _, session := range sessions {
			if session.UserAgent == currentUserAgent && session.ClientIp == currentClientIP && !session.IsBlocked {
				refreshToken = session.RefreshToken
				break
			}
		}
	}

	if refreshToken == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "refresh token required"})
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	sessions, err := server.store.ListSessionsByUsername(ctx.Request.Context(), refreshPayload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sessions"})
		return
	}

	var sessionID uuid.UUID
	var sessionFound bool

	for _, session := range sessions {
		if session.RefreshToken == refreshToken {
			sessionID = session.ID
			sessionFound = true
			break
		}
	}

	if !sessionFound {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	err = server.store.BlockSession(ctx.Request.Context(), sessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}
