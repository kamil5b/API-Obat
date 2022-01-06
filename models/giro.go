package models

import "time"

type Giro struct {
	NomorGiro   string
	Nominal     int
	TanggalGiro time.Time
	BankGiro    Bank
}
