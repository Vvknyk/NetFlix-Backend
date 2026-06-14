package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type NetFlix struct {
	Id      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Movie   string             `json:"movie,omitempty" bson:"movie,omitempty"`     // ✅ fixed typo + closed backtick
	Watched bool               `json:"watched,omitempty" bson:"watched,omitempty"` // ✅ closed backtick
}
