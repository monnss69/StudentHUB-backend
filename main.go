package handler

import (
	"backend/auth"
	"backend/db"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Handler for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize database on first request
	if db.DB == nil {
		db.Initialize()
	}

	router := gin.Default()

	// Simple CORS Configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://student-hub-frontend.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Authentication Routes
	router.POST("/users", db.CreateUser)
	router.GET("/users", db.ListUsers) // Supports ?username=query
	router.GET("/users/:id", db.GetUser)
	router.DELETE("/users/:id", auth.AuthMiddleware(), db.DeleteUser)

	// Auth routes
	router.POST("/login", db.Login)
	router.POST("/logout", db.Logout)

	// Post routes
	router.POST("/posts", auth.AuthMiddleware(), db.CreatePost)
	router.GET("/posts/:id", db.GetPost)
	router.GET("/posts/category/:category", db.ListPostsByCategory)
	router.PUT("/posts/:id", auth.AuthMiddleware(), db.UpdatePost)
	router.DELETE("/posts/:id", auth.AuthMiddleware(), db.DeletePost)

	// Comment routes
	router.GET("/posts/:id/comments", db.ListPostComments)

	// Category routes
	router.GET("/categories", db.ListCategories)

	// Serve the request
	router.ServeHTTP(w, r)
}
