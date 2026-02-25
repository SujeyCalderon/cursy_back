package controllers

import (
	"context"
	"net/http"
	"time"

	"cursy_back/config"
	"cursy_back/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetUsers(c *gin.Context) {
	// Obtener el ID del usuario actual para no mostrarse a sí mismo
	currentUserIDStr, _ := c.Get("userID")
	currentUserID, _ := primitive.ObjectIDFromHex(currentUserIDStr.(string))

	search := c.Query("q") 

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("users")

	// Filtro base: no mostrar al usuario actual
	filter := bson.M{
		"_id": bson.M{"$ne": currentUserID},
	}

	if search != "" {
		filter["name"] = bson.M{"$regex": search, "$options": "i"}
	}

	findOptions := options.Find().SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener usuarios"})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al decodificar usuarios"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}
