package models

type Sales struct {
	NomorSales      int
	Karyawan        User
	FakturPenjualan Faktur
	TotalPenjualan  int
	Insentif        string
}
