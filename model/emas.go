package model

// Struktur barang emas
type Emas struct {
	ID      uint    `json:"id" gorm:"primaryKey"`
	Nama    string  `json:"nama"`
	Karatan int     `json:"karatan"`
	Berat   float64 `json:"berat"`
	Harga   float64 `json:"harga"`
}
