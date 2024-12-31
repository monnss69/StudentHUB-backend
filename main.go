package handler

import (
	"backend/auth"
	"backend/db"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Middleware to handle CORS manually
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allowed origins
		allowedOrigins := []string{
			"https://student-hub-frontend.vercel.app",
			"http://localhost:3000",
		}

		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		isAllowedOrigin := false
		for _, allowedOrigin := range allowedOrigins {
			if strings.EqualFold(origin, allowedOrigin) {
				isAllowedOrigin = true
				break
			}
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			if isAllowedOrigin {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.AbortWithStatus(204)
				return
			}
		}

		// Set CORS headers for non-OPTIONS requests
		if isAllowedOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Next()
	}
}

// Handler for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize database on first request
	if db.DB == nil {
		db.Initialize()
	}

	router := gin.Default()

	// Use custom CORS middleware
	router.Use(corsMiddleware())

	// Authentication Routes
	router.POST("/auth", db.AuthenticateUser)
	router.POST("/auth/logout", db.LogOutUser)

	// User Routes
	router.POST("/users", db.CreateUser)
	router.DELETE("/users/:id", auth.AuthMiddleware(), db.DeleteUser)
	router.GET("/users", db.GetAllUser)
	router.GET("/users/:id", db.GetUserID)

	// Post Routes
	router.POST("/post", auth.AuthMiddleware(), db.CreatePost)
	router.DELETE("/post/:post_id", auth.AuthMiddleware(), db.DeletePost)
	router.GET("/post/:post_id", auth.AuthMiddleware(), db.GetPostID)
	router.PUT("/post/:post_id", auth.AuthMiddleware(), db.EditPost)
	router.GET("/:category", auth.AuthMiddleware(), db.GetCategoryPost)

	// Comment Route
	router.GET("/comment/:post_id", auth.AuthMiddleware(), db.GetCommentPost)

	// Category Route
	router.GET("/category", db.GetCategory)

	// Serve the request
	router.ServeHTTP(w, r)
}

// Main function for local development
func main() {
	db.Initialize()

	router := gin.Default()

	// Use custom CORS middleware
	router.Use(corsMiddleware())

	// Same route setup as in Handler
	router.POST("/auth", db.AuthenticateUser)
	router.POST("/auth/logout", db.LogOutUser)
	router.POST("/users", db.CreateUser)
	router.DELETE("/users/:id", auth.AuthMiddleware(), db.DeleteUser)
	router.GET("/users", db.GetAllUser)
	router.GET("/users/:id", db.GetUserID)
	router.POST("/post", auth.AuthMiddleware(), db.CreatePost)
	router.DELETE("/post/:post_id", auth.AuthMiddleware(), db.DeletePost)
	router.GET("/post/:post_id", auth.AuthMiddleware(), db.GetPostID)
	router.PUT("/post/:post_id", auth.AuthMiddleware(), db.EditPost)
	router.GET("/:category", auth.AuthMiddleware(), db.GetCategoryPost)
	router.GET("/comment/:post_id", auth.AuthMiddleware(), db.GetCommentPost)
	router.GET("/category", db.GetCategory)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3333"
	}

	http.ListenAndServe(":"+port, router)
}
