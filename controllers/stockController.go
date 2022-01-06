package controllers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/API-Obat/database"
	"github.com/kamil5b/API-Obat/models"
	"github.com/kamil5b/API-Obat/utils"
)

type stockQuery struct {
	NomorFaktur    int
	TanggalFaktur  time.Time
	NomorStock     int
	KodeBarang     string
	Expired        time.Time
	BigQty         int
	MediumQty      int
	SmallQty       int
	HargaBeliKecil int
}

func GetStock(nomorstock int) models.Stock {
	var tmp stockQuery
	var barang models.Barang
	query := `SELECT stock.NomorStock, faktur.NomorFaktur, 
	faktur.TanggalFaktur, stock.KodeBarang, 
	stock.Expired,stock.BigQty, stock.MediumQty, stock.SmallQty, 
	stock.HargaBeliKecil FROM stock 
	JOIN pembelian ON pembelian.NomorStock = stock.NomorStock 
	JOIN fakturpembelian on fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian 
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur 
	WHERE stock.NomorStock = ?`
	database.DB.Raw(query, nomorstock).Find(&tmp)
	db := database.DB.Table("barang").Where("`KodeBarang` = ?", tmp.KodeBarang).Find(&barang)
	if db.Error != nil {
		fmt.Println(db.Error)
	}
	stock := models.Stock{
		NomorStock:     tmp.NomorStock,
		NomorFaktur:    tmp.NomorFaktur,
		TanggalFaktur:  tmp.TanggalFaktur,
		BarangStock:    barang,
		Expired:        tmp.Expired,
		BigQty:         tmp.BigQty,
		MediumQty:      tmp.MediumQty,
		SmallQty:       tmp.SmallQty,
		HargaBeliKecil: tmp.HargaBeliKecil,
	}
	return stock
}

func GetStockByFaktur(nomorfaktur int) ([]models.Stock, error) {
	var tmp []stockQuery
	var stocks []models.Stock
	query := `SELECT stock.NomorStock, faktur.NomorFaktur, 
	faktur.TanggalFaktur, stock.KodeBarang, stock.Expired,
	stock.BigQty, stock.MediumQty, stock.SmallQty, 
	stock.HargaBeliKecil FROM stock 
	JOIN pembelian ON pembelian.NomorStock = stock.NomorStock 
	JOIN fakturpembelian on fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian 
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur 
	WHERE faktur.NomorFaktur = ? 
	AND (BigQty > 0 || MediumQty > 0 || SmallQty > 0) 
	ORDER BY NomorStock ASC`
	db := database.DB.Raw(query, nomorfaktur).Find(&tmp)
	if db.Error != nil {
		return nil, db.Error
	}
	if len(tmp) == 0 {
		return nil, errors.New("stock habis")
	}
	for _, tmpstock := range tmp {
		var barang models.Barang
		database.DB.Table("barang").Where("`KodeBarang` = ?", tmpstock.KodeBarang).Find(&barang)
		stock := models.Stock{
			NomorStock:     tmpstock.NomorStock,
			NomorFaktur:    tmpstock.NomorFaktur,
			TanggalFaktur:  tmpstock.TanggalFaktur,
			BarangStock:    barang,
			Expired:        tmpstock.Expired,
			BigQty:         tmpstock.BigQty,
			MediumQty:      tmpstock.MediumQty,
			SmallQty:       tmpstock.SmallQty,
			HargaBeliKecil: tmpstock.HargaBeliKecil,
		}
		stocks = append(stocks, stock)
	}
	//fmt.Println(stocks)
	return stocks, nil
}
func GetStockByFakturBarang(nomorfaktur int, kodebarang string) models.Stock {
	var tmp stockQuery
	var stock models.Stock
	var barang models.Barang
	query := `SELECT stock.NomorStock, faktur.NomorFaktur, 
	faktur.TanggalFaktur, stock.KodeBarang, stock.Expired,
	stock.BigQty, stock.MediumQty, stock.SmallQty, 
	stock.HargaBeliKecil FROM stock 
	JOIN pembelian ON pembelian.NomorStock = stock.NomorStock 
	JOIN fakturpembelian on fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian 
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur 
	WHERE faktur.NomorFaktur = ? AND stock.KodeBarang = ?
	AND (BigQty > 0 || MediumQty > 0 || SmallQty > 0) 
	ORDER BY NomorStock ASC`
	database.DB.Raw(query, nomorfaktur, kodebarang).Find(&tmp)
	db := database.DB.Table("barang").Where("`KodeBarang` = ?", tmp.KodeBarang).Find(&barang)
	if db.Error != nil {
		fmt.Println(db.Error)
	}
	stock = models.Stock{
		NomorStock:     tmp.NomorStock,
		NomorFaktur:    tmp.NomorFaktur,
		TanggalFaktur:  tmp.TanggalFaktur,
		BarangStock:    barang,
		Expired:        tmp.Expired,
		BigQty:         tmp.BigQty,
		MediumQty:      tmp.MediumQty,
		SmallQty:       tmp.SmallQty,
		HargaBeliKecil: tmp.HargaBeliKecil,
	}
	return stock
}

func GetStockByKode(kodebarang string) ([]models.Stock, error) {
	var tmp []stockQuery
	var stocks []models.Stock
	query := `SELECT stock.NomorStock, faktur.NomorFaktur, 
	faktur.TanggalFaktur, stock.KodeBarang, stock.Expired,
	stock.BigQty, stock.MediumQty, stock.SmallQty, 
	stock.HargaBeliKecil FROM stock 
	JOIN pembelian ON pembelian.NomorStock = stock.NomorStock 
	JOIN fakturpembelian on fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian 
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur 
	WHERE stock.KodeBarang = ? 
	AND (BigQty > 0 || MediumQty > 0 || SmallQty > 0) 
	ORDER BY NomorStock ASC`
	db := database.DB.Raw(query, kodebarang).Find(&tmp)
	if db.Error != nil {
		return nil, db.Error
	}
	if len(tmp) == 0 {
		return nil, errors.New("stock habis")
	}
	for _, tmpstock := range tmp {
		var barang models.Barang
		database.DB.Table("barang").Where("`KodeBarang` = ?", tmpstock.KodeBarang).Find(&barang)
		stock := models.Stock{
			NomorStock:     tmpstock.NomorStock,
			NomorFaktur:    tmpstock.NomorFaktur,
			TanggalFaktur:  tmpstock.TanggalFaktur,
			BarangStock:    barang,
			Expired:        tmpstock.Expired,
			BigQty:         tmpstock.BigQty,
			MediumQty:      tmpstock.MediumQty,
			SmallQty:       tmpstock.SmallQty,
			HargaBeliKecil: tmpstock.HargaBeliKecil,
		}
		stocks = append(stocks, stock)
	}
	//fmt.Println(stocks)
	return stocks, nil
}

//GET
func GetAllStock(c *fiber.Ctx) error {
	/*
		SELECT stock.NomorStock,
		faktur.NomorFaktur, faktur.TanggalFaktur,
		stock.Expired, stock.KodeBarang,
		stock.BigQty, stock.MediumQty, stock.SmallQty,
		pembelian.HargaBeliKecil
		FROM stock
		join pembelian on pembelian.NomorStock = stock.NomorStock
		join fakturpembelian on pembelian.TransaksiPembelian = fakturpembelian.TransaksiPembelian
		join faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur
		WHERE stock.BigQty > 0 or stock.MediumQty > 0 or stock.SmallQty
	*/
	var stocks []models.Stock

	var tmp []stockQuery
	query := `SELECT stock.NomorStock, faktur.NomorFaktur, 
	faktur.TanggalFaktur, stock.KodeBarang, 
	stock.Expired,stock.BigQty, stock.MediumQty, stock.SmallQty, 
	stock.HargaBeliKecil FROM stock 
	JOIN pembelian ON pembelian.NomorStock = stock.NomorStock 
	JOIN fakturpembelian on fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian 
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur`
	db := database.DB.Raw(query).Find(&tmp)
	if db.Error != nil {
		return db.Error
	}
	for _, tmpstock := range tmp {
		var barang models.Barang
		database.DB.Table("barang").Where("`KodeBarang` = ?", tmpstock.KodeBarang).Find(&barang)
		stock := models.Stock{
			NomorStock:     tmpstock.NomorStock,
			NomorFaktur:    tmpstock.NomorFaktur,
			TanggalFaktur:  tmpstock.TanggalFaktur,
			BarangStock:    barang,
			Expired:        tmpstock.Expired,
			BigQty:         tmpstock.BigQty,
			MediumQty:      tmpstock.MediumQty,
			SmallQty:       tmpstock.SmallQty,
			HargaBeliKecil: tmpstock.HargaBeliKecil,
		}
		stocks = append(stocks, stock)
	}
	return c.JSON(stocks)
}

//GET
func GetStockSummary(c *fiber.Ctx) error {
	/*
		SELECT barang.KodeBarang,barang.NamaBarang,
		SUM(`BigQty`),barang.TipeBigQty,
		SUM(`MediumQty`),barang.TipeMediumQty,
		SUM(`SmallQty`),barang.TipeSmallQty
		FROM `stock` JOIN `barang`
		ON stock.KodeBarang = barang.KodeBarang
		GROUP BY stock.KodeBarang

		1.
		SELECT `KodeBarang`,
		SUM(`BigQty`) as Big,
		SUM(`MediumQty`) as Med,
		SUM(`SmallQty`) as Small
		FROM `stock`
		GROUP BY stock.KodeBarang

		2.
		SELECT * FROM barang
		WHERE KodeBarang = {stock.KodeBarang}
	*/
	type StockSummary struct {
		BarangStock models.Barang
		BigQty      int
		MediumQty   int
		SmallQty    int
	}
	var stocks []StockSummary
	rows, err := database.DB.Table("stock").Select("`KodeBarang`,sum(`BigQty`), sum(`MediumQty`), sum(`SmallQty`)").Group("`KodeBarang`").Rows()
	if err != nil {
		return err
	}
	for rows.Next() {
		var stock StockSummary
		var barang models.Barang
		var kode string
		err = rows.Scan(&kode, &stock.BigQty, &stock.MediumQty, &stock.SmallQty)
		if err != nil {
			return err
		}
		database.DB.Table("barang").Where("`KodeBarang` = ?", kode).Find(&barang)
		stock.BarangStock = barang
		stocks = append(stocks, stock)
	}

	return c.JSON(stocks)
}

//PUT
func UnboxStock(c *fiber.Ctx) error {
	/*
		{
			nomorstock :
			quantity :
			tipequantity :
		}
	*/
	var data map[string]string
	//var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	query := ""
	from, to := 0, 0
	dataint := utils.MapStringToInt(data)
	stock := GetStock(dataint["nomorstock"])
	if data["tipequantity"] == stock.BarangStock.TipeBigQty {
		stock.BigToMedium(dataint["quantity"])
		from = stock.BigQty
		to = stock.MediumQty
		query = "UPDATE `stock` SET BigQty = ?, MediumQty = ? WHERE `stock`.`NomorStock` = ?"
	} else if data["tipequantity"] == stock.BarangStock.TipeMediumQty {
		stock.MediumToSmall(dataint["quantity"])
		from = stock.MediumQty
		to = stock.SmallQty
		query = "UPDATE `stock` SET MediumQty = ?, SmallQty = ? WHERE `stock`.`NomorStock` = ?"
	} else {
		return c.JSON(fiber.Map{
			"message": "invalid type",
		})
	}

	database.DB.Exec(query, from, to, stock.NomorStock)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
