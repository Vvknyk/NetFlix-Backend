package main

import (
	"fmt"
	"log"
	"net/http"

	routes "Netflix/routes"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Hello from mongo")
	r := mux.NewRouter()
	routes.RegisterRoutes(r)
	log.Fatal(http.ListenAndServe(":8080", r)) // ✅ colon before port
}
