package routes

import (
	"proyek3/controller"
	"proyek3/middleware"

	"github.com/gorilla/mux"
)

// InitRoutes menginisialisasi semua route
func InitRoutes() *mux.Router {
	router := mux.NewRouter()

	// Middleware CORS untuk semua route
	router.Use(middleware.CorsMiddleware)

	// Route umum tanpa autentikasi
	router.HandleFunc("/api/register", controller.Register).Methods("POST")
	router.HandleFunc("/api/login", controller.Login).Methods("POST")
	router.HandleFunc("/verify-email", controller.VerifyEmail).Methods("GET")

	// Subrouter untuk endpoint yang memerlukan autentikasi JWT
	protected := router.PathPrefix("/api/protected").Subrouter()
	protected.Use(middleware.AuthMiddleware) // Middleware autentikasi JWT

	// Endpoint dengan autentikasi
	protected.HandleFunc("/tambah-emas", controller.TambahEmas).Methods("POST")
	protected.HandleFunc("/emas", controller.GetAllEmas).Methods("GET")
	protected.HandleFunc("/emas/update", controller.UpdateEmas).Methods("PUT")
	protected.HandleFunc("/payment", controller.CreatePayment).Methods("POST")

	// Endpoint untuk pesanan, termasuk menambah dan mendapatkan pesanan
	protected.HandleFunc("/orders", controller.HandleAddOrder).Methods("POST")         // Menambahkan order
	protected.HandleFunc("/orders", controller.HandleGetOrders).Methods("GET")         // Mendapatkan semua pesanan
	protected.HandleFunc("/orders/{id}", controller.HandleGetOrdersByUserId).Methods("GET") // Mendapatkan pesanan berdasarkan ID

	return router
}
