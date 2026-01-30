package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name               string             `bson:"name" json:"name" binding:"required"`
	Email              string             `bson:"email" json:"email" binding:"required,email"`
	PasswordHash       string             `bson:"password_hash" json:"-"`
	ProfileImage       string             `bson:"profile_image" json:"profile_image"`
	Bio                string             `bson:"bio" json:"bio"`
	INEUrl             string             `bson:"ine_url" json:"ine_url" binding:"required"`
	IsVerified         bool               `bson:"is_verified" json:"is_verified"`
	HasPublishedCourse bool               `bson:"has_published_course" json:"has_published_course"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

// UserRegisterInput represents the input for user registration
type UserRegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	INEUrl   string `json:"ine_url" binding:"required"`
}

// UserLoginInput represents the input for user login
type UserLoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserUpdateInput represents the input for updating user profile
type UserUpdateInput struct {
	Name         string `json:"name"`
	ProfileImage string `json:"profile_image"`
	Bio          string `json:"bio"`
}

// PasswordRecoveryInput represents the input for password recovery
type PasswordRecoveryInput struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetInput represents the input for password reset
type PasswordResetInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
