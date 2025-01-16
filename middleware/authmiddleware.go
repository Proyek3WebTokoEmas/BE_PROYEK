package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"proyek3/config"
	"github.com/dgrijalva/jwt-go"
)

// Definisikan tipe kunci khusus untuk context
type contextKey string

// Definisikan konstanta untuk kunci context
const claimsKey contextKey = "claims"

// AuthMiddleware adalah middleware untuk memeriksa token JWT
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		// Periksa apakah token ada
		if tokenString == "" {
			http.Error(w, "Token is required", http.StatusUnauthorized)
			return
		}

		// Hapus prefix "Bearer " jika ada
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		} else {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// Parsing token untuk memverifikasi
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Memastikan metode signing adalah HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Mengembalikan secret key untuk verifikasi
			return []byte(config.JwtSecret), nil
		})

		// Menangani error parsing token
		if err != nil {
			log.Printf("Error parsing token: %v", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Verifikasi klaim jika token valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Memeriksa waktu kadaluarsa token
			expTime := claims["exp"].(float64)
			if time.Now().Unix() > int64(expTime) {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}

			// Set klaim dalam context menggunakan kunci khusus
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		}
	})
}



func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Membolehkan akses dari semua origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Menangani pre-flight request (OPTIONS)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
