package main

import (
	"log"
	"net/http"
	"proyek3/config"
	"proyek3/database"
	"proyek3/routes"

	"github.com/rs/cors"
)

func init() {
	// Inisialisasi database saat server dimulai
	database.InitDB()
}

// Handler function yang digunakan oleh Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Inisialisasi konfigurasi dan database (hanya dilakukan sekali)
	config.InitConfig()
	database.InitDB()

	// Inisialisasi router dengan middleware CORS
	router := routes.InitRoutes()
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1:5501", "https://proyek3webtokoemas.github.io"}, // Disesuaikan
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(router)

	// Lanjutkan permintaan ke router
	corsHandler.ServeHTTP(w, r)
}

func main() {
	database.InitDB()

	http.HandleFunc("/", Handler)
	log.Println("Starting server on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

