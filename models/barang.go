package models

import (
	"fmt"
	"strings"
)

type Barang struct {
	KodeBarang     string
	NamaBarang     string
	TipeBigQty     string
	BigToMedium    int
	TipeMediumQty  string
	MediumToSmall  int
	TipeSmallQty   string
	HargaJualKecil int
	TipeBarang     string
}

/*
	kode:
	nama:
	tipebig:
	btm:
	tipemedium:
	mts:
	tipesmall:
	hargakecil:
	tipebarang:
*/
func (barang Barang) ConvertQty(from, to string, qty int) int {
	fmt.Println(from, barang.TipeMediumQty)
	if strings.EqualFold(from, "big") ||
		strings.EqualFold(from, "besar") ||
		strings.EqualFold(from, strings.ToLower(barang.TipeBigQty)) {
		if strings.EqualFold(to, "medium") ||
			strings.EqualFold(to, "sedang") ||
			strings.EqualFold(to, strings.ToLower(barang.TipeMediumQty)) {
			fmt.Println(barang.BigToMedium, "*", qty, "=", barang.BigToMedium*qty)
			return barang.BigToMedium * qty
		} else if strings.EqualFold(to, "small") ||
			strings.EqualFold(to, "kecil") ||
			strings.EqualFold(to, strings.ToLower(barang.TipeSmallQty)) {
			fmt.Println(barang.BigToMedium, "*", barang.MediumToSmall, "*", qty, "=", barang.BigToMedium*barang.MediumToSmall*qty)
			return barang.BigToMedium * barang.MediumToSmall * qty
		}
	} else if strings.EqualFold(from, "medium") ||
		strings.EqualFold(from, "sedang") ||
		strings.EqualFold(from, strings.ToLower(barang.TipeMediumQty)) {
		if strings.EqualFold(to, "small") ||
			strings.EqualFold(to, "kecil") ||
			strings.EqualFold(to, strings.ToLower(barang.TipeSmallQty)) {
			fmt.Println(barang.MediumToSmall, "*", qty, "=", barang.MediumToSmall*barang.MediumToSmall*qty)
			return barang.MediumToSmall * qty
		}
	} else if strings.EqualFold(from, "small") ||
		strings.EqualFold(from, "kecil") ||
		strings.EqualFold(from, strings.ToLower(barang.TipeSmallQty)) {
		if strings.EqualFold(to, "small") ||
			strings.EqualFold(to, "kecil") ||
			strings.EqualFold(to, strings.ToLower(barang.TipeSmallQty)) {
			return qty
		}
	}
	return 0
}
