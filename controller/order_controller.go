package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"log"
	"net/http"
	"proyek3/config"
	"proyek3/database"
	"proyek3/model"

	"github.com/dgrijalva/jwt-go"
)

// HandleGetOrders mengembalikan semua data pesanan dari database
func HandleGetOrders(w http.ResponseWriter, r *http.Request) {
	// Ambil data orders dari database
	rows, err := database.DB.Query(`SELECT jenis_perhiasan, jenis_emas, berat_emas, campuran_tambahan, persentase_emas, total_harga FROM custom_orders`)
	if err != nil {
		http.Error(w, "Error fetching orders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []model.OrderRequest
	for rows.Next() {
		var OrderRequest model.OrderRequest
		if err := rows.Scan(&OrderRequest.JenisPerhiasan, &OrderRequest.JenisEmas, &OrderRequest.BeratEmas, &OrderRequest.CampuranTambahan, &OrderRequest.PersentaseEmas, &OrderRequest.TotalHarga); err != nil {
			http.Error(w, "Error scanning order data", http.StatusInternalServerError)
			return
		}
		orders = append(orders, OrderRequest)
	}

	// Kirimkan respons sebagai JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

var jwtKey = []byte("ClWIuIWqLaIhl1GRZi-bYjsM4niJxUeQe4Ot4nObVHY") // Ganti dengan secret key Anda
// HandleGetOrderById mengembalikan detail pesanan berdasarkan ID
func HandleGetOrdersByUserId(w http.ResponseWriter, r *http.Request) {
	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"message": "Authorization header missing"}`, http.StatusUnauthorized)
		return
	}

	// Pisahkan "Bearer" dan tokennya
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, `{"message": "Invalid Authorization format"}`, http.StatusUnauthorized)
		return
	}

	// Parsing token
	claims := &config.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JwtSecret), nil
	})

	// Periksa jika token valid
	if err != nil || !token.Valid {
		http.Error(w, `{"message": "Invalid token"}`, http.StatusUnauthorized)
		return
	}

	// Ambil user_id dari klaim token
	userID := claims.UserID
	fmt.Println("User ID:", userID)

	// Query untuk mengambil data pesanan berdasarkan user_id
	rows, err := database.DB.Query(`
        SELECT id, jenis_perhiasan, jenis_emas, berat_emas, campuran_tambahan, persentase_emas, total_harga
        FROM custom_orders
        WHERE user_id = $1
    `, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		fmt.Println("Database query error:", err)
		return
	}
	defer rows.Close()

	// Parsing hasil query ke dalam slice
	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.ID,
			&order.JenisPerhiasan,
			&order.JenisEmas,
			&order.BeratEmas,
			&order.CampuranTambahan,
			&order.PersentaseEmas,
			&order.TotalHarga,
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error reading data"})
			fmt.Println("Error scanning row:", err)
			return
		}
		orders = append(orders, order)
	}

	// Kirim respons JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// HandleAddOrder meng-handle POST request untuk menambahkan order ke database
func HandleAddOrder(w http.ResponseWriter, r *http.Request) {
	// Logging untuk debugging
	log.Println("Handling Add Order Request")

	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"message": "Authorization header missing"}`, http.StatusUnauthorized)
		return
	}

	// Pisahkan "Bearer" dan tokennya
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, `{"message": "Invalid Authorization format"}`, http.StatusUnauthorized)
		return
	}

	// Parsing token
	claims := &config.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JwtSecret), nil
	})

	// Periksa jika token valid
	if err != nil || !token.Valid {
		http.Error(w, `{"message": "Invalid token"}`, http.StatusUnauthorized)
		return
	}

	// Ambil user_id dari klaim token
	userID := claims.UserID
	log.Printf("User ID from token: %d", userID)

	// Parse data dari body request
	var order model.OrderRequest
	body, err := io.ReadAll(r.Body) // Membaca isi body request
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		log.Printf("Error reading body: %v", err)
		return
	}

	log.Printf("Request Body: %s", string(body)) // Debug isi body

	err = json.Unmarshal(body, &order) // Decode JSON ke struct
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		log.Printf("Error decoding JSON: %v", err)
		return
	}

	// Validasi data yang diperlukan
	if order.JenisPerhiasan == "" || order.JenisEmas == "" || order.BeratEmas == 0 || order.PersentaseEmas == 0 || order.TotalHarga == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		log.Println("Validation failed: Missing required fields")
		return
	}

	// Logging data yang diterima
	log.Printf("Received Order Data: %+v", order)

	// Query untuk memasukkan data ke database
	query := `INSERT INTO custom_orders (user_id, jenis_perhiasan, jenis_emas, berat_emas, campuran_tambahan, persentase_emas, total_harga) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var id int
	err = database.DB.QueryRow(query, userID, order.JenisPerhiasan, order.JenisEmas, order.BeratEmas, order.CampuranTambahan, order.PersentaseEmas, order.TotalHarga).Scan(&id)
	if err != nil {
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		log.Printf("Error inserting data into database: %v", err)
		return
	}

	// Response sukses
	response := map[string]interface{}{
		"message": "Emas berhasil ditambahkan",
		"id":      id,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
