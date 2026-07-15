package controllers

import (
	"Netflix/config"
	helper "Netflix/helpers"
	"fmt"
	"log"

	"Netflix/model"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getContentCollection() *mongo.Collection {
	return config.GetCollection("allShows")
}

func CreateContent(w http.ResponseWriter, r *http.Request) {
	var input model.CreateContentInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helper.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// basic validation
	if input.Title == "" || input.ContentType == "" {
		helper.SendError(w, http.StatusBadRequest, "Title and type are required")
		return
	}

	// build full content struct
	content := model.Content{
		ID:             primitive.NewObjectID(),
		Title:          input.Title,
		Description:    input.Description,
		ContentType:    input.ContentType,
		ReleaseYear:    input.ReleaseYear,
		MaturityRating: input.MaturityRating,
		Language:       input.Languages,
		ThumbnailUrl:   input.ThumbnailUrl,
		VideoUrl:       input.VideoUrl,
		TrailerUrl:     input.TrailerUrl,
		Genre:          input.Genre,
		Cast:           input.Cast,
		Director:       input.Director,
		TotalSeasons:   input.TotalSeasons,
		IsFeatured:     input.IsFeatured,
		TotalViews:     0,
		AverageRating:  0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = getContentCollection().InsertOne(context.Background(), content)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Error creating content")
		return
	}

	helper.SendSuccess(w, http.StatusCreated, "Content created successfully", content)
}

func GetAllContent(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cacheKey := "content:all"
	redisData, err := config.RedisClient.Get(ctx, cacheKey).Result()

	if err == nil {
		// cache hit → return Redis data directly
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(redisData))
		fmt.Println("data came from redis")
		return
	}

	response, err := getContentCollection().Find(context.Background(), bson.M{})

	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "response failed")
	}

	var movies []primitive.M
	for response.Next(context.Background()) {
		var movie bson.M
		if err := response.Decode(&movie); err != nil {
			log.Fatal(err)
		}
		movies = append(movies, movie)
	}
	jsonData, err := json.Marshal(movies)
	if err == nil {
		config.RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute)
	}
	fmt.Println("data came from mongo")
	helper.SendSuccess(w, http.StatusCreated, "Successful response", movies)
}

func GetContentByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	contentId := params["id"]

	id, err := primitive.ObjectIDFromHex(contentId)
	if err != nil {
		helper.SendError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var content model.Content
	err = getContentCollection().FindOne(context.Background(), bson.M{"_id": id}).Decode(&content)
	if err != nil {
		helper.SendError(w, http.StatusNotFound, "Content not found")
		return
	}
	//Update the redis trending count
	config.RedisClient.ZIncrBy(context.Background(), "trending", 1, contentId)

	helper.SendSuccess(w, http.StatusOK, "Content fetched successfully", content)
}

func UpdateContent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	contentId := params["id"]

	id, err := primitive.ObjectIDFromHex(contentId)
	if err != nil {
		helper.SendError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// decode request body into input struct
	var input model.CreateContentInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helper.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"title":           input.Title,
		"description":     input.Description,
		"type":            input.ContentType,
		"genre":           input.Genre,
		"cast":            input.Cast,
		"director":        input.Director,
		"thumbnail_url":   input.ThumbnailUrl,
		"video_url":       input.VideoUrl,
		"maturity_rating": input.MaturityRating,
		"is_featured":     input.IsFeatured,
		"updated_at":      time.Now(),
	}}

	result, err := getContentCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Failed to update content")
		return
	}

	if result.ModifiedCount == 0 {
		helper.SendError(w, http.StatusNotFound, "Content not found")
		return
	}

	helper.SendSuccess(w, http.StatusOK, "Content updated successfully", nil)
}

func DeleteContent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	contentId := params["id"]

	id, err := primitive.ObjectIDFromHex(contentId)
	if err != nil {
		helper.SendError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	result, err := getContentCollection().DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Failed to delete content")
		return
	}

	if result.DeletedCount == 0 {
		helper.SendError(w, http.StatusNotFound, "Content not found")
		return
	}

	helper.SendSuccess(w, http.StatusAccepted, "Content deleted successfully", nil)
}

//Read CRUD properly pin to pin

func GetTrendingContent(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Step 1 — read top 10 from Redis sorted set
	// ZRevRange returns IDs sorted by score highest first
	// 0 = start, 9 = end (top 10)
	result, err := config.RedisClient.ZRevRange(ctx, "trending", 0, 9).Result()
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Failed to fetch trending")
		return
	}

	// Step 2 — fetch content details from MongoDB for each ID
	var contents []model.Content

	for _, res := range result {
		id, err := primitive.ObjectIDFromHex(res)
		if err != nil {
			continue // skip invalid IDs
		}

		var content model.Content
		err = getContentCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&content)
		if err != nil {
			continue // skip if content not found
		}

		contents = append(contents, content)
	}

	helper.SendSuccess(w, http.StatusOK, "Trending content", contents)
}

func SearchContent(w http.ResponseWriter, r *http.Request) {
	// Step 1 — get search query from URL params
	text := r.URL.Query().Get("q")

	// Step 2 — validate
	if text == "" {
		helper.SendError(w, http.StatusBadRequest, "Please provide search text")
		return
	}

	// Step 3 — build text search filter
	filter := bson.M{
		"$text": bson.M{
			"$search": text, // ← lowercase $search
		},
	}

	// Step 4 — query MongoDB
	cursor, err := getContentCollection().Find(context.Background(), filter)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Search failed")
		return
	}
	defer cursor.Close(context.Background())

	// Step 5 — decode cursor into slice
	var contents []model.Content
	for cursor.Next(context.Background()) {
		var content model.Content
		if err := cursor.Decode(&content); err != nil {
			helper.SendError(w, http.StatusInternalServerError, "Error decoding results")
			return
		}
		contents = append(contents, content)
	}

	// Step 6 — return results
	helper.SendSuccess(w, http.StatusOK, "Search results", contents)
}
