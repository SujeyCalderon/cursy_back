package routes

import (
	"cursy_back/controllers"
	"cursy_back/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configura todas las rutas del API
func SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/recover-password", controllers.RecoverPassword)
		}

		authProtected := api.Group("/auth")
		authProtected.Use(middleware.JWTAuth())
		{
			authProtected.POST("/logout", controllers.Logout)
			authProtected.DELETE("/account", controllers.DeleteAccount)
		}

		courses := api.Group("/courses")
		courses.Use(middleware.JWTAuth())
		{
			courses.POST("", controllers.CreateCourse)

			courses.GET("", controllers.GetFeed)

			courses.GET("/:id", middleware.RequirePublishedCourse(), controllers.GetCourseDetail)

			courses.PUT("/:id", controllers.UpdateCourse)
			courses.DELETE("/:id", controllers.DeleteCourse)

			courses.PUT("/:id/publish", controllers.PublishCourse)

			courses.POST("/:id/save", middleware.RequirePublishedCourse(), controllers.SaveCourse)
			courses.DELETE("/:id/save", controllers.UnsaveCourse)
		}

		profile := api.Group("/profile")
		profile.Use(middleware.JWTAuth())
		{
			profile.GET("", controllers.GetMyProfile)
			profile.PUT("", controllers.UpdateProfile)
			profile.GET("/courses", controllers.GetMyCourses)
			profile.GET("/saved", controllers.GetSavedCourses)
		}

		users := api.Group("/users")
		users.Use(middleware.JWTAuth())
		{
			users.GET("", controllers.GetUsers)
			users.GET("/online", controllers.GetOnlineUsers)
		}

		chats := api.Group("/chats")
		chats.Use(middleware.JWTAuth())
		{
			chats.GET("", controllers.GetConversations)
			chats.POST("", controllers.CreateConversation)
			chats.GET("/:id/messages", controllers.GetMessages)
		}

		// Conexión en tiempo real (WebSocket)
		api.GET("/ws", middleware.JWTAuth(), controllers.WSHandler)

		// Subida de archivos (protegidas)
		upload := api.Group("/upload")
		upload.Use(middleware.JWTAuth())
		{
			upload.POST("", controllers.UploadImage)
		}
	}
}
