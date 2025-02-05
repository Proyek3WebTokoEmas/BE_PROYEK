package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"proyek3/database"
	"proyek3/model"

	"github.com/gorilla/mux"
)

// CreateUser - Menambahkan user baru
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser model.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	newUser.Role = "user"

	query := `INSERT INTO public."user" (name, email, password, verified, role) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = database.DB.QueryRow(query, newUser.Name, newUser.Email, newUser.Password, newUser.Verified, newUser.Role).Scan(&newUser.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

// GetUsers - Mengambil daftar user
func GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`SELECT id, name, email, password, verified FROM public."user"`) // Tambahkan schema public
	if err != nil {
		log.Println("Error fetching users:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Verified); err != nil {
			log.Println("Error scanning user:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over users:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetUserByID - Mengambil user berdasarkan ID
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// Convert id ke integer
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid ID:", id)
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var user model.User
	err = database.DB.QueryRow(`SELECT id, name, email, password, verified FROM public."user" WHERE id = $1`, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Verified,
	)

	if err != nil {
		log.Println("Error fetching user with ID:", userID, "| Error:", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateUser - Memperbarui data user
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// Konversi id ke integer agar aman digunakan dalam query
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	var updatedUser model.User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Menyusun query UPDATE secara dinamis
	var queryParts []string
	var args []interface{}

	if updatedUser.Name != "" {
		queryParts = append(queryParts, "name = $"+strconv.Itoa(len(args)+1))
		args = append(args, updatedUser.Name)
	}
	if updatedUser.Email != "" {
		queryParts = append(queryParts, "email = $"+strconv.Itoa(len(args)+1))
		args = append(args, updatedUser.Email)
	}
	if updatedUser.Password != "" {
		queryParts = append(queryParts, "password = $"+strconv.Itoa(len(args)+1))
		args = append(args, updatedUser.Password)
	}
	if updatedUser.Verified {
		queryParts = append(queryParts, "verified = $"+strconv.Itoa(len(args)+1))
		args = append(args, updatedUser.Verified)
	}

	// Jika tidak ada field yang diupdate, hentikan proses
	if len(queryParts) == 0 {
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	// Menyusun query akhir
	query := "UPDATE public.\"user\" SET " + strings.Join(queryParts, ", ") + " WHERE id = $" + strconv.Itoa(len(args)+1)
	args = append(args, userID)

	// Eksekusi query
	_, err = database.DB.Exec(query, args...)
	if err != nil {
		http.Error(w, "Error updating user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Response sukses
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User updated successfully"))
}


// DeleteUser - Menghapus user berdasarkan ID
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// Konversi ID ke integer agar lebih aman
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM public."user" WHERE id = $1`
	_, err = database.DB.Exec(query, userID)
	if err != nil {
		http.Error(w, "Error deleting user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted successfully"))
}
