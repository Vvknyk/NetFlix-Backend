package controllers

import (
	"Netflix/config"
	helper "Netflix/helpers"
	"fmt"
	"log"
	"strings"

	"Netflix/model"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/qdrant/go-client/qdrant"
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

	if input.Title == "" || input.ContentType == "" {
		helper.SendError(w, http.StatusBadRequest, "Title and type are required")
		return
	}

	content := model.Content{
		ID:             primitive.NewObjectID(),
		Title:          input.Title,
		Description:    input.Description,
		ContentType:    input.ContentType,
		ReleaseYear:    input.ReleaseYear,
		MaturityRating: input.MaturityRating,
		Languages:      input.Languages,
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

	// Step 1 — insert into MongoDB first
	_, err = getContentCollection().InsertOne(context.Background(), content)
	if err != nil {
		fmt.Println("error is", err)
		helper.SendError(w, http.StatusInternalServerError, "Error creating content")
		return
	}

	// Step 2 — generate embedding text
	embeddingText := input.Title + " " +
		input.Description + " " +
		strings.Join(input.Genre, " ") + " " +
		strings.Join(input.Cast, " ") + " " +
		input.ContentType

	// Step 3 — generate vector via Ollama
	vector, err := config.GenerateEmbedding(embeddingText)
	if err != nil {
		fmt.Println("Vector generation failed:", err)
	} else {
		// Step 4 — convert float64 to float32 because float64 used more memory
		float32Vector := make([]float32, len(vector))
		for i, v := range vector {
			float32Vector[i] = float32(v)
		}

		// Step 5 — upsert to Qdrant
		_, err = config.QdrantClient.Upsert(context.Background(), &qdrant.UpsertPoints{
			CollectionName: "content_vectors",
			Points: []*qdrant.PointStruct{
				{
					Id:      qdrant.NewIDNum(uint64(time.Now().UnixNano())),
					Vectors: qdrant.NewVectors(float32Vector...),
					Payload: qdrant.NewValueMap(map[string]any{
						"content_id": content.ID.Hex(),
						"title":      content.Title,
						"type":       content.ContentType,
					}),
				},
			},
		})
		if err != nil {
			fmt.Println("Qdrant upsert failed:", err)
		} else {
			fmt.Println("Vector stored in Qdrant successfully")
		}
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
	text := r.URL.Query().Get("q")
	if text == "" {
		helper.SendError(w, http.StatusBadRequest, "Please provide search text")
		return
	}

	// Step 1 — generate vector from query
	vector, err := config.GenerateEmbedding(text)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Vector generation failed")
		return
	}

	// Step 2 — convert float64 to float32
	float32Vector := make([]float32, len(vector))
	for i, v := range vector {
		float32Vector[i] = float32(v)
	}

	// Step 3 — search Qdrant
	limit := uint64(10)
	results, err := config.QdrantClient.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "content_vectors",
		Query:          qdrant.NewQuery(float32Vector...),
		Limit:          &limit,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Search failed")
		return
	}

	if len(results) == 0 {
		helper.SendSuccess(w, http.StatusOK, "No results found", nil)
		return
	}

	// Step 4 — extract content IDs from payload
	var contentIDs []primitive.ObjectID
	for _, point := range results {
		contentIDStr := point.Payload["content_id"].GetStringValue()
		contentID, err := primitive.ObjectIDFromHex(contentIDStr)
		if err != nil {
			continue
		}
		contentIDs = append(contentIDs, contentID)
	}

	// Step 5 — fetch all content from MongoDB in ONE query using $in
	cursor, err := getContentCollection().Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": contentIDs}},
	)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Failed to fetch content")
		return
	}
	defer cursor.Close(context.Background())

	// Step 6 — decode results
	var contents []model.Content
	for cursor.Next(context.Background()) {
		var content model.Content
		if err := cursor.Decode(&content); err != nil {
			continue
		}
		contents = append(contents, content)
	}

	helper.SendSuccess(w, http.StatusOK, "Search results", contents)
}
