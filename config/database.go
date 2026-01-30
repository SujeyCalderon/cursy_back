package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

const (
	DatabaseName = "cursy_db"
	MongoURI     = "mongodb+srv://calderonmartinezsujey_db_user:3llCkOyW3E9F6ysN@cursy.4gfq57d.mongodb.net/"
)

func ConnectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB: ", err)
	}

	DB = client.Database(DatabaseName)
	log.Println("Connected to MongoDB successfully!")
}

// GetCollection returns a MongoDB collection
func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}
