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

// GetMyProfile returns the current user's profile
func GetMyProfile(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user
	usersCollection := config.GetCollection("users")
	var user models.User
	err = usersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get user's courses count
	coursesCollection := config.GetCollection("courses")
	publishedCount, _ := coursesCollection.CountDocuments(ctx, bson.M{
		"author_id": userID,
		"status":    models.CourseStatusPublished,
	})
	draftCount, _ := coursesCollection.CountDocuments(ctx, bson.M{
		"author_id": userID,
		"status":    models.CourseStatusDraft,
	})

	// Get saved courses count
	savedCoursesCollection := config.GetCollection("saved_courses")
	savedCount, _ := savedCoursesCollection.CountDocuments(ctx, bson.M{"user_id": userID})

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":                   user.ID.Hex(),
			"name":                 user.Name,
			"email":                user.Email,
			"profile_image":        user.ProfileImage,
			"bio":                  user.Bio,
			"is_verified":          user.IsVerified,
			"has_published_course": user.HasPublishedCourse,
			"created_at":           user.CreatedAt,
		},
		"stats": gin.H{
			"published_courses": publishedCount,
			"draft_courses":     draftCount,
			"saved_courses":     savedCount,
		},
	})
}

// UpdateProfile updates the current user's profile
func UpdateProfile(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input models.UserUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build update document
	update := bson.M{"updated_at": time.Now()}
	if input.Name != "" {
		update["name"] = input.Name
	}
	if input.ProfileImage != "" {
		update["profile_image"] = input.ProfileImage
	}
	if input.Bio != "" {
		update["bio"] = input.Bio
	}

	collection := config.GetCollection("users")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// GetMyCourses returns the current user's courses (both published and drafts)
func GetMyCourses(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("courses")
	findOptions := options.Find().SetSort(bson.D{{"created_at", -1}})
	
	cursor, err := collection.Find(ctx, bson.M{"author_id": userID}, findOptions)
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

	// Separate published and drafts
	var published, drafts []models.Course
	for _, course := range courses {
		if course.Status == models.CourseStatusPublished {
			published = append(published, course)
		} else {
			drafts = append(drafts, course)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"published": published,
		"drafts":    drafts,
	})
}

// GetSavedCourses returns courses saved by the current user
func GetSavedCourses(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get saved course IDs
	savedCoursesCollection := config.GetCollection("saved_courses")
	cursor, err := savedCoursesCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching saved courses"})
		return
	}
	defer cursor.Close(ctx)

	var savedCourses []models.SavedCourse
	if err = cursor.All(ctx, &savedCourses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding saved courses"})
		return
	}

	// Get course IDs
	var courseIDs []primitive.ObjectID
	for _, saved := range savedCourses {
		courseIDs = append(courseIDs, saved.CourseID)
	}

	if len(courseIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"courses": []models.Course{}})
		return
	}

	// Get courses
	coursesCollection := config.GetCollection("courses")
	cursor, err = coursesCollection.Find(ctx, bson.M{"_id": bson.M{"$in": courseIDs}})
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

	// Get author info for each course
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
	})
}
