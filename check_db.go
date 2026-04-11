package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	uri := "mongodb+srv://calderonmartinezsujey_db_user:3llCkOyW3E9F6ysN@cursy.4gfq57d.mongodb.net/"
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("cursy_db")
	usersColl := db.Collection("users")

	cursor, err := usersColl.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	fmt.Println("Usuarios en cursy_db.users:")
	for cursor.Next(ctx) {
		var user bson.M
		if err := cursor.Decode(&user); err != nil {
			log.Fatal(err)
		}
		name := user["name"]
		token := user["fcm_token"]
		email := user["email"]
		fmt.Printf("Name: %v, Email: %v, FCM Token: [%v]\n", name, email, token)
	}
}
