package main

import (
	"log"
	"marketprogo/config"
	"marketprogo/internal/handlers"
	"marketprogo/internal/middleware"
	"marketprogo/pkg/database"
	"marketprogo/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	logger.InitLogger()

	// Initialize database
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create Gin router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/refresh", handlers.RefreshToken)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.Auth())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", handlers.GetUserProfile)
				users.PUT("/profile", handlers.UpdateUserProfile)
			}

			// Product routes
			products := protected.Group("/products")
			{
				products.GET("", handlers.GetProducts)
				products.GET("/:id", handlers.GetProduct)
				products.POST("", handlers.CreateProduct)
				products.PUT("/:id", handlers.UpdateProduct)
				products.DELETE("/:id", handlers.DeleteProduct)
			}

			// Order routes
			orders := protected.Group("/orders")
			{
				orders.GET("", handlers.GetOrders)
				orders.GET("/:id", handlers.GetOrder)
				orders.POST("", handlers.CreateOrder)
				orders.PUT("/:id", handlers.UpdateOrder)
			}

			// B2B specific routes
			b2b := protected.Group("/b2b")
			{
				// Contract routes
				contracts := b2b.Group("/contracts")
				{
					contracts.GET("", handlers.GetContracts)
					contracts.GET("/:id", handlers.GetContract)
					contracts.POST("", handlers.CreateContract)
					contracts.PUT("/:id", handlers.UpdateContract)
				}

				// Purchase order routes
				pos := b2b.Group("/purchase-orders")
				{
					pos.GET("", handlers.GetPurchaseOrders)
					pos.GET("/:id", handlers.GetPurchaseOrder)
					pos.POST("", handlers.CreatePurchaseOrder)
					pos.PUT("/:id", handlers.UpdatePurchaseOrder)
				}
			}
		}
	}

	// Start server
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
