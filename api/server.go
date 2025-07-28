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

	auth := v1.Group("/auth")
	auth.POST("/register", server.register)
	auth.POST("/login", server.login)

	v1.GET("/posts", server.getPosts)
	v1.GET("/posts/:id", server.getPostByID)

	server.router = router
}

func (server *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Gin CMS API is running",
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
