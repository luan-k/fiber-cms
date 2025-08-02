package api

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	db "github.com/go-live-cms/go-live-cms/db/sqlc"
	"github.com/go-live-cms/go-live-cms/token"
	"github.com/go-live-cms/go-live-cms/util"
)

type Server struct {
	store      db.Store
	router     *gin.Engine
	config     util.Config
	tokenMaker token.Maker
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create token maker: %w", err)
	}
	server := &Server{
		store:      store,
		config:     config,
		tokenMaker: tokenMaker,
	}

	server.setupRoutes()
	return server, nil
}

func (server *Server) setupRoutes() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	v1 := router.Group("/api/v1")

	router.GET("/health", server.healthCheck)

	auth := v1.Group("/auth")
	auth.POST("/register", server.register)
	auth.POST("/login", server.loginUser)
	auth.POST("/refresh", server.renewAccessToken)
	auth.POST("/logout", authMiddleware(server.tokenMaker), server.logoutUser)

	sessions := v1.Group("/sessions")
	sessions.Use(authMiddleware(server.tokenMaker))
	sessions.GET("", server.getUserSessions)    // GET /api/v1/sessions
	sessions.PUT("/block", server.blockSession) // PUT /api/v1/sessions/block

	users := v1.Group("/users")
	users.POST("", server.createUser)                          // POST /api/v1/users
	users.GET("", server.getUsers)                             // implement content limiter // GET /api/v1/users
	users.GET("/:id", server.getUserByID)                      // GET /api/v1/users/:id
	users.GET("/username/:username", server.getUserByUsername) // GET /api/v1/users/username/:username
	users.GET("/email/:email", server.getUserByEmail)          // GET /api/v1/users/email/:email
	users.PUT("/:id", server.updateUser)                       // PUT /api/v1/users/:id
	users.DELETE("/:id", server.deleteUser)                    // DELETE /api/v1/users/:id

	posts := v1.Group("/posts")
	posts.POST("", server.createPost)                      // POST /api/v1/posts
	posts.GET("", server.getPosts)                         // GET /api/v1/posts
	posts.GET("/:id", server.getPostByID)                  // GET /api/v1/posts/:id
	posts.PUT("/:id", server.updatePost)                   // PUT /api/v1/posts/:id
	posts.DELETE("/:id", server.deletePost)                // DELETE /api/v1/posts/:id
	posts.GET("/user/:id", server.getPostsByUser)          // GET /api/v1/posts/user/:id
	posts.GET("/:id/taxonomies", server.getPostTaxonomies) // GET /api/v1/posts/:id/taxonomies

	taxonomies := v1.Group("/taxonomies")
	taxonomies.POST("", server.createTaxonomy)              // POST /api/v1/taxonomies
	taxonomies.GET("", server.getTaxonomies)                // GET /api/v1/taxonomies
	taxonomies.GET("/popular", server.getPopularTaxonomies) // GET /api/v1/taxonomies/popular
	taxonomies.GET("/search", server.searchTaxonomies)      // GET /api/v1/taxonomies/search
	taxonomies.GET("/:id", server.getTaxonomyByID)          // GET /api/v1/taxonomies/:id
	taxonomies.GET("/name/:name", server.getTaxonomyByName) // GET /api/v1/taxonomies/name/:name
	taxonomies.PUT("/:id", server.updateTaxonomy)           // PUT /api/v1/taxonomies/:id
	taxonomies.DELETE("/:id", server.deleteTaxonomy)        // DELETE /api/v1/taxonomies/:id
	taxonomies.GET("/:id/posts", server.getTaxonomyPosts)   // GET /api/v1/taxonomies/:id/posts

	media := v1.Group("/media")
	media.POST("", server.createMedia)            // POST /api/v1/media
	media.GET("", server.getMedia)                // GET /api/v1/media
	media.GET("/popular", server.getPopularMedia) // GET /api/v1/media/popular
	media.GET("/search", server.searchMedia)      // GET /api/v1/media/search
	media.GET("/:id", server.getMediaByID)        // GET /api/v1/media/:id
	media.PUT("/:id", server.updateMedia)         // PUT /api/v1/media/:id
	media.DELETE("/:id", server.deleteMedia)      // DELETE /api/v1/media/:id
	media.GET("/user/:id", server.getMediaByUser) // GET /api/v1/media/user/:id
	media.GET("/post/:id", server.getMediaByPost) // GET /api/v1/media/post/:id

	server.router = router
}

func (server *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Gin CMS API is running",
		"version": "v0.0.1",
	})
}

func (server *Server) register(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Register endpoint - coming soon",
	})
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
