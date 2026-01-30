package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CourseStatus represents the status of a course
type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "DRAFT"
	CourseStatusPublished CourseStatus = "PUBLISHED"
)

// ContentBlockType represents the type of content block
type ContentBlockType string

const (
	ContentBlockTypeText  ContentBlockType = "TEXT"
	ContentBlockTypeImage ContentBlockType = "IMAGE"
)

// ContentBlock represents a block of content in a course
type ContentBlock struct {
	Type    ContentBlockType `bson:"type" json:"type" binding:"required"`
	Content string           `bson:"content" json:"content" binding:"required"`
	Order   int              `bson:"order" json:"order"`
}

// Course represents a course/publication in the system
type Course struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AuthorID    primitive.ObjectID `bson:"author_id" json:"author_id"`
	Title       string             `bson:"title" json:"title" binding:"required"`
	Description string             `bson:"description" json:"description"`
	CoverImage  string             `bson:"cover_image" json:"cover_image"`
	Status      CourseStatus       `bson:"status" json:"status"`
	Blocks      []ContentBlock     `bson:"blocks" json:"blocks"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// CourseCreateInput represents the input for creating a course
type CourseCreateInput struct {
	Title       string         `json:"title" binding:"required"`
	Description string         `json:"description"`
	CoverImage  string         `json:"cover_image"`
	Blocks      []ContentBlock `json:"blocks"`
}

// CourseUpdateInput represents the input for updating a course
type CourseUpdateInput struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	CoverImage  string         `json:"cover_image"`
	Blocks      []ContentBlock `json:"blocks"`
}

// CourseResponse represents a course response with author info
type CourseResponse struct {
	Course
	AuthorName  string `json:"author_name"`
	AuthorImage string `json:"author_image"`
}
