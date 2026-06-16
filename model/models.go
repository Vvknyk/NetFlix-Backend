package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NetFlix struct {
	Id      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Movie   string             `json:"movie,omitempty" bson:"movie,omitempty"`     // ✅ fixed typo + closed backtick
	Watched bool               `json:"watched,omitempty" bson:"watched,omitempty"` // ✅ closed backtick
}

type User struct {
	ID                 primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name               string             `json:"name" bson:"name"`
	Email              string             `json:"email" bson:"email"`
	Password           string             `json:"password" bson:"password"`
	Role               string             `json:"role" bson:"role"`
	SubscriptionPlan   string             `json:"subscription_plan" bson:"subscription_plan"`
	SubscriptionExpiry time.Time          `json:"subscription_expiry" bson:"subscription_expiry"`
	RefreshToken       string             `json:"refresh_token" bson:"refresh_token"`
	IsActive           bool               `json:"is_active" bson:"is_active"`
	CreatedAt          time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" bson:"updated_at"`
}
