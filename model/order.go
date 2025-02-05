package model

// Order merepresentasikan data pesanan
type Order struct {
    ID               int     `json:"id"`
    JenisPerhiasan   string  `json:"jenis_perhiasan"`
    JenisEmas        string  `json:"jenis_emas"`
    BeratEmas        float64 `json:"berat_emas"`
    CampuranTambahan string  `json:"campuran_tambahan"`
    PersentaseEmas   float64 `json:"persentase_emas"`
    TotalHarga       float64 `json:"total_harga"`
    Status           string  `json:"status"`
}
