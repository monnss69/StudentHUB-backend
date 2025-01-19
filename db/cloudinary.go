package db

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// CloudinaryService holds the Cloudinary client configuration
type CloudinaryService struct {
	Cld *cloudinary.Cloudinary
}

// NewCloudinaryService creates a new Cloudinary service instance
func NewCloudinaryService() (*CloudinaryService, error) {

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found. Ensure environment variables are set.")
	}

	// Get Cloudinary credentials from environment variables (this is set in Vercel environment variables)
	cloudName := os.Getenv("CLOUD_NAME")
	apiKey := os.Getenv("CLOUD_API_KEY")
	apiSecret := os.Getenv("CLOUD_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("missing required Cloudinary environment variables")
	}

	// Create Cloudinary instance
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	return &CloudinaryService{Cld: cld}, nil
}

// UploadImage handles the image upload to Cloudinary
func (s *CloudinaryService) UploadImage(file *multipart.FileHeader, username string) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer src.Close()

	// Sanitize username for use in public_id for access from the URL stored in Supabase
	sanitizedUsername := strings.ReplaceAll(username, " ", "_")
	publicID := fmt.Sprintf("avatars/%s_avatar", sanitizedUsername)

	// Set upload parameters
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		Folder:       "avatars",
		ResourceType: "auto",
		// Add transformations for image optimization (size, quality, format)
		Transformation: "w_200,h_200,c_fill,q_auto,f_auto",
	}

	// Upload the file to Cloudinary
	uploadResult, err := s.Cld.Upload.Upload(context.Background(), src, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %v", err)
	}

	return uploadResult.SecureURL, nil
}

// DeleteImage removes an image from Cloudinary
func (s *CloudinaryService) DeleteImage(username string) error {
	publicID := fmt.Sprintf("avatars/%s_avatar", username)

	// Delete the image
	_, err := s.Cld.Upload.Destroy(context.Background(), uploader.DestroyParams{
		PublicID: publicID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete image: %v", err)
	}

	return nil
}

// UploadHandler handles the HTTP request for image upload
func UploadHandler(c *gin.Context) {
	// Get the Cloudinary service instance
	cloudinaryService, err := NewCloudinaryService()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to initialize Cloudinary: %v", err)})
		return
	}

	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}

	// Get username from form data
	username := c.PostForm("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "Username is required"})
		return
	}

	// Check file size (e.g., 5MB limit)
	if file.Size > 5*1024*1024 {
		c.JSON(400, gin.H{"error": "File size too large (max 5MB)"})
		return
	}

	// Check file type
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		c.JSON(400, gin.H{"error": "File must be an image"})
		return
	}

	// Upload the image
	imageURL, err := cloudinaryService.UploadImage(file, username)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to upload image: %v", err)})
		return
	}

	c.JSON(200, gin.H{"url": imageURL})
}

// DeleteImageHandler handles the HTTP request for image deletion (same error handling as above)
func DeleteImageHandler(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "Username is required"})
		return
	}

	cloudinaryService, err := NewCloudinaryService()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to initialize Cloudinary: %v", err)})
		return
	}

	if err := cloudinaryService.DeleteImage(username); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete image: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": "Image deleted successfully"})
}
