package db

import (
	"backend/auth"
	"backend/interfaces"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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

func GetUserPost(c *gin.Context) {
	// Get the user ID from the URL parameter
	userID := c.Param("id")

	// Query all posts for this user
	var posts []interfaces.Post
	if err := DB.Where("author_id = ?", userID).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user posts",
		})
		return
	}

	c.JSON(http.StatusOK, posts)
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

func Login(c *gin.Context) {
	var authUser interfaces.AuthenticateUser

	if err := c.BindJSON(&authUser); err != nil {
		c.JSON(400, gin.H{"error": "Invalid login data"})
		return
	}

	// Find user by username
	var user interfaces.User
	result := DB.Where("username = ?", authUser.Username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(authUser.PasswordHash))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create JWT token
	tokenString, err := auth.CreateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating token"})
		return
	}

	// Set secure cookie
	c.SetCookie(
		"token",     // name
		tokenString, // value
		60*60*24*30, // maxAge (30 days in seconds)
		"/",         // path
		"",          // domain (empty = current domain)
		true,        // secure
		true,        // httpOnly
	)

	// Also return token in response for client-side storage
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
		},
	})
}

func Logout(c *gin.Context) {
	// Clear the cookie by setting maxAge to -1
	c.SetCookie(
		"token", // name
		"",      // value
		-1,      // maxAge
		"/",     // path
		"",      // domain
		true,    // secure
		true,    // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// Add these functions to your db/db.go file

func SyncToken(c *gin.Context) {
	var tokenData struct {
		Token string `json:"token"`
	}

	if err := c.BindJSON(&tokenData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token data"})
		return
	}

	// Verify the token is valid
	_, err := auth.VerifyToken(tokenData.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Set the cookie with the token
	c.SetCookie(
		"token",
		tokenData.Token,
		60*60*24*30, // 30 days
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "Token synchronized successfully"})
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req interfaces.UpdateUserRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	// Only update specific fields
	result := DB.Model(&interfaces.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"username":   req.Username,
			"email":      req.Email,
			"avatar_url": req.AvatarURL,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
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
	pageIndex := c.Param("pageIndex")

	// Convert pageIndex string to int
	page, err := strconv.Atoi(pageIndex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page index"})
		return
	}

	// Define items per page
	const itemsPerPage = 10
	offset := page * itemsPerPage

	var categoryRecord interfaces.Category
	if err := DB.Where("name = ?", category).First(&categoryRecord).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var posts []interfaces.Post
	if err := DB.Order("created_at DESC").
		Where("category_id = ?", categoryRecord.ID).
		Limit(itemsPerPage). // Limit to 10 items
		Offset(offset).      // Skip previous pages
		Find(&posts).Error; err != nil {
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

	var req interfaces.UpdatePostRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post data"})
		return
	}

	// Only update specific fields
	result := DB.Model(&interfaces.Post{}).
		Where("id = ?", postID).
		Updates(map[string]interface{}{
			"title":   req.Title,
			"content": req.Content,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

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

// Tag handlers
func ListTags(c *gin.Context) {
	var tags []interfaces.Tag

	if err := DB.Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving tags"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

func GetTag(c *gin.Context) {
	id := c.Param("id")

	var tag interfaces.Tag
	result := DB.First(&tag, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	c.JSON(http.StatusOK, tag)
}

func CreatePostTag(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var tags []interfaces.Tag
	if err := c.BindJSON(&tags); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag data"})
		return
	}

	var post interfaces.Post
	if err := DB.First(&post, "id = ?", postID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching post"})
		return
	}

	for _, tag := range tags {
		var tagRecord interfaces.Tag
		if err := DB.First(&tagRecord, "name = ?", tag.Name).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding tag"})
			return
		}

		// Create the PostTag record only if it doesn't exist
		postTag := interfaces.PostTag{
			PostID: post.ID,
			TagID:  tagRecord.ID,
		}

		if err := DB.Create(&postTag).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error creating post tag: %v", err)})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tags added to post successfully"})
}

func ListPostTags(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var tags []interfaces.Tag

	if result := DB.Joins("JOIN posts_tags ON posts_tags.tag_id = tags.id").
		Where("posts_tags.post_id = ?", postID).
		Find(&tags); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving tags"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

func DeletePostTag(c *gin.Context) {
	postID, _ := uuid.Parse(c.Param("id"))
	tagID, _ := uuid.Parse(c.Param("tag_id"))

	postTag := interfaces.PostTag{
		PostID: postID,
		TagID:  tagID,
	}

	result := DB.Delete(&postTag)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found on post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tags removed from post successfully"})
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

func CreateComment(c *gin.Context) {
	postID := c.Param("id")
	var newComment interfaces.CommentInput

	// Check if post exists
	var post interfaces.Post
	if err := DB.First(&post, "id = ?", postID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching post"})
		return
	}

	if err := c.BindJSON(&newComment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment data"})
		return
	}

	comment := interfaces.Comment{
		Content:  newComment.Content,
		AuthorID: newComment.AuthorID,
		PostID:   post.ID,
	}

	if err := DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
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

func GetCategory(c *gin.Context) {
	// Get the ID parameter
	id := c.Param("id")

	var category interfaces.Category

	result := DB.First(&category, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, category)
}
