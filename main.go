package main

import (
	"log"

	"cursy_back/config"
	"cursy_back/controllers"
	"cursy_back/routes"
	"cursy_back/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Cargar archivo .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

func main() {
	// Conexión a MongoDB
	config.ConnectDB()

	// Inicialización de Firebase
	services.InitFirebase()

	// Crear router de Gin
	router := gin.Default()

	// Habilitar CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Manejo de preflight (OPTIONS)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Configurar rutas
	routes.SetupRoutes(router)

	// Iniciar Hub de WebSockets
	go controllers.MainHub.Run()

	// Servir archivos estáticos (subidas locales)
	router.Static("/uploads", "./uploads")

	// Iniciar servidor HTTP en puerto 8080
	log.Println("=== CURSY VERSION 2.0 STARTING ===")
	log.Println("Server starting on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
