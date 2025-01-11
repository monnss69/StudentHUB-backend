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
		AllowOrigins:     []string{"https://student-hub-frontend.vercel.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Check route
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API is running"})
	})

	// Authentication Routes
	router.POST("/users", db.CreateUser)
	router.GET("/users", auth.AuthMiddleware(), db.ListUsers)
	router.GET("/users/:id", auth.AuthMiddleware(), db.GetUser)
	router.DELETE("/users/:id", auth.AuthMiddleware(), db.DeleteUser)
	router.GET("users/:id/posts", auth.AuthMiddleware(), db.GetUserPost)

	// Auth routes
	router.POST("/login", db.Login)
	router.POST("/logout", db.Logout)

	// Post routes
	router.POST("/posts", auth.AuthMiddleware(), db.CreatePost)
	router.GET("/posts/:id", db.GetPost)
	router.GET("/posts/category/:category", auth.AuthMiddleware(), db.ListPostsByCategory)
	router.PUT("/posts/:id", auth.AuthMiddleware(), db.UpdatePost)
	router.DELETE("/posts/:id", auth.AuthMiddleware(), db.DeletePost)

	// Tag routes
	router.GET("/tags", db.ListTags)
	router.GET("/tags/:id", db.GetTag)
	router.GET("/posts/:id/tags", db.ListPostTags)

	// Comment routes
	router.GET("/posts/:id/comments", auth.AuthMiddleware(), db.ListPostComments)
	router.POST("/posts/:id/comments", auth.AuthMiddleware(), db.CreateComment)

	// Category routes
	router.GET("/categories", db.ListCategories)
	router.GET("/categories/:id", db.GetCategory)

	// Serve the request
	router.ServeHTTP(w, r)
}
