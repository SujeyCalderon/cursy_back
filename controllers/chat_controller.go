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

// CreateConversation crea un nuevo chat o devuelve uno existente entre dos usuarios
func CreateConversation(c *gin.Context) {
	currentUserIDStr, _ := c.Get("userID")
	currentUserID, _ := primitive.ObjectIDFromHex(currentUserIDStr.(string))

	var input struct {
		OtherUserID string `json:"other_user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario requerido"})
		return
	}

	otherUserID, err := primitive.ObjectIDFromHex(input.OtherUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("conversations")

	// Verificar si ya existe una conversación entre estos dos usuarios
	filter := bson.M{
		"participants": bson.M{"$all": []primitive.ObjectID{currentUserID, otherUserID}},
	}

	var existingConversation models.Conversation
	err = collection.FindOne(ctx, filter).Decode(&existingConversation)

	if err == nil {
		// Ya existe, devolverla
		c.JSON(http.StatusOK, existingConversation)
		return
	}

	// No existe, crear una nueva
	conversation := models.Conversation{
		ID:           primitive.NewObjectID(),
		Participants: []primitive.ObjectID{currentUserID, otherUserID},
		LastMessage:  "",
		UpdatedAt:    time.Now(),
	}

	_, err = collection.InsertOne(ctx, conversation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear la conversación"})
		return
	}

	c.JSON(http.StatusCreated, conversation)
}

// GetConversations devuelve todos los chats del usuario actual
func GetConversations(c *gin.Context) {
	currentUserIDStr, _ := c.Get("userID")
	currentUserID, _ := primitive.ObjectIDFromHex(currentUserIDStr.(string))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("conversations")
	usersCollection := config.GetCollection("users")

	filter := bson.M{
		"participants": bson.M{"$in": []primitive.ObjectID{currentUserID}},
	}

	findOptions := options.Find().SetSort(bson.D{primitive.E{Key: "updated_at", Value: -1}})
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener conversaciones"})
		return
	}
	defer cursor.Close(ctx)

	var responses []models.ConversationResponse
	for cursor.Next(ctx) {
		var conv models.Conversation
		cursor.Decode(&conv)

		// Encontrar al otro participante para mostrar su info
		var otherUserID primitive.ObjectID
		for _, p := range conv.Participants {
			if p != currentUserID {
				otherUserID = p
				break
			}
		}

		var otherUser models.User
		usersCollection.FindOne(ctx, bson.M{"_id": otherUserID}).Decode(&otherUser)

		responses = append(responses, models.ConversationResponse{
			ID:          conv.ID.Hex(),
			OtherUser:   otherUser,
			LastMessage: conv.LastMessage,
			UpdatedAt:   conv.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, responses)
}

// GetMessages devuelve el historial de mensajes de un chat específico (ID de la conversación)
func GetMessages(c *gin.Context) {
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de conversación inválido"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("messages")

	filter := bson.M{"conversation_id": conversationID}
	findOptions := options.Find().SetSort(bson.D{primitive.E{Key: "created_at", Value: 1}}) // Orden cronológico

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener mensajes"})
		return
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al decodificar mensajes"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
