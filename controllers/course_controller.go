package controllers

import (
	"context"
	"log"
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

// CreateCourse crea un nuevo curso en estado borrador
func CreateCourse(c *gin.Context) {
	var input models.CourseCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Asignar orden a los bloques si no se proporcionó
	for i := range input.Blocks {
		if input.Blocks[i].Order == 0 {
			input.Blocks[i].Order = i + 1
		}
	}

	course := models.Course{
		ID:          primitive.NewObjectID(),
		AuthorID:    userID,
		Title:       input.Title,
		Description: input.Description,
		CoverImage:  input.CoverImage,
		Status:      models.CourseStatusDraft,
		Blocks:      input.Blocks,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	collection := config.GetCollection("courses")
	_, err = collection.InsertOne(ctx, course)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating course"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Course created as draft",
		"course":  course,
	})
}

// UpdateCourse actualiza un curso existente
func UpdateCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := primitive.ObjectIDFromHex(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input models.CourseUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("courses")

	// Verificar si el curso existe y pertenece al usuario
	var course models.Course
	err = collection.FindOne(ctx, bson.M{"_id": courseID, "author_id": userID}).Decode(&course)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found or you don't have permission to edit it"})
		return
	}

	// Construir documento de actualización
	update := bson.M{"updated_at": time.Now()}
	if input.Title != "" {
		update["title"] = input.Title
	}
	if input.Description != "" {
		update["description"] = input.Description
	}
	if input.CoverImage != "" {
		update["cover_image"] = input.CoverImage
	}
	if input.Blocks != nil {
		// Asignar orden a los bloques
		for i := range input.Blocks {
			if input.Blocks[i].Order == 0 {
				input.Blocks[i].Order = i + 1
			}
		}
		update["blocks"] = input.Blocks
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": courseID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating course"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Course updated successfully",
	})
}

// PublishCourse publica un curso que estaba en borrador
func PublishCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := primitive.ObjectIDFromHex(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coursesCollection := config.GetCollection("courses")

	// Verificar si el curso existe y pertenece al usuario
	var course models.Course
	err = coursesCollection.FindOne(ctx, bson.M{"_id": courseID, "author_id": userID}).Decode(&course)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found or you don't have permission to publish it"})
		return
	}

	// Validar que el curso tenga contenido
	if course.Title == "" || len(course.Blocks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Course must have a title and at least one content block"})
		return
	}

	// Cambiar estado a publicado
	_, err = coursesCollection.UpdateOne(ctx, bson.M{"_id": courseID}, bson.M{
		"$set": bson.M{
			"status":     models.CourseStatusPublished,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error publishing course"})
		return
	}

	// Marcar que el usuario ya ha publicado un curso (Trueque)
	usersCollection := config.GetCollection("users")
	_, err = usersCollection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$set": bson.M{
			"has_published_course": true,
			"updated_at":           time.Now(),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Course published successfully! You can now access other users' courses.",
	})

	// Notificar a todos los demás usuarios sobre el nuevo curso
	go func(authorID primitive.ObjectID, courseTitle string) {
		log.Printf("Buscando usuarios para notificar nuevo curso (excluyendo autor %s)...", authorID.Hex())
		cursor, err := usersCollection.Find(context.Background(), bson.M{
			"_id":       bson.M{"$ne": authorID},
			"fcm_token": bson.M{"$ne": ""},
		})
		if err != nil {
			log.Printf("Error buscando usuarios con FCM Token: %v", err)
			return
		}
		defer cursor.Close(context.Background())

		var users []models.User
		if err = cursor.All(context.Background(), &users); err != nil {
			log.Printf("Error al decodificar usuarios para notif curso: %v", err)
			return
		}

		log.Printf("Se encontraron %d usuarios con FCM Token para notificar", len(users))
		notificationBody := "Se ha publicado un nuevo curso: " + courseTitle
		for _, user := range users {
			data := map[string]string{
				"type":      "new_course",
				"target_id": courseID.Hex(),
			}
			tokenDisplay := user.FCMToken
			if len(tokenDisplay) > 10 {
				tokenDisplay = tokenDisplay[:10] + "..."
			}
			log.Printf("Enviando notificación a usuario %s (Token: %s)", user.Name, tokenDisplay)
			err := services.SendPushNotification(user.FCMToken, "Nuevo curso en Cursy 🎓", notificationBody, data)
			if err != nil {
				log.Printf("Error enviando a %s: %v", user.Name, err)
			}
		}
	}(userID, course.Title)

	// Broadcast vía WebSocket para que el Feed se actualice en tiempo real
	MainHub.Broadcast(map[string]string{
		"type":    "new_course",
		"content": "A new course has been published: " + course.Title,
	})
}

// DeleteCourse elimina un curso
func DeleteCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := primitive.ObjectIDFromHex(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("courses")

	// Verificar si el curso existe y pertenece al usuario
	result, err := collection.DeleteOne(ctx, bson.M{"_id": courseID, "author_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting course"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found or you don't have permission to delete it"})
		return
	}

	// También eliminar de la biblioteca de otros usuarios que lo guardaron
	savedCoursesCollection := config.GetCollection("saved_courses")
	_, _ = savedCoursesCollection.DeleteMany(ctx, bson.M{"course_id": courseID})

	c.JSON(http.StatusOK, gin.H{
		"message": "Course deleted successfully",
	})
}

// GetFeed devuelve todos los cursos publicados
func GetFeed(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("courses")

	// Buscar todos los cursos publicados
	findOptions := options.Find().SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, bson.M{"status": models.CourseStatusPublished}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching courses"})
		return
	}
	defer cursor.Close(ctx)

	var courses []models.Course
	if err = cursor.All(ctx, &courses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding courses"})
		return
	}

	// Obtener información del autor para cada curso
	usersCollection := config.GetCollection("users")
	var coursesWithAuthors []models.CourseResponse
	for _, course := range courses {
		var author models.User
		err := usersCollection.FindOne(ctx, bson.M{"_id": course.AuthorID}).Decode(&author)

		courseResponse := models.CourseResponse{
			Course:      course,
			AuthorName:  "",
			AuthorImage: "",
		}

		if err == nil {
			courseResponse.AuthorName = author.Name
			courseResponse.AuthorImage = author.ProfileImage
		}

		coursesWithAuthors = append(coursesWithAuthors, courseResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"courses": coursesWithAuthors,
		"count":   len(coursesWithAuthors),
	})
}

// GetCourseDetail devuelve un curso con todos sus detalles
func GetCourseDetail(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := primitive.ObjectIDFromHex(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("courses")

	var course models.Course
	err = collection.FindOne(ctx, bson.M{"_id": courseID}).Decode(&course)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Obtener ID del usuario actual
	userIDStr, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))
	isOwner := course.AuthorID == userID

	// Si no es el autor y no está publicado, denegar acceso
	if course.Status != models.CourseStatusPublished && !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "This course is a draft and only the author can see it"})
		return
	}

	// Obtener info del autor
	usersCollection := config.GetCollection("users")
	var author models.User
	usersCollection.FindOne(ctx, bson.M{"_id": course.AuthorID}).Decode(&author)

	// Verificar si el usuario ya guardó este curso
	savedCoursesCollection := config.GetCollection("saved_courses")
	var savedCourse models.SavedCourse
	isSaved := savedCoursesCollection.FindOne(ctx, bson.M{
		"user_id":   userID,
		"course_id": courseID,
	}).Decode(&savedCourse) == nil

	c.JSON(http.StatusOK, gin.H{
		"course": models.CourseResponse{
			Course:      course,
			AuthorName:  author.Name,
			AuthorImage: author.ProfileImage,
		},
		"is_owner": isOwner,
		"is_saved": isSaved,
	})
}

// SaveCourse guarda un curso en la biblioteca del usuario
func SaveCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := primitive.ObjectIDFromHex(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verificar si el curso existe
	coursesCollection := config.GetCollection("courses")
	var course models.Course
	err = coursesCollection.FindOne(ctx, bson.M{"_id": courseID, "status": models.CourseStatusPublished}).Decode(&course)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Verificar si ya está guardado
	savedCoursesCollection := config.GetCollection("saved_courses")
	var existingSaved models.SavedCourse
	err = savedCoursesCollection.FindOne(ctx, bson.M{"user_id": userID, "course_id": courseID}).Decode(&existingSaved)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Course already saved to library"})
		return
	}

	// Save course
	savedCourse := models.SavedCourse{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		CourseID:  courseID,
		CreatedAt: time.Now(),
	}

	_, err = savedCoursesCollection.InsertOne(ctx, savedCourse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving course"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Course saved to library",
	})
}

// UnsaveCourse elimina un curso de la biblioteca del usuario
func UnsaveCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := primitive.ObjectIDFromHex(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	savedCoursesCollection := config.GetCollection("saved_courses")
	result, err := savedCoursesCollection.DeleteOne(ctx, bson.M{"user_id": userID, "course_id": courseID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing course from library"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found in library"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Course removed from library",
	})
}
