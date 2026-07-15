package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbName = "netflix"

var Client *mongo.Client // ← exported, capital C

func ConnectDB() {
	// read from .env file, not hardcoded
	mongoURL := os.Getenv("MONGO_URL")

	clientOptions := options.Client().ApplyURI(mongoURL)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// ping to confirm connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	fmt.Println("MongoDB connected successfully")
	Client = client // store globally so GetCollection can use it
}

func GetCollection(collectionName string) *mongo.Collection {
	return Client.Database(dbName).Collection(collectionName)
}

func CreateTextIndex() {
	collection := GetCollection("allShows")

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "description", Value: "text"},
			{Key: "cast", Value: "text"},
			{Key: "genre", Value: "text"},
			{Key: "director", Value: "text"},
		},
		Options: options.Index().SetName("content_text_index"),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Println("Index creation failed:", err)
		return
	}
	log.Println("Text index created successfully")
}
