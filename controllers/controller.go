package controllers

import (
	"Netflix/config"
	"Netflix/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// get collection from config instead of local connection
func getCollection() *mongo.Collection {
	return config.GetCollection("watchlist")
}

// ─── DB Helpers (unexported) ──────────────────────────────────────────────────

func insertOneMovie(movie model.NetFlix) {
	fmt.Println(movie)
	inserted, err := getCollection().InsertOne(context.Background(), movie)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted movie:", inserted.InsertedID)
}

func updateOneMovie(movieID string) {
	id, _ := primitive.ObjectIDFromHex(movieID)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"watched": true}} // ✅ lowercase $set

	result, err := getCollection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated count:", result.ModifiedCount)
}

func getAllMovies() []primitive.M {
	cursor, err := getCollection().Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	var movies []primitive.M
	for cursor.Next(context.Background()) {
		var movie bson.M
		if err := cursor.Decode(&movie); err != nil {
			log.Fatal(err)
		}
		movies = append(movies, movie)
	}
	return movies
}

func deleteOneMovie(movieID string) {
	id, _ := primitive.ObjectIDFromHex(movieID)
	filter := bson.M{"_id": id}
	count, err := getCollection().DeleteOne(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted count:", count.DeletedCount)
}

// ─── HTTP Handlers (exported) ─────────────────────────────────────────────────

func GetAllMovies(w http.ResponseWriter, r *http.Request) {
	movies := getAllMovies()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

func CreateMovie(w http.ResponseWriter, r *http.Request) {
	var movie model.NetFlix
	json.NewDecoder(r.Body).Decode(&movie)
	insertOneMovie(movie)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func MarkAsWatched(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r) // you'll need to import gorilla/mux here too
	updateOneMovie(params["id"])
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(params["id"])
}

func DeleteMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deleteOneMovie(params["id"])
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(params["id"])
}
