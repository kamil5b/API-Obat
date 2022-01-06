package models

import "time"

type Hutang struct {
	NomorHutang  int
	FakturHutang Faktur
	TokoDihutang Toko
	NominalRetur int
	SisaHutang   int
	JatuhTempo   time.Time
}
