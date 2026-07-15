package routes

import (
	"Netflix/controllers"
	"Netflix/middleware"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	// auth routes
	r.HandleFunc("/api/auth/register", controllers.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", controllers.Login).Methods("POST")

	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/api/auth/logout", controllers.Logout).Methods("POST")
	protected.HandleFunc("/api/auth/refresh-token", controllers.RefreshToken).Methods("POST")
	r.HandleFunc("/content/search", controllers.SearchContent).Methods("GET")
	admin := r.PathPrefix("/api/admin").Subrouter()
	admin.Use(middleware.AuthMiddleware)
	admin.Use(middleware.AdminCheck)

	admin.HandleFunc("/createContent", controllers.CreateContent).Methods("POST")
	r.HandleFunc("/GetAllContent", controllers.GetAllContent).Methods("GET")
	admin.HandleFunc("/GetContentByID/{id}", controllers.GetContentByID).Methods("GET")
	admin.HandleFunc("/UpdateContent/{id}", controllers.GetAllContent).Methods("GET")
	admin.HandleFunc("/DeleteContent/{id}", controllers.GetAllContent).Methods("GET")

}
