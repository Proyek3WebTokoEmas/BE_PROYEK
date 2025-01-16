package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"proyek3/database"
	"proyek3/model"
)

// Fungsi ini sekarang kompatibel dengan mux
func TambahEmas(w http.ResponseWriter, r *http.Request) {
	// Parse data dari body request
	var emas model.Emas
	err := json.NewDecoder(r.Body).Decode(&emas)
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// Query untuk memasukkan data ke database
	query := `INSERT INTO emas (nama, karatan, berat, harga) VALUES ($1, $2, $3, $4)`
	_, err = database.DB.Exec(query, emas.Nama, emas.Karatan, emas.Berat, emas.Harga)
	if err != nil {
		log.Printf("Error inserting data into database: %v", err)
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		return
	}

	// Response sukses
	response := map[string]string{
		"message": "Emas berhasil ditambahkan",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetAllEmas(w http.ResponseWriter, r *http.Request) {
	var emasList []model.Emas

	// Query untuk mengambil semua data emas
	Query := `SELECT id, nama, karatan, berat, harga FROM emas`
	rows, err := database.DB.Query(Query)
	if err != nil {
		log.Printf("Error reading data from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error reading data"})
		return
	}
	defer rows.Close()

	// Iterasi hasil query
	for rows.Next() {
		var emas model.Emas
		if err := rows.Scan(&emas.ID, &emas.Nama, &emas.Karatan, &emas.Berat, &emas.Harga); err != nil {
			log.Printf("Error scanning data: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error scanning data"})
			return
		}
		emasList = append(emasList, emas)
	}

	// Check for any error during iteration
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error iterating rows"})
		return
	}

	// Mengirim response JSON
	response := map[string]interface{}{
		"status": "success",
		"data":   emasList,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func UpdateEmas(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Invalid query parameter", http.StatusBadRequest)
		return
	}

	// Parse data dari body request
	var emasRequest struct {
		Nama    string  `json:"nama"`
		Karatan int     `json:"karatan"`
		Berat   float64 `json:"berat"`
		Harga   float64 `json:"harga"`
	}
	err := json.NewDecoder(r.Body).Decode(&emasRequest)
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	Query := `UPDATE emas SET nama=$1, karatan=$2, berat=$3, harga=$4 WHERE id=$5`
	_, err = database.DB.Exec(Query, emasRequest.Nama, emasRequest.Karatan, emasRequest.Berat, emasRequest.Harga, id)
	if err != nil {
		log.Printf("Error updating data: %v", err)
		http.Error(w, "Error updating data", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Data emas berhasil diupdate",
		"id":      id,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
