package main

import (
	"backend/auth"
	"backend/db"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
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
		MaxAge:           30 * 24 * time.Hour,
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
	router.PUT("/users/:id", auth.AuthMiddleware(), db.UpdateUser)

	// Auth routes
	router.POST("/login", db.Login)
	router.POST("/logout", db.Logout)
	router.POST("/auth/sync", db.SyncToken)

	// Post routes
	router.POST("/posts", auth.AuthMiddleware(), db.CreatePost)
	router.GET("/posts/:id", db.GetPost)
	router.GET("/posts/category/:category/:pageIndex", auth.AuthMiddleware(), db.ListPostsByCategory)
	router.PUT("/posts/:id", auth.AuthMiddleware(), db.UpdatePost)
	router.DELETE("/posts/:id", auth.AuthMiddleware(), db.DeletePost)

	// Tag routes
	router.GET("/tags", db.ListTags)
	router.GET("/tags/:id", db.GetTag)
	router.GET("/posts/:id/tags", auth.AuthMiddleware(), db.ListPostTags)
	router.POST("/posts/:id/tags", auth.AuthMiddleware(), db.CreatePostTag)
	router.DELETE("/posts/:id/tags/:tag_id", auth.AuthMiddleware(), db.DeletePostTag)

	// Comment routes
	router.GET("/posts/:id/comments", auth.AuthMiddleware(), db.ListPostComments)
	router.POST("/posts/:id/comments", auth.AuthMiddleware(), db.CreateComment)

	// Category routes
	router.GET("/categories", db.ListCategories)
	router.GET("/categories/:id", db.GetCategory)

	// Image routes
	router.POST("/cloudinary/upload", db.UploadHandler)
	router.DELETE("/cloudinary/upload/:username", db.DeleteImageHandler)

	http.ListenAndServe(":8080", router)
}
