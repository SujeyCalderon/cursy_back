package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"cursy_back/config"
	"cursy_back/models"
	"cursy_back/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Register maneja el registro de nuevos usuarios
func Register(c *gin.Context) {
	var input models.UserRegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verificar si el email ya existe en la DB
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	// Crear objeto de usuario con valores por defecto
	user := models.User{
		ID:                 primitive.NewObjectID(),
		Name:               input.Name,
		Email:              input.Email,
		PasswordHash:       hashedPassword,
		INEUrl:             input.INEUrl,
		IsVerified:         false, 
		HasPublishedCourse: false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	// Generar token JWT de acceso
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

// Login maneja la autenticación de usuarios existentes
func Login(c *gin.Context) {
	var input models.UserLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Buscar usuario por email en la DB
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Validar contraseña
	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generar token JWT
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

// RecoverPassword maneja la solicitud de recuperación de contraseña
func RecoverPassword(c *gin.Context) {
	var input models.PasswordRecoveryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verificar existencia (no revela si el email existe por seguridad)
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a recovery link has been sent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a recovery link has been sent",
	})
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// DeleteAccount maneja la eliminación total de la cuenta y sus datos
func DeleteAccount(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Borrar cursos publicados por el usuario
	coursesCollection := config.GetCollection("courses")
	_, err = coursesCollection.DeleteMany(ctx, bson.M{"author_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user's courses"})
		return
	}

	// Borrar registros de cursos guardados por el usuario
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

// UpdateFCMToken actualiza el token de Firebase del usuario actual
func UpdateFCMToken(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	var input struct {
		FCMToken string `json:"fcm_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FCM Token es requerido"})
		return
	}

	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Actualizando FCM Token para usuario %s: [%s...]", userID.Hex(), input.FCMToken[:10])
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"fcm_token": input.FCMToken, "updated_at": time.Now()}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar el token de FCM"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "FCM Token actualizado correctamente"})
}
