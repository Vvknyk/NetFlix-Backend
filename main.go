package main

import (
	"fmt"
	"log"
	"net/http"

	"Netflix/config"
	"Netflix/routes"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// load .env first, before anything else
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found")
	}

	// connect to databases
	config.ConnectDB()
	config.ConnectRedis()
	config.CreateTextIndex()

	// setup router
	r := mux.NewRouter()
	routes.RegisterRoutes(r)

	fmt.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
