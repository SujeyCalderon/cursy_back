package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func ConnectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}

	// Verificar la conexión con la base de datos
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB: ", err)
	}

	dbName := getEnv("DB_NAME", "cursy_db")
	DB = client.Database(dbName)
	log.Println("Connected to MongoDB successfully!")
}

// GetCollection obtiene una colección de MongoDB
func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}
