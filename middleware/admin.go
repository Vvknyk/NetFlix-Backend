package middleware

import (
	helper "Netflix/helpers"
	"net/http"
)

func AdminCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		role := r.Context().Value("role").(string)

		if role != "admin" {
			helper.SendError(w, http.StatusForbidden, "you do noy have access to this page")
			return
		}

		next.ServeHTTP(w, r)

	})

}
