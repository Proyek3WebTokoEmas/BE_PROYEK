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

	router.HandleFunc("/users", controller.GetUsers).Methods("GET")         // GET semua user
	router.HandleFunc("/users/{id}", controller.GetUserByID).Methods("GET") // GET user by ID
	router.HandleFunc("/users", controller.CreateUser).Methods("POST")      // POST user baru
	router.HandleFunc("/users/{id}", controller.UpdateUser).Methods("PUT")  // PUT update user
	router.HandleFunc("/users/{id}", controller.DeleteUser).Methods("DELETE") // DELETE user

	router.HandleFunc("/orders", controller.HandleAddOrder).Methods("POST")         // Menambahkan order
	router.HandleFunc("/orders", controller.HandleGetOrders).Methods("GET")         // Mendapatkan semua pesanan
	router.HandleFunc("/orders/{id}", controller.HandleGetOrdersByUserId).Methods("GET") // Mendapatkan pesanan berdasarkan ID
	
	router.HandleFunc("/payment", controller.CreatePayment).Methods("POST")
	router.HandleFunc("/payment/status", controller.GetAllPayments).Methods("GET") // Mendapatkan status pembayaran
	router.HandleFunc("/webhook/midtrans", controller.WebhookHandler).Methods("POST")

	// Subrouter untuk endpoint yang memerlukan autentikasi JWT
	protected := router.PathPrefix("/api/protected").Subrouter()
	protected.Use(middleware.AuthMiddleware) // Middleware autentikasi JWT

	// Endpoint dengan autentikasi
	// protected.HandleFunc("/tambah-emas", controller.TambahEmas).Methods("POST")
	// protected.HandleFunc("/emas", controller.GetAllEmas).Methods("GET")
	// protected.HandleFunc("/emas/update", controller.UpdateEmas).Methods("PUT")

	// Endpoint untuk pesanan, termasuk menambah dan mendapatkan pesanan
	// protected.HandleFunc("/orders", controller.HandleAddOrder).Methods("POST")         // Menambahkan order
	// protected.HandleFunc("/orders", controller.HandleGetOrders).Methods("GET")         // Mendapatkan semua pesanan
	// protected.HandleFunc("/orders/{id}", controller.HandleGetOrdersByUserId).Methods("GET") // Mendapatkan pesanan berdasarkan ID


	return router
}
