package handler

import (
	"backend/auth"
	"backend/db"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Handler exports the function for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize database connection if not already done
	if db.DB == nil {
		db.Initialize()
	}

	// Set up Gin in release mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.New() // Use New() instead of Default() to avoid unnecessary middleware

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://student-hub-frontend.vercel.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           30 * 24 * time.Hour,
	}))

	// Routes
	setupRoutes(router)

	// Serve the request
	router.ServeHTTP(w, r)
}

// setupRoutes configures all the routes
func setupRoutes(router *gin.Engine) {
	// Check route
	router.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API is running"})
	})

	// Authentication Routes
	router.POST("/api/users", db.CreateUser)
	router.GET("/api/users", auth.AuthMiddleware(), db.ListUsers)
	router.GET("/api/users/:id", auth.AuthMiddleware(), db.GetUser)
	router.DELETE("/api/users/:id", auth.AuthMiddleware(), db.DeleteUser)
	router.GET("/api/users/:id/posts", auth.AuthMiddleware(), db.GetUserPost)
	router.PUT("/api/users/:id", auth.AuthMiddleware(), db.UpdateUser)

	// Auth routes
	router.POST("/api/login", db.Login)
	router.POST("/api/logout", db.Logout)
	router.POST("/api/auth/sync", db.SyncToken)

	// Post routes
	router.POST("/api/posts", auth.AuthMiddleware(), db.CreatePost)
	router.GET("/api/posts/:id", db.GetPost)
	router.GET("/api/posts/category/:category/:pageIndex", auth.AuthMiddleware(), db.ListPostsByCategory)
	router.PUT("/api/posts/:id", auth.AuthMiddleware(), db.UpdatePost)
	router.DELETE("/api/posts/:id", auth.AuthMiddleware(), db.DeletePost)

	// Tag routes
	router.GET("/api/tags", db.ListTags)
	router.GET("/api/tags/:id", db.GetTag)
	router.GET("/api/posts/:id/tags", auth.AuthMiddleware(), db.ListPostTags)
	router.POST("/api/posts/:id/tags", auth.AuthMiddleware(), db.CreatePostTag)
	router.DELETE("/api/posts/:id/tags/:tag_id", auth.AuthMiddleware(), db.DeletePostTag)

	// Comment routes
	router.GET("/api/posts/:id/comments", auth.AuthMiddleware(), db.ListPostComments)
	router.POST("/api/posts/:id/comments", auth.AuthMiddleware(), db.CreateComment)

	// Category routes
	router.GET("/api/categories", db.ListCategories)
	router.GET("/api/categories/:id", db.GetCategory)

	// Image routes
	router.POST("/api/cloudinary/upload", db.UploadHandler)
	router.DELETE("/api/cloudinary/upload/:username", db.DeleteImageHandler)
}
