package controller

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"proyek3/database"
	"proyek3/services"

	"github.com/veritrans/go-midtrans"
)

type CustomerDetails struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type PaymentRequest struct {
	CustomOrderID   int             `json:"custom_order_id"`
	GrossAmount     int64           `json:"gross_amount"`
	CustomerDetails CustomerDetails `json:"customer_details"`
}

func CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req PaymentRequest

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Generate Order ID
	orderID := "order-" + time.Now().Format("20060102150405")

	// Dapatkan client Midtrans
	midtransClient := services.MidtransClient()

	// Prepare Snap transaction request
	snapGateway := midtrans.SnapGateway{Client: *midtransClient}
	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: req.GrossAmount,
		},
		CustomerDetail: &midtrans.CustDetail{
			FName: req.CustomerDetails.Name,
			Email: req.CustomerDetails.Email,
			Phone: req.CustomerDetails.Phone,
		},
	}

	// Dapatkan token Snap
	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		http.Error(w, "Failed to create payment", http.StatusInternalServerError)
		log.Printf("Error getting snap token: %v", err)
		return
	}

	// Simpan pembayaran ke `payments`
	// Pastikan order_id ada di custom_orders sebelum menyimpan di payments
_, err = database.DB.Exec(`
UPDATE custom_orders
SET order_id = $1
WHERE id = $2`,
orderID, req.CustomOrderID) // pastikan req.CustomOrderID adalah ID order yang benar

if err != nil {
log.Printf("Error updating custom_orders with order_id: %v", err)
http.Error(w, "Failed to update custom_orders", http.StatusInternalServerError)
return
}


	// **Update order_id di custom_orders**
	_, err = database.DB.Exec(`
    INSERT INTO payments (
        order_id, gross_amount, customer_name, customer_email, 
        customer_phone, token, redirect_url, status
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')`,
    orderID, req.GrossAmount, req.CustomerDetails.Name, req.CustomerDetails.Email, 
    req.CustomerDetails.Phone, snapResp.Token, snapResp.RedirectURL)

if err != nil {
    log.Printf("Error saving payment data: %v", err)
    http.Error(w, "Failed to save payment data", http.StatusInternalServerError)
    return
}


	// Kirim response token dan redirect URL
	response := map[string]string{
		"token":        snapResp.Token,
		"redirect_url": snapResp.RedirectURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Debug log
	log.Printf("Received Webhook: %s", string(body))

	// Decode JSON payload
	var notificationPayload map[string]interface{}
	if err := json.Unmarshal(body, &notificationPayload); err != nil {
		log.Println("Invalid JSON payload:", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Ambil Order ID
	orderID, ok := notificationPayload["order_id"].(string)
	if !ok {
		log.Println("Order ID tidak valid")
		http.Error(w, "Order ID tidak valid", http.StatusBadRequest)
		return
	}

	// Ambil Status Pembayaran
	transactionStatus, ok := notificationPayload["transaction_status"].(string)
	if !ok {
		log.Println("Status transaksi tidak ditemukan")
		http.Error(w, "Status transaksi tidak valid", http.StatusBadRequest)
		return
	}

	// Jika status "capture", cek fraud_status
	if transactionStatus == "capture" {
		fraudStatus, fraudOk := notificationPayload["fraud_status"].(string)
		if fraudOk && fraudStatus == "accept" {
			transactionStatus = "settlement"
		} else {
			transactionStatus = "failed"
		}
	}

	// Perbarui status pembayaran di tabel payments
	_, err = database.DB.Exec(`
		UPDATE payments
		SET status = $1, updated_at = NOW()
		WHERE order_id = $2`,
		transactionStatus, orderID)

	if err != nil {
		log.Printf("Error updating payments: %v", err)
		http.Error(w, "Gagal memperbarui status pembayaran", http.StatusInternalServerError)
		return
	}

	// **Mapping status Midtrans ke status custom_orders**
	var orderStatus string
	switch transactionStatus {
	case "settlement":
		orderStatus = "paid"
	case "pending":
		orderStatus = "waiting_payment"
	case "deny", "cancel", "expire":
		orderStatus = "failed"
	case "refund":
		orderStatus = "refunded"
	default:
		orderStatus = "unknown"
	}

	// **Update status di custom_orders**
	_, err = database.DB.Exec(`
		UPDATE custom_orders
		SET status = $1, updated_at = NOW()
		WHERE order_id = $2`,
		orderStatus, orderID)

	if err != nil {
		log.Printf("Error updating custom_orders: %v", err)
		http.Error(w, "Gagal memperbarui status pesanan", http.StatusInternalServerError)
		return
	}

	log.Printf("Status pembayaran dan pesanan diperbarui: %s -> %s", orderID, transactionStatus)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment and order status updated"))
}


func GetAllPayments(w http.ResponseWriter, r *http.Request) {
    // Query untuk mengambil data yang diinginkan dari tabel payments
    rows, err := database.DB.Query(`
        SELECT id, jenis_perhiasan, jenis_emas, berat_emas, total_harga, status 
        FROM payments
    `)
    if err != nil {
        log.Printf("Error fetching payments: %v", err)
        http.Error(w, "Failed to fetch payments", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    // Struct untuk menyimpan data pembayaran yang akan dikirim ke frontend
    type Payment struct {
        ID            int     `json:"id"`
        JenisPerhiasan string  `json:"jenis_perhiasan"`
        JenisEmas      string  `json:"jenis_emas"`
        BeratEmas      float64 `json:"berat_emas"`
        TotalHarga     float64 `json:"total_harga"`
        Status         string  `json:"status"`
    }

    var payments []Payment

    // Looping setiap hasil query
    for rows.Next() {
        var p Payment

        // Scan data dari database ke dalam struct
        err := rows.Scan(&p.ID, &p.JenisPerhiasan, &p.JenisEmas, &p.BeratEmas, &p.TotalHarga, &p.Status)
        if err != nil {
            log.Printf("Error scanning payment: %v", err)
            http.Error(w, "Failed to parse payments", http.StatusInternalServerError)
            return
        }

        payments = append(payments, p)
    }

    // Cek apakah ada error setelah iterasi
    if err := rows.Err(); err != nil {
        log.Printf("Error iterating payments: %v", err)
        http.Error(w, "Failed to process payments", http.StatusInternalServerError)
        return
    }

    // Kirim response dalam format JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(payments)
}

