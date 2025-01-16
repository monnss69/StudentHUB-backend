package auth

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var secretKey string

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: No .env file found")
	}

	secretKey = os.Getenv("JWT_SECRET")
	if secretKey == "" {
		panic("JWT_SECRET is not set")
	}
}

// CreateToken generates a JWT token with a longer expiration time
func CreateToken(username string) (string, error) {
	// Create token with 30-day expiration
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days
		"iat": time.Now().Unix(),
	})

	tokenString, err := claims.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// verifyToken verifies and parses a JWT token
func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

// AuthMiddleware verifies the JWT token from cookies
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First try to get token from cookie
		tokenString, err := c.Cookie("token")
		if err != nil {
			// If no cookie, try Authorization header
			authHeader := c.GetHeader("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid authentication token found"})
				c.Abort()
				return
			}
		}

		// Verify the token
		token, err := verifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract and validate claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if username, ok := claims["sub"].(string); ok {
				c.Set("username", username)
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
	}
}
