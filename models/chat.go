package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Conversation representa un chat entre dos o más usuarios
type Conversation struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Participants []primitive.ObjectID `bson:"participants" json:"participants"` // Lista de IDs de los usuarios en el chat
	LastMessage  string               `bson:"last_message" json:"last_message"`
	UpdatedAt    time.Time            `bson:"updated_at" json:"updated_at"`
}

// Message representa un mensaje individual dentro de una conversación
type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ConversationID primitive.ObjectID `bson:"conversation_id" json:"conversation_id"`
	SenderID       primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	Content        string             `bson:"content" json:"content"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// ConversationResponse representa un chat con info básica para la lista
type ConversationResponse struct {
	ID          string    `json:"id"`
	OtherUser   User      `json:"other_user"`
	LastMessage string    `json:"last_message"`
	UpdatedAt   time.Time `json:"updated_at"`
}
