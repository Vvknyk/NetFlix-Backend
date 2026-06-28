package middleware

import (
	"Netflix/config"
	helper "Netflix/helpers"
	"context"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Step 1 — get token ✅ you did this, fix spelling
		tokenAuth := r.Header.Get("Authorization")
		if tokenAuth == "" { // fix: == not =
			helper.SendError(w, http.StatusUnauthorized, "No token provided")
			return // ← add this, you were missing it
		}

		// Step 2 — strip "Bearer " prefix  ← you missed this
		tokenStr := strings.TrimPrefix(tokenAuth, "Bearer ")

		// Step 3 — check Redis blacklist  ← you missed this
		ctx := context.Background()
		val, err := config.RedisClient.Get(ctx, "blacklist:"+tokenStr).Result()
		if err == nil && val == "true" {
			helper.SendError(w, http.StatusUnauthorized, "Token is logged out")
			return
		}

		// Step 4 — validate JWT ✅ you did this, just fix return
		claims, err := helper.ValidateAccessToken(tokenStr)
		if err != nil {
			helper.SendError(w, http.StatusUnauthorized, "Token not valid")
			return // ← add this
		}

		// Step 5 — attach claims to context  ← you missed this
		ctx = context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "role", claims.Role)
		r = r.WithContext(ctx)

		// Step 6 — call next handler  ← you missed this
		next.ServeHTTP(w, r)
	})
}
