package model

// OrderRequest digunakan untuk menerima input data dari klien.
type OrderRequest struct {
    JenisPerhiasan   string  `json:"jenis_perhiasan"`
    JenisEmas        string  `json:"jenis_emas"`
    BeratEmas        float64 `json:"berat_emas"`
    CampuranTambahan string  `json:"campuran_tambahan"`
    PersentaseEmas   float64 `json:"persentase_emas"`
    TotalHarga       float64 `json:"total_harga"`
}
