package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Image represents an image stored in external storage (iDrive)
type Image struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	URL       string             `bson:"url" json:"url" binding:"required"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CourseID  primitive.ObjectID `bson:"course_id" json:"course_id,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// ImageUploadInput represents the input for uploading an image
type ImageUploadInput struct {
	URL      string `json:"url" binding:"required"`
	CourseID string `json:"course_id"`
}
