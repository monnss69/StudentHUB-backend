package db

import (
	"backend/auth"
	"backend/interfaces"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Initialize initializes the database connection
func Initialize() {
	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found. Ensure environment variables are set.")
	}

	// Get database connection URL from the environment
	databaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("Error: SUPABASE_DATABASE_URL not set in environment")
	}

	// Connect to the database using GORM
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Successfully connected to the database!")

	// Set up connection pool settings
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
}

// User handlers
func CreateUser(c *gin.Context) {
	var newUser interfaces.User

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing password"})
		return
	}

	newUser.PasswordHash = string(hashedPassword)

	if err := DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	newUser.PasswordHash = "" // Don't send password hash back
	c.JSON(http.StatusCreated, newUser)
}

func ListUsers(c *gin.Context) {
	var users []interfaces.User
	username := c.Query("username")

	query := DB
	if username != "" {
		query = query.Where("username = ?", username)
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user interfaces.User

	if err := DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	result := DB.Delete(&interfaces.User{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func Logout(c *gin.Context) {
	cookie := &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Domain:   "studenthub-backend.vercel.app",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}

	// Clear the cookie
	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

func Login(c *gin.Context) {
	var authUser interfaces.AuthenticateUser

	if err := c.BindJSON(&authUser); err != nil {
		c.JSON(400, gin.H{"error": "Error binding JSON"})
		return
	}

	// First find the user by username
	var user interfaces.User
	result := DB.Where("username = ?", authUser.Username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong username/password"})
		return
	}

	// Compare the provided password with stored hash
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(authUser.PasswordHash))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong username/password"})
		return
	} else {
		tokenString, err := auth.CreateToken(user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating JWT token"})
			return
		}

		cookie := &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Path:     "/",
			Domain:   "studenthub-backend.vercel.app",
			MaxAge:   3600,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		}

		// Set the cookie
		http.SetCookie(c.Writer, cookie)
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}

// Post handlers
func CreatePost(c *gin.Context) {
	var post interfaces.Post

	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post data"})
		return
	}

	if err := DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating post"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func GetPost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post interfaces.Post
	if err := DB.First(&post, "id = ?", postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func ListPostsByCategory(c *gin.Context) {
	category := c.Param("category")

	var categoryRecord interfaces.Category
	if err := DB.Where("name = ?", category).First(&categoryRecord).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var posts []interfaces.Post
	if err := DB.Where("category_id = ?", categoryRecord.ID).Find(&posts).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No posts found in this category"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func UpdatePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var updatedPost interfaces.Post
	if err := c.BindJSON(&updatedPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post data"})
		return
	}

	result := DB.Model(&interfaces.Post{}).Where("id = ?", postID).Updates(&updatedPost)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post updated successfully"})
}

func DeletePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	result := DB.Delete(&interfaces.Post{}, postID)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// Comment handlers
func ListPostComments(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var comments []interfaces.Comment
	if err := DB.Where("post_id = ?", postID).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// Category handlers
func ListCategories(c *gin.Context) {
	var categories []interfaces.Category

	if err := DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}
