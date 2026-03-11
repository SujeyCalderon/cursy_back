package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment representa un comentario de un usuario en un curso específico
type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CourseID  primitive.ObjectID `bson:"course_id" json:"course_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Content   string             `bson:"content" json:"content" binding:"required"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// CommentCreateInput son los datos que manda Android para crear el comentario
type CommentCreateInput struct {
	Content string `json:"content" binding:"required"`
}

// para decirle quién escribió el comentario y mostrar su foto.
type CommentResponse struct {
	Comment
	UserName  string `json:"user_name"`
	UserImage string `json:"user_image"`
}
