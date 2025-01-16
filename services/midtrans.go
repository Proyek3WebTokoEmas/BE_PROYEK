package services

import (
	"github.com/veritrans/go-midtrans"
)

var (
	serverKey = "SB-Mid-server-Et1YEh9uv5uulLqDOyNVI6xJ" // Ganti dengan Server Key Sandbox Anda
	clientKey = "SB-Mid-client-5PeEj9gpVGmvsFD2"         // Ganti dengan Client Key Sandbox Anda
)

// MidtransClient menginisialisasi client Midtrans
func MidtransClient() *midtrans.Client {
	c := midtrans.NewClient()
	c.ServerKey = serverKey
	c.ClientKey = clientKey
	c.APIEnvType = midtrans.Sandbox // Gunakan Sandbox untuk testing, ubah ke Production untuk live
	return &c                       // Mengembalikan pointer ke client
}
