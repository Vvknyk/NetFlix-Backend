package controllers

import (
	"Netflix/config"
	helper "Netflix/helpers"
	"Netflix/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// ─── Register ─────────────────────────────────────────────────────────────────

func Register(w http.ResponseWriter, r *http.Request) {
	var user model.User

	// decode request body
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		helper.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// validate input
	if user.Name == "" || user.Email == "" || user.Password == "" {
		helper.SendError(w, http.StatusBadRequest, "Name, email and password are required")
		return
	}

	col := config.GetCollection("users")
	ctx := context.Background()

	// check if email already exists
	var existingUser model.User
	err = col.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		helper.SendError(w, http.StatusConflict, "Email already registered")
		return
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Error processing password")
		return
	}

	// set default values
	user.ID = primitive.NewObjectID()
	user.Password = string(hashedPassword)
	user.Role = "user"
	user.SubscriptionPlan = "basic"
	user.IsActive = true
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// insert into mongodb
	_, err = col.InsertOne(ctx, user)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	// never return password
	user.Password = ""

	helper.SendSuccess(w, http.StatusCreated, "User registered successfully", user)
}

// ─── Login ────────────────────────────────────────────────────────────────────

func Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helper.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// validate input
	if input.Email == "" || input.Password == "" {
		helper.SendError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	col := config.GetCollection("users")
	ctx := context.Background()

	// find user by email
	var user model.User
	err = col.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		helper.SendError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// check if user is active
	if !user.IsActive {
		helper.SendError(w, http.StatusForbidden, "Account is disabled")
		return
	}

	// compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	fmt.Println("pww", bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)))
	if err != nil {
		helper.SendError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// generate tokens
	accessToken, err := helper.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	refreshToken, err := helper.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Error generating refresh token")
		return
	}

	// save refresh token in mongodb
	_, err = col.UpdateOne(ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"refresh_token": refreshToken,
			"updated_at":    time.Now(),
		}},
	)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Error saving session")
		return
	}

	// send response
	helper.SendSuccess(w, http.StatusOK, "Login successful", map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]interface{}{
			"id":    user.ID.Hex(),
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
			"plan":  user.SubscriptionPlan,
		},
	})
}

// ─── Logout ───────────────────────────────────────────────────────────────────

func Logout(w http.ResponseWriter, r *http.Request) {
	// get token from header
	tokenStr := r.Header.Get("Authorization")
	if tokenStr == "" {
		helper.SendError(w, http.StatusUnauthorized, "No token provided")
		return
	}

	// validate token to get expiry
	claims, err := helper.ValidateAccessToken(tokenStr)
	if err != nil {
		helper.SendError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// add token to redis blacklist with TTL = remaining expiry time
	ctx := context.Background()
	expiry := time.Until(claims.ExpiresAt.Time)
	config.RedisClient.Set(ctx, "blacklist:"+tokenStr, "true", expiry)

	// clear refresh token in mongodb
	col := config.GetCollection("users")
	col.UpdateOne(ctx,
		bson.M{"_id": claims.UserID},
		bson.M{"$set": bson.M{
			"refresh_token": "",
			"updated_at":    time.Now(),
		}},
	)

	helper.SendSuccess(w, http.StatusOK, "Logged out successfully", nil)
}

// ─── Refresh Token ────────────────────────────────────────────────────────────

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil || input.RefreshToken == "" {
		helper.SendError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// validate refresh token
	claims, err := helper.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		helper.SendError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// find user and verify refresh token matches
	col := config.GetCollection("users")
	ctx := context.Background()

	userID, _ := primitive.ObjectIDFromHex(claims.UserID)
	var user model.User
	err = col.FindOne(ctx, bson.M{
		"_id":           userID,
		"refresh_token": input.RefreshToken,
	}).Decode(&user)
	if err != nil {
		helper.SendError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// generate new access token
	accessToken, err := helper.GenerateAccessToken(user.ID.Hex(), user.Email, user.Role)
	if err != nil {
		helper.SendError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	helper.SendSuccess(w, http.StatusOK, "Token refreshed", map[string]string{
		"access_token": accessToken,
	})
}
