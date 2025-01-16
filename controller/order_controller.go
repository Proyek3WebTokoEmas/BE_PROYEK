package controller

import (
	"encoding/json"
	"io"

	"log"
	"net/http"
	"proyek3/database"
	"proyek3/model"
	"strconv"

	"github.com/gorilla/mux"
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

// HandleGetOrderById mengembalikan detail pesanan berdasarkan ID
func HandleGetOrderById(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari parameter URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Query untuk mengambil data pesanan berdasarkan ID
	var order model.Order
	err = database.DB.QueryRow(`
		SELECT id, jenis_perhiasan, jenis_emas, berat_emas, campuran_tambahan, persentase_emas, total_harga 
		FROM custom_orders WHERE id = $1
	`, id).Scan(
		&order.ID,
		&order.JenisPerhiasan,
		&order.JenisEmas,
		&order.BeratEmas,
		&order.CampuranTambahan,
		&order.PersentaseEmas,
		&order.TotalHarga,
	)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Kirim respons dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// HandleAddOrder meng-handle POST request untuk menambahkan order ke database
func HandleAddOrder(w http.ResponseWriter, r *http.Request) {
	// Logging untuk debugging
	log.Println("Handling Add Order Request")

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
	query := `INSERT INTO custom_orders (jenis_perhiasan, jenis_emas, berat_emas, campuran_tambahan, persentase_emas, total_harga) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	var id int
	err = database.DB.QueryRow(query, order.JenisPerhiasan, order.JenisEmas, order.BeratEmas, order.CampuranTambahan, order.PersentaseEmas, order.TotalHarga).Scan(&id)
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
// 	beratEmasStr := r.FormValue("berat_emas")
// 	persentaseEmasStr := r.FormValue("persentase_emas")
// 	totalHargaStr := r.FormValue("total_harga")

// 	if beratEmasStr == "" || persentaseEmasStr == "" || totalHargaStr == "" {
// 		http.Error(w, "Missing required fields", http.StatusBadRequest)
// 		log.Println("Error: Missing one or more required fields")
// 		return
// 	}

// 	beratEmas, err := strconv.ParseFloat(beratEmasStr, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid berat_emas value", http.StatusBadRequest)
// 		log.Println("Error converting berat_emas:", err)
// 		return
// 	}

// 	persentaseEmas, err := strconv.ParseFloat(persentaseEmasStr, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid persentase_emas value", http.StatusBadRequest)
// 		log.Println("Error converting persentase_emas:", err)
// 		return
// 	}

// 	totalHarga, err := strconv.ParseFloat(totalHargaStr, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid total_harga value", http.StatusBadRequest)
// 		log.Println("Error converting total_harga:", err)
// 		return
// 	}


// 	// Membuat struct OrderRequest
// 	orderRequest := model.OrderRequest{
// 		JenisPerhiasan:   r.FormValue("jenis_perhiasan"),
// 		JenisEmas:        r.FormValue("jenis_emas"),
// 		BeratEmas:        beratEmas,
// 		CampuranTambahan: r.FormValue("campuran_tambahan"),
// 		PersentaseEmas:   persentaseEmas,
// 		TotalHarga:       totalHarga,
// 	}

// 	// Lanjutkan dengan proses lainnya (seperti validasi dan penyimpanan ke database)
// 	log.Printf("Received order: %+v\n", orderRequest)

// 	// Query untuk memasukkan data order ke dalam tabel 'orders'
// 	query := `INSERT INTO custom_orders (jenis_perhiasan, jenis_emas, berat_emas, campuran_tambahan, persentase_emas, total_harga)
//               VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

// 	var id int
// 	err = database.DB.QueryRow(query, orderRequest.JenisPerhiasan, orderRequest.JenisEmas, orderRequest.BeratEmas, orderRequest.CampuranTambahan, orderRequest.PersentaseEmas, orderRequest.TotalHarga).Scan(&id)
// 	if err != nil {
// 		http.Error(w, "Failed to insert data", http.StatusInternalServerError)
// 		log.Println("Error inserting data into DB:", err)
// 		return
// 	}

// 	// Mengirim respons jika berhasil
// 	w.Header().Set("Content-Type", "application/json")
// 	response := map[string]interface{}{"status": "success", "id": id}
// 	json.NewEncoder(w).Encode(response)

