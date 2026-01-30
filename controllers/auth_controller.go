package controllers

import (
	"context"
	"net/http"
	"time"

	"cursy_back/config"
	"cursy_back/models"
	"cursy_back/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Register handles user registration
func Register(c *gin.Context) {
	var input models.UserRegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if email already exists
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	// Create user
	user := models.User{
		ID:                 primitive.NewObjectID(),
		Name:               input.Name,
		Email:              input.Email,
		PasswordHash:       hashedPassword,
		INEUrl:             input.INEUrl,
		IsVerified:         false, // Will be verified after INE validation
		HasPublishedCourse: false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":    user.ID.Hex(),
			"name":  user.Name,
			"email": user.Email,
		},
		"token": token,
	})
}

// Login handles user login
func Login(c *gin.Context) {
	var input models.UserLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user by email
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check password
	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":                   user.ID.Hex(),
			"name":                 user.Name,
			"email":                user.Email,
			"profile_image":        user.ProfileImage,
			"has_published_course": user.HasPublishedCourse,
		},
		"token": token,
	})
}

// RecoverPassword handles password recovery request
func RecoverPassword(c *gin.Context) {
	var input models.PasswordRecoveryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user exists
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		// Don't reveal if email exists or not for security
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a recovery link has been sent"})
		return
	}

	// TODO: In production, send email with recovery token
	// For now, just return success message
	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a recovery link has been sent",
	})
}

// Logout handles user logout (client-side token removal)
func Logout(c *gin.Context) {
	// In a more complete implementation, we would:
	// 1. Add the token to a blacklist in Redis/DB
	// 2. Set token expiration
	// For now, logout is handled client-side by removing the token
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// DeleteAccount handles account deletion
func DeleteAccount(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Delete user's courses
	coursesCollection := config.GetCollection("courses")
	_, err = coursesCollection.DeleteMany(ctx, bson.M{"author_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user's courses"})
		return
	}

	// Delete user's saved courses
	savedCoursesCollection := config.GetCollection("saved_courses")
	_, err = savedCoursesCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user's saved courses"})
		return
	}

	// Delete user
	usersCollection := config.GetCollection("users")
	_, err = usersCollection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account deleted successfully",
	})
}
