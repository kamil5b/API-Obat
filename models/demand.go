package models

type Demand struct {
	ID               int
	Barang           Barang
	QuantityDemand   int
	QuantityThen     int
	QuantityRightNow int
	TipeQuantity     string
}
