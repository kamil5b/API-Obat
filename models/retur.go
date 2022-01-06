package models

type Retur struct {
	FakturBarang   Faktur
	Status         string
	BarangRetur    Barang
	Quantity       int
	TipeQuantity   string
	DiskontilRetur int
	TotalNominal   int
	Description    string
}
