package models

import "time"

type Piutang struct {
	NomorPiutang    int
	FakturPiutang   Faktur
	CustomerPiutang Customer
	NominalRetur    int
	SisaPiutang     int
	JatuhTempo      time.Time
}
