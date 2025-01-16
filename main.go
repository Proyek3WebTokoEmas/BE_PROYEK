package main

import (
	"log"
	"net/http"
	"proyek3/config"
	"proyek3/database"
	"proyek3/routes"

	"github.com/rs/cors"
)

func main() {
	config.InitConfig()
	database.InitDB()

	// Menambahkan CORS untuk mengizinkan permintaan dari http://127.0.0.1:5501
	router := routes.InitRoutes()
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://127.0.0.1:5501"}, // Disesuaikan
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(router)	
	
	log.Println("Server started at :8081")
	log.Fatal(http.ListenAndServe(":8081", corsHandler))
}
