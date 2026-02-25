package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User representa a un usuario en el sistema
type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name               string             `bson:"name" json:"name" binding:"required"`
	Email              string             `bson:"email" json:"email" binding:"required,email"`
	PasswordHash       string             `bson:"password_hash" json:"-"`
	ProfileImage       string             `bson:"profile_image" json:"profile_image"`
	Bio                string             `bson:"bio" json:"bio"`
	INEUrl             string             `bson:"ine_url" json:"ine_url" binding:"required"`
	IsVerified         bool               `bson:"is_verified" json:"is_verified"` // Se verifica tras validar el INE
	HasPublishedCourse bool               `bson:"has_published_course" json:"has_published_course"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
	University         string             `bson:"university" json:"university"`
}

// UserRegisterInput representa los datos para el registro de usuario
type UserRegisterInput struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	INEUrl     string `json:"ine_url" binding:"required"`
	University string `json:"university"`
}

// UserLoginInput representa los datos para el inicio de sesión
type UserLoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserUpdateInput representa los datos para actualizar el perfil
type UserUpdateInput struct {
	Name         string `json:"name"`
	ProfileImage string `json:"profile_image"`
	Bio          string `json:"bio"`
	University   string `json:"university"`
}

// PasswordRecoveryInput representa la solicitud de recuperación de contraseña
type PasswordRecoveryInput struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetInput representa el cambio de contraseña con token
type PasswordResetInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
