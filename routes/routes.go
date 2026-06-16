package routes

import (
	"Netflix/controllers"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	// auth routes
	r.HandleFunc("/api/auth/register", controllers.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", controllers.Login).Methods("POST")
	r.HandleFunc("/api/auth/logout", controllers.Logout).Methods("POST")
	r.HandleFunc("/api/auth/refresh-token", controllers.RefreshToken).Methods("POST")

	r.HandleFunc("/api/movies", controllers.GetAllMovies).Methods("GET")
	r.HandleFunc("/api/movie", controllers.CreateMovie).Methods("POST")
	r.HandleFunc("/api/movie/{id}", controllers.MarkAsWatched).Methods("PUT")
	r.HandleFunc("/api/movie/{id}", controllers.DeleteMovie).Methods("DELETE")
}
