package models

import (
	"errors"
	"strings"
	"time"
)

type Stock struct {
	NomorFaktur    int
	TanggalFaktur  time.Time
	NomorStock     int
	BarangStock    Barang
	Expired        time.Time
	BigQty         int
	MediumQty      int
	SmallQty       int
	HargaBeliKecil int
}

/*
	nomorfaktur:
	tanggal:
	nomorstock:
	barang:{
		kode:
		nama:
		tipebig:
		btm:
		tipemedium:
		mts:
		tipesmall:
		hargakecil:
		tipebarang:
	}
	expired:
	bigqty:
	medqty:
	smallqty:
	hargabeli:
*/

func (S *Stock) Unbox(from, to string, qty int) (int, error) {

	if S.BigQty > 0 &&
		(strings.EqualFold(from, "big") ||
			strings.EqualFold(from, "besar") ||
			strings.EqualFold(from, strings.ToLower(S.BarangStock.TipeBigQty))) {
		if strings.EqualFold(to, "medium") ||
			strings.EqualFold(to, "sedang") ||
			strings.EqualFold(to, strings.ToLower(S.BarangStock.TipeMediumQty)) {
			return (*S).BigToMedium(qty)
		}
	} else if S.MediumQty > 0 &&
		(strings.EqualFold(from, "medium") ||
			strings.EqualFold(from, "sedang") ||
			strings.EqualFold(from, strings.ToLower(S.BarangStock.TipeMediumQty))) {
		if strings.EqualFold(to, "small") ||
			strings.EqualFold(to, "kecil") ||
			strings.EqualFold(to, strings.ToLower(S.BarangStock.TipeSmallQty)) {
			return (*S).MediumToSmall(qty)
		}
	}
	return qty, errors.New("cannot unbox")
}

func (S *Stock) BigToMedium(qty int) (int, error) {
	if S.BigQty > 0 {
		if S.BigQty >= qty {
			(*S).MediumQty += qty * S.BarangStock.BigToMedium
			(*S).BigQty = S.BigQty - qty
			return 0, nil
		} else {
			tmp := qty - S.BigQty
			(*S).MediumQty += S.BigQty * S.BarangStock.BigToMedium
			(*S).BigQty = 0
			return tmp, nil
		}
	} else {
		return qty, errors.New("big qty is 0")
	}
}

func (S *Stock) MediumToSmall(qty int) (int, error) {
	if S.MediumQty > 0 {
		if S.MediumQty >= qty {
			(*S).SmallQty += qty * S.BarangStock.MediumToSmall
			(*S).MediumQty = S.MediumQty - qty
			return 0, nil
		} else {
			tmp := qty - S.MediumQty
			(*S).SmallQty += S.MediumQty * S.BarangStock.MediumToSmall
			(*S).MediumQty = 0
			return tmp, nil
		}
	} else {
		return qty, errors.New("medium qty is 0")
	}
}

func (S Stock) ConvertQty(from, to string) int {
	if strings.EqualFold(from, "big") ||
		strings.EqualFold(from, "besar") ||
		strings.EqualFold(from, strings.ToLower(S.BarangStock.TipeBigQty)) {
		return S.BarangStock.ConvertQty(from, to, S.BigQty)
	} else if strings.EqualFold(from, "medium") ||
		strings.EqualFold(from, "sedang") ||
		strings.EqualFold(from, strings.ToLower(S.BarangStock.TipeMediumQty)) {
		return S.BarangStock.ConvertQty(from, to, S.MediumQty)
	}
	return 0
}
