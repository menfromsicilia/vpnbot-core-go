package api

import (
	"github.com/gofiber/fiber/v2"
	"vpnbot-core-go/internal/middleware"
)

func SetupRoutes(app *fiber.App, handlers *Handlers, apiKey string) {
	api := app.Group("/api")

	// Public endpoint
	api.Get("/health", handlers.Health)

	// Protected endpoints
	secured := api.Use(middleware.APIKeyMiddleware(apiKey))

	// User operations
	secured.Post("/create", handlers.CreateUser)
	secured.Delete("/deleteUser", handlers.DeleteUser)
	secured.Post("/getUsers", handlers.GetUsers)
	secured.Post("/getInbounds", handlers.GetInbounds)

	// Server management
	servers := secured.Group("/servers")
	servers.Get("/", handlers.GetServers)
	servers.Post("/", handlers.PostServers)
	servers.Put("/", handlers.PutServers)
	servers.Delete("/", handlers.DeleteServers)

	// Statistics
	secured.Get("/stats", handlers.GetStats)
	
	// Users
	secured.Get("/users", handlers.GetAllUsers)
	secured.Get("/users/count", handlers.GetUsersCount)
	
	// Nodes with users
	secured.Get("/nodes/users", handlers.GetNodesWithUsers)
	
	// Cleanup
	secured.Get("/cleanup/pending", handlers.GetPendingDeletionsHandler)
	secured.Post("/cleanup", handlers.CleanupPendingDeletionsHandler)
	secured.Delete("/cleanup/pending", handlers.DeletePendingDeletionHandler)
}

