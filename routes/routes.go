package routes

import (
	"Netflix/controllers"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/movies", controllers.GetAllMovies).Methods("GET")
	r.HandleFunc("/api/movie", controllers.CreateMovie).Methods("POST")
	r.HandleFunc("/api/movie/{id}", controllers.MarkAsWatched).Methods("PUT")
	r.HandleFunc("/api/movie/{id}", controllers.DeleteMovie).Methods("DELETE")
}
