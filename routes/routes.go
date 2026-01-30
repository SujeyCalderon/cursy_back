package routes

import (
	"cursy_back/controllers"
	"cursy_back/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	api := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/recover-password", controllers.RecoverPassword)
		}

		// Protected auth routes
		authProtected := api.Group("/auth")
		authProtected.Use(middleware.JWTAuth())
		{
			authProtected.POST("/logout", controllers.Logout)
			authProtected.DELETE("/account", controllers.DeleteAccount)
		}

		// Course routes (protected)
		courses := api.Group("/courses")
		courses.Use(middleware.JWTAuth())
		{
			// Create course (doesn't require published course)
			courses.POST("", controllers.CreateCourse)

			// Get feed (doesn't require published course - just shows the cards)
			courses.GET("", controllers.GetFeed)

			// Course detail - REQUIRES published course (trueque logic)
			courses.GET("/:id", middleware.RequirePublishedCourse(), controllers.GetCourseDetail)

			// Update and delete own courses
			courses.PUT("/:id", controllers.UpdateCourse)
			courses.DELETE("/:id", controllers.DeleteCourse)

			// Publish course
			courses.PUT("/:id/publish", controllers.PublishCourse)

			// Save/unsave course to library - REQUIRES published course
			courses.POST("/:id/save", middleware.RequirePublishedCourse(), controllers.SaveCourse)
			courses.DELETE("/:id/save", controllers.UnsaveCourse)
		}

		// Profile routes (protected)
		profile := api.Group("/profile")
		profile.Use(middleware.JWTAuth())
		{
			profile.GET("", controllers.GetMyProfile)
			profile.PUT("", controllers.UpdateProfile)
			profile.GET("/courses", controllers.GetMyCourses)
			profile.GET("/saved", controllers.GetSavedCourses)
		}

		// Upload routes (protected)
		upload := api.Group("/upload")
		upload.Use(middleware.JWTAuth())
		{
			upload.POST("", controllers.UploadImage)
		}
	}
}
