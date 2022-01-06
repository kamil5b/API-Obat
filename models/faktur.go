package models

import "time"

type Faktur struct {
	NomorUrut      int
	NomorFaktur    int
	TanggalFaktur  time.Time
	TipeTransaksi  string
	TipePembayaran string
	NomorEntitas   int
	JatuhTempo     time.Time
	NomorGiro      string
}
