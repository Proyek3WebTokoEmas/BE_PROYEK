package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"proyek3/services"

	"github.com/veritrans/go-midtrans"
)

// Struktur untuk informasi customer (update sesuai dengan format Midtrans)
type CustomerDetails struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// Struktur request pembayaran yang hanya membutuhkan gross_amount dan customer_details
type PaymentRequest struct {
	GrossAmount     int64           `json:"gross_amount"`
	CustomerDetails CustomerDetails `json:"customer_details"`
}

func CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Generate order ID jika tidak diberikan
	orderID := "order-" + time.Now().Format("20060102150405")

	// Dapatkan client dari service
	midtransClient := services.MidtransClient()

	// Prepare Snap transaction request
	snapGateway := midtrans.SnapGateway{Client: *midtransClient}
	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: req.GrossAmount,
		},
		CustomerDetail: &midtrans.CustDetail{
			FName: req.CustomerDetails.FirstName,
			Email: req.CustomerDetails.Email,
			Phone: req.CustomerDetails.Phone,
		},
	}

	// Dapatkan token Snap
	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		http.Error(w, "Failed to create payment", http.StatusInternalServerError)
		return
	}

	// Berikan respons token dan redirect URL
	response := map[string]string{
		"token":        snapResp.Token,
		"redirect_url": snapResp.RedirectURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
