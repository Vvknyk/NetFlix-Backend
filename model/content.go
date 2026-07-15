package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Content struct {
	// ID
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`

	// Basic Info
	Title          string   `json:"title" bson:"title"`
	Description    string   `json:"description" bson:"description"`
	ContentType    string   `json:"type" bson:"type"`
	ReleaseYear    int      `json:"release_year" bson:"release_year"`
	MaturityRating string   `json:"maturity_rating" bson:"maturity_rating"`
	Language       []string `json:"language" bson:"language"`

	// Media
	ThumbnailUrl string `json:"thumbnail_url" bson:"thumbnail_url"`
	VideoUrl     string `json:"video_url" bson:"video_url"`
	TrailerUrl   string `json:"trailer_url" bson:"trailer_url"`

	// Categorization
	Genre    []string `json:"genre" bson:"genre"`
	Cast     []string `json:"cast" bson:"cast"`
	Director string   `json:"director" bson:"director"`

	// Stats
	TotalViews    int     `json:"total_views" bson:"total_views"`
	AverageRating float64 `json:"average_rating" bson:"average_rating"`

	// Series only
	TotalSeasons int `json:"total_seasons,omitempty" bson:"total_seasons,omitempty"`

	// Admin Control
	IsFeatured bool      `json:"is_featured" bson:"is_featured"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

// used when admin creates or updates content
type CreateContentInput struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	ContentType    string   `json:"type"`
	ReleaseYear    int      `json:"release_year"`
	MaturityRating string   `json:"maturity_rating"`
	Languages      []string `json:"languages"`
	ThumbnailUrl   string   `json:"thumbnail_url"`
	VideoUrl       string   `json:"video_url"`
	TrailerUrl     string   `json:"trailer_url"`
	Genre          []string `json:"genre"`
	Cast           []string `json:"cast"`
	Director       string   `json:"director"`
	TotalSeasons   int      `json:"total_seasons"`
	IsFeatured     bool     `json:"is_featured"`
}
