package middleware

import (
	"net/http"
	"strings"

	"cursy_back/config"
	"cursy_back/models"
	"cursy_back/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JWTAuth middleware validates JWT tokens
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// RequirePublishedCourse middleware checks if user has published a course (trueque logic)
func RequirePublishedCourse() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// Get user from database
		collection := config.GetCollection("users")
		var user models.User
		err = collection.FindOne(c.Request.Context(), bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Check if user has published a course
		if !user.HasPublishedCourse {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "You must publish a course first to access other courses",
				"message": "Para ver cursos de otros usuarios, primero debes publicar tu propio curso",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
