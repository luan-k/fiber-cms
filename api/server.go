package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	db "github.com/luan-k/fiber-cms/db/sqlc"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := &Server{
		store: store,
	}

	server.setupRoutes()
	return server
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

	// todo: implement auth middleware
	auth := v1.Group("/auth")
	auth.POST("/register", server.register)
	auth.POST("/login", server.login)

	users := v1.Group("/users")
	users.POST("", server.createUser)
	users.GET("", server.getUsers) // todo: implement content limiter
	users.GET("/:id", server.getUserByID)
	users.GET("/username/:username", server.getUserByUsername)
	users.GET("/email/:email", server.getUserByEmail)
	users.PUT("/:id", server.updateUser)
	users.DELETE("/:id", server.deleteUser)

	posts := v1.Group("/posts")
	posts.POST("", server.createPost)       // POST /api/v1/posts
	posts.GET("", server.getPosts)          // GET /api/v1/posts
	posts.GET("/:id", server.getPostByID)   // GET /api/v1/posts/:id
	posts.PUT("/:id", server.updatePost)    // PUT /api/v1/posts/:id
	posts.DELETE("/:id", server.deletePost) // DELETE /api/v1/posts/:id
	posts.GET("/user/:id", server.getPostsByUser)
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

func (server *Server) login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Login endpoint - coming soon",
	})
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
