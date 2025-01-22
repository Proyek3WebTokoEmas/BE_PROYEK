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

        if tokenString == "" {
            http.Error(w, "Authorization header is required", http.StatusUnauthorized)
            return
        }

        if strings.HasPrefix(tokenString, "Bearer ") {
            tokenString = strings.TrimPrefix(tokenString, "Bearer ")
        } else {
            http.Error(w, "Invalid token format", http.StatusUnauthorized)
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(config.JwtSecret), nil
        })

        if err != nil {
            if strings.Contains(err.Error(), "token is expired") {
                http.Error(w, "Token expired", http.StatusUnauthorized)
            } else {
                log.Printf("Error parsing token: %v", err)
                http.Error(w, "Invalid token", http.StatusUnauthorized)
            }
            return
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			exp, ok := claims["exp"].(float64)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
		
			// Log waktu kadaluarsa dan waktu sekarang
			log.Printf("Token expires at: %v, Current time: %v", time.Unix(int64(exp), 0), time.Now())
		
			// Periksa jika token telah kedaluwarsa
			if time.Now().Unix() > int64(exp) {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
		
			// Simpan klaim di context
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
