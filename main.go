package handler

import (
	"backend/auth"
	"backend/db"
	"net/http"
	"os"

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

	// Simple CORS Configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://student-hub-frontend.vercel.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

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
