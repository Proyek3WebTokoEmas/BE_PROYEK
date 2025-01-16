package services

import (
	"proyek3/model"
)

func CalculatePrice(order model.Order) float64 {
	var hargaEmas, hargaCampuran float64

	switch order.JenisEmas {
	case "Emas 18K":
		hargaEmas = 1485700
	case "Emas 22K":
		hargaEmas = 3121420
	case "Emas 20K":
		hargaEmas = 1485700
	}

	switch order.CampuranTambahan {
	case "Perak":
		hargaCampuran = 500000
	case "Palladium":
		hargaCampuran = 600000
	case "Platinum":
		hargaCampuran = 700000
	}

	return (hargaEmas * order.BeratEmas * order.PersentaseEmas / 100) + hargaCampuran
}
