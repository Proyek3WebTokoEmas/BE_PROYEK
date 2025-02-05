package config

import (
	"log"
	"os"

	"github.com/dgrijalva/jwt-go"
	// "github.com/joho/godotenv"
)

var JwtSecret string
var SendGridAPIKey string

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

func InitConfig() {
	// Memuat file
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	// Ambil nilai variabel environment JWT_SECRET dari file .env
	JwtSecret = os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in the .env file")
	}
	log.Println("JWT_SECRET loaded successfully")

	// Ambil nilai API Key untuk SendGrid
	SendGridAPIKey = os.Getenv("SENDGRID_API_KEY")
	if SendGridAPIKey == "" {
		log.Fatal("SENDGRID_API_KEY is not set in the .env file")
	}
	log.Println("SENDGRID_API_KEY loaded successfully")
}
