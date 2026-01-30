package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SavedCourse represents a course saved by a user to their library
type SavedCourse struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CourseID  primitive.ObjectID `bson:"course_id" json:"course_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
