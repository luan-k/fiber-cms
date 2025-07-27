package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	db "github.com/luan-k/fiber-cms/db/sqlc"
)

type Server struct {
	store db.Store
	app   *fiber.App
}

func NewServer(store db.Store) *Server {
	server := &Server{
		store: store,
	}

	server.app = fiber.New(fiber.Config{
		ErrorHandler: server.errorHandler,
		AppName:      "Fiber CMS API",
	})

	server.setupRoutes()

	return server
}

func (server *Server) errorHandler(c *fiber.Ctx, err error) error {

	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": message,
	})
}

func (server *Server) setupRoutes() {

	server.app.Use(recover.New())

	server.app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))

	server.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	v1 := server.app.Group("/api/v1")

	server.app.Get("/health", server.healthCheck)

	auth := v1.Group("/auth")
	auth.Post("/register", server.register)
	auth.Post("/login", server.login)
	v1.Get("/posts", server.getPosts)
	v1.Get("/posts/:id", server.getPostByID)
}

func (server *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Fiber CMS API is running",
	})
}

func (server *Server) register(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Register endpoint - coming soon",
	})
}

func (server *Server) login(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Login endpoint - coming soon",
	})
}

func (server *Server) Start(address string) error {
	return server.app.Listen(address)
}
