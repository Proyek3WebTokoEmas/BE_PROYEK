package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"proyek3/config"
	"proyek3/database"
	"proyek3/model"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/crypto/bcrypt"
)

// Struktur untuk menyimpan email dan password yang diterima dalam request
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Claims untuk JWT
type Claims struct {
	UserID int    `json:"user_id"`
	Email string `json:"email"`
	jwt.StandardClaims
}

func CreateToken(user model.User) (string, error) {
    // Tambahkan klaim id, email, dan role
    claims := jwt.MapClaims{
        "id":      user.ID,
        "email":   user.Email,
        "role":    user.Role,  // Menambahkan role ke klaim
        "exp":     time.Now().Add(72 * time.Hour).Unix(),
    }

    // Membuat token dengan klaim
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Tanda tangani token dengan secret key
    tokenString, err := token.SignedString([]byte(config.JwtSecret))
    if err != nil {
        return "", fmt.Errorf("error signing token: %w", err)
    }

    return tokenString, nil
}

// Fungsi untuk mengirim email konfirmasi menggunakan SendGrid
func sendVerificationEmail(toEmail, verificationToken string) error {
	apiKey := config.SendGridAPIKey
	verificationLink := fmt.Sprintf("https://be-3.vercel.app/verify-email?token=%s", verificationToken) // Cukup kirimkan token

	from := mail.NewEmail("Your App", "fathir080604@gmail.com")
	to := mail.NewEmail("User", toEmail)
	subject := "Konfirmasi Registrasi"

	htmlContent := `
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; background-color: #f4f4f4; color: #333; margin: 0; padding: 0; }
				.container { max-width: 600px; margin: 30px auto; padding: 20px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1); }
				.header { text-align: center; padding-bottom: 20px; }
				.header h2 { color: #007BFF; }
				.content p { font-size: 16px; }
				.button { display: inline-block; padding: 10px 20px; background-color: #28a745; color: #fff; border-radius: 5px; text-decoration: none; font-size: 16px; }
				.footer { margin-top: 20px; font-size: 14px; color: #888; text-align: center; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h2>Terima kasih telah mendaftar !</h2>
				</div>
				<div class="content">
					<p>Hi,</p>
					<p>Terima kasih telah mendaftar ! Untuk melanjutkan, silakan klik tombol di bawah ini untuk memverifikasi akun Anda:</p>
					<p><a href="` + verificationLink + `" class="button">Verifikasi Akun</a></p>
					<p>Jika Anda tidak mendaftar di website kami, abaikan email ini.</p>
				</div>
				<div class="footer">
					<p>&copy; 2024 OurApp. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>
	`

	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
	client := sendgrid.NewSendClient(apiKey)
	_, err := client.Send(message)
	return err
}

// Fungsi untuk memverifikasi email dengan token
func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	// Ambil token dari URL
	tokenURL := r.URL.Query().Get("token")

	// Tambahkan log di sini
	log.Println("Token received:", tokenURL)

	// Lakukan parsing token dan validasi
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenURL, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JwtSecret), nil
	})
	if err != nil {
		log.Printf("Error parsing token: %v\n", err)
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	// Pastikan token valid
	if !tkn.Valid {
		log.Println("Token is not valid")
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Log email yang didapat dari token
	log.Println("Parsed email from token:", claims.Email)

	// Proses verifikasi di database
	_, err = database.DB.Exec(`UPDATE "user" SET verified=true WHERE email=$1`, claims.Email)
	if err != nil {
		log.Printf("Database error: %v\n", err)
		http.Error(w, "Failed to verify email", http.StatusInternalServerError)
		return
	}

	// Log sukses verifikasi
	log.Println("Email verified successfully for:", claims.Email)

	w.Write([]byte("Email verified successfully"))
}

// Fungsi helper untuk mengembalikan nilai default jika kosong
func defaultIfEmpty(value, defaultValue string) string {
    if value == "" {
        return defaultValue
    }
    return value
}


// Fungsi untuk register pengguna baru
func Register(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}


	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	role := "" // Misalnya nilai role yang kosong
	role = defaultIfEmpty(role, "user")

	log.Println("Role:", role)

	// Insert user into the database and retrieve the ID
	var userID int
	err = database.DB.QueryRow(`
		INSERT INTO "user" (name, email, password, verified, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, user.Name, user.Email, string(hashedPassword), false, "user").Scan(&userID)

	if err != nil {
		http.Error(w, "Error saving user", http.StatusInternalServerError)
	
		log.Printf("Error saving user: %v\n", err)
		return
	}

	// Create token
	tokenString, err := CreateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Send verification email
	err = sendVerificationEmail(user.Email, tokenString)
	if err != nil {
		http.Error(w, "Error sending verification email", http.StatusInternalServerError)
		return
	}

	// Return success response with user ID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registration successful, verify your email",
		"user_id": userID,
	})
}

// Fungsi untuk login pengguna
func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, `{"message": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Ambil data user berdasarkan email
	var user model.User
	err = database.DB.QueryRow(`
    SELECT id, email, password, verified, name, role 
    FROM "user" 
    WHERE email = $1`, creds.Email).Scan(&user.ID, &user.Email, &user.Password, &user.Verified, &user.Name, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"message": "User not found"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Verifikasi jika user belum terverifikasi
	if !user.Verified {
		http.Error(w, `{"message": "Email not verified"}`, http.StatusUnauthorized)
		return
	}

	// Verifikasi password yang di-hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		http.Error(w, `{"message": "Invalid password"}`, http.StatusUnauthorized)
		return
	}

	// Generate JWT Token
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		UserID: user.ID,
		Email: creds.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JwtSecret))
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Kirim token dan nama ke client
	response := map[string]interface{}{
		"message": "Login successful",
		"token":   tokenString, // Pastikan token dikirim
		"name":    user.Name,
		"role":    user.Role,
	}
	json.NewEncoder(w).Encode(response)
}

// Fungsi untuk memverifikasi token JWT
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ambil token dari header Authorization
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// Verifikasi token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNoLocation
			}
			return []byte(config.JwtSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Token valid, lanjutkan ke handler berikutnya
		next.ServeHTTP(w, r)
	})
}

func verifyToken(tokenString string) (int, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return []byte(config.JwtSecret), nil
    })

    if err != nil || !token.Valid {
        return 0, fmt.Errorf("invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || claims["id"] == nil {
        return 0, fmt.Errorf("no id in token")
    }

    userID, ok := claims["id"].(float64)
    if !ok {
        return 0, fmt.Errorf("invalid id in token")
    }

    return int(userID), nil
}