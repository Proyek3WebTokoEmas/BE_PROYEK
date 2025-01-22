package model

type Payment struct {
    ID             uint   `gorm:"primaryKey"`
    OrderID        string `gorm:"unique"`
    TransactionID  string
    Status         string
    // Tambahkan field lain yang diperlukan
}
