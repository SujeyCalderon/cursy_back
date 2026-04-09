package controllers

import (
	"context"
	"net/http"
	"time"

	"cursy_back/config"
	"cursy_back/models"
	"cursy_back/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateComment permite a un usuario autenticado dejar un comentario en un curso
func CreateComment(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := primitive.ObjectIDFromHex(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de curso inválido"})
		return
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido en el token"})
		return
	}

	var input models.CommentCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El comentario no puede estar vacío"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verificar si el curso existe y está publicado
	coursesCollection := config.GetCollection("courses")
	var course models.Course
	err = coursesCollection.FindOne(ctx, bson.M{"_id": courseID, "status": models.CourseStatusPublished}).Decode(&course)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "El curso no existe o no está publicado"})
		return
	}

	comment := models.Comment{
		ID:        primitive.NewObjectID(),
		CourseID:  courseID,
		UserID:    userID,
		Content:   input.Content,
		CreatedAt: time.Now().Truncate(time.Second),
	}

	commentsCollection := config.GetCollection("comments")
	_, err = commentsCollection.InsertOne(ctx, comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el comentario"})
		return
	}

	// Buscar datos del autor para devolver la respuesta completa de inmediato
	usersCollection := config.GetCollection("users")
	var author models.User
	err = usersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&author)

	commentResponse := models.CommentResponse{
		Comment:   comment,
		UserName:  "Usuario Desconocido",
		UserImage: "",
	}

	if err == nil {
		commentResponse.UserName = author.Name
		commentResponse.UserImage = author.ProfileImage
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Comentario publicado exitosamente",
		"comment": commentResponse,
	})

	// Notificar al autor del curso sobre el nuevo comentario
	go func(courseAuthorID, commenterID primitive.ObjectID, courseTitle, commenterName, content string) {
		// No notificar si el autor del curso es el mismo que comenta
		if courseAuthorID == commenterID {
			return
		}

		usersCollection := config.GetCollection("users")
		var author models.User
		err := usersCollection.FindOne(context.Background(), bson.M{"_id": courseAuthorID}).Decode(&author)
		
		if err == nil && author.FCMToken != "" {
			title := "Nuevo comentario en tu curso 💬"
			body := commenterName + " comentó en '" + courseTitle + "': " + content
			services.SendPushNotification(author.FCMToken, title, body)
		}
	}(course.AuthorID, userID, course.Title, author.Name, input.Content)
}

// GetCommentsByCourse devuelve la lista de comentarios de un curso con el autor
func GetCommentsByCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := primitive.ObjectIDFromHex(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de curso inválido"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	commentsCollection := config.GetCollection("comments")
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := commentsCollection.Find(ctx, bson.M{"course_id": courseID}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar los comentarios"})
		return
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al decodificar los comentarios"})
		return
	}

	// Lookup manual: Buscar la info del usuario para cada comentario
	usersCollection := config.GetCollection("users")
	var commentsWithAuthors []models.CommentResponse

	for _, comment := range comments {
		var author models.User
		err := usersCollection.FindOne(ctx, bson.M{"_id": comment.UserID}).Decode(&author)

		commentResponse := models.CommentResponse{
			Comment:   comment,
			UserName:  "Usuario Desconocido",
			UserImage: "",
		}

		if err == nil {
			commentResponse.UserName = author.Name
			commentResponse.UserImage = author.ProfileImage
		}

		commentsWithAuthors = append(commentsWithAuthors, commentResponse)
	}

	if commentsWithAuthors == nil {
		commentsWithAuthors = []models.CommentResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": commentsWithAuthors,
		"count":    len(commentsWithAuthors),
	})
}
