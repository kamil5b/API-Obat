package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/API-Obat/database"
	"github.com/kamil5b/API-Obat/models"
	"github.com/kamil5b/API-Obat/utils"
)

type sumpenjualan struct {
	NomorFaktur    int
	TanggalFaktur  time.Time
	TotalDiskontil int
	TotalPenjualan int
}
type summarypenjualan struct {
	Details        []sumpenjualan
	TotalDiskontil int
	TotalPenjualan int
}

type jual struct {
	NomorFaktur        int
	TanggalFaktur      time.Time
	KodeBarang         string
	NamaBarang         string
	Expired            time.Time
	Quantity           int
	TipeQuantity       string
	HargaJualKecil     int
	TipePembayaran     string
	DiskontilPenjualan int
	TotalHarga         int
	NomorCustomer      int
}

//POST
func JualBarang(c *fiber.Ctx) error {
	/*
		{
			nik :
			nomorfaktur:
			quantity:
			tipequantity:
			tipepembayaran:
			kodebarang:
			diskontil:
			nomorcustomer:
			jatuhtempo:
		}

		Alur :
		1. Toko udah diregister
		2. Barang udah diregister
		3. Faktur sudah dibuat
		3a. Nomor giro udah di register
		4. Buat stock dulu!
		4. Buat penjualan

	*/
	var data map[string]string
	var faktur models.Faktur
	var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	db := database.DB.Table("faktur").Where("`NomorFaktur` = ?", data["nomorfaktur"]).Scan(&faktur)
	if db.Error != nil {
		fmt.Println(db.Error)
		return c.JSON(fiber.Map{
			"message": "nomor faktur belum terdaftar",
		})
	}
	tipeqty := ""
	database.DB.Table("barang").Where("`KodeBarang` = ?", data["kodebarang"]).Find(&barang)
	if strings.EqualFold(data["tipequantity"], barang.TipeBigQty) {
		tipeqty = "BigQty"
	} else if strings.EqualFold(data["tipequantity"], barang.TipeMediumQty) {
		tipeqty = "MediumQty"
	} else if strings.EqualFold(data["tipequantity"], barang.TipeSmallQty) {
		tipeqty = "SmallQty"
	} else {
		return c.JSON(fiber.Map{
			"message": "tipe quantity tidak ditemukan pada barang",
		})
	}

	dataint := utils.MapStringToInt(data)

	qtystock := 0
	selection := "sum(`" + tipeqty + "`)"
	database.DB.Table("stock").Select(selection).Where("`KodeBarang` = ?", data["kodebarang"]).Find(&qtystock)
	stocks, err := GetStockByKode(data["kodebarang"])
	if err != nil {
		query := "INSERT INTO demands(KodeBarang, QuantityDemand, QuantityAvailable, TipeQuantity) VALUES (?,?,?,?)"
		database.DB.Exec(query,
			data["kodebarang"],
			dataint["quantity"],
			qtystock,
			data["tipequantity"],
		)
		return c.JSON(fiber.Map{
			"message": "stock untuk barang ini habis",
		})
	}
	process := func(stock models.Stock, qty int) error {
		totalhargajual := 0
		totalhargabeli := 0
		if stock.BarangStock.ConvertQty(data["tipequantity"],
			"small", dataint["quantity"]) != 0 {
			totalhargajual = (stock.BarangStock.ConvertQty(
				data["tipequantity"],
				stock.BarangStock.TipeSmallQty,
				qty,
			) * stock.BarangStock.HargaJualKecil) - dataint["diskontil"]
			totalhargabeli = (stock.BarangStock.ConvertQty(
				data["tipequantity"],
				stock.BarangStock.TipeSmallQty,
				qty,
			) * stock.HargaBeliKecil)
		} else {
			return c.JSON(fiber.Map{
				"message": "tipe quantity tidak ditemukan pada barang",
			})
		}
		query := `INSERT INTO profits(NomorFaktur, TanggalFaktur, 
			KodeBarang, TotalPembelian, TotalPenjualan, TotalProfit) 
			VALUES (?,?,?,?,?,?)`
		db = database.DB.Exec(query,
			faktur.NomorFaktur,
			faktur.TanggalFaktur,
			data["kodebarang"],
			totalhargabeli,
			totalhargajual,
			totalhargajual-totalhargabeli,
		)
		if db.Error != nil {
			return c.JSON(fiber.Map{
				"message": "profit insertion error",
			})
		}
		query = `INSERT INTO penjualan(NomorStock, 
		Quantity, TipeQuantity, DiskontilPenjualan, 
		TotalHarga) VALUES (?,?,?,?,?)`
		db = database.DB.Exec(query,
			stock.NomorStock,
			qty,
			data["tipequantity"],
			dataint["diskontil"],
			totalhargajual,
		)

		if db.Error != nil {
			return c.JSON(fiber.Map{
				"message": "penjualan error 1",
			})
		}
		err := InsertSales(data["nik"], faktur, totalhargajual)
		if err != nil {
			query = "delete from  penjualan order by TransaksiPenjualan desc limit 1"
			database.DB.Exec(query)
			query = "delete from  profits order by NomorProfit desc limit 1"
			database.DB.Exec(query)
			return c.JSON(fiber.Map{
				"message": "sales error",
			})
		}
		var tipetransaksi string
		query = "SELECT `TipePembayaran` FROM faktur WHERE `NomorFaktur`=?"
		database.DB.Raw(query, dataint["nomorfaktur"]).Find(&tipetransaksi)
		if tipetransaksi == "KREDIT" {
			PiutangBarang(dataint["nomorfaktur"], dataint["diskontil"], totalhargajual)
		}
		var notransaksi int
		query = "SELECT `TransaksiPenjualan` FROM penjualan ORDER BY `TransaksiPenjualan` DESC LIMIT 1"
		database.DB.Raw(query).Find(&notransaksi)
		query = "INSERT INTO `fakturpenjualan`(`NomorFaktur`, `TransaksiPenjualan`) VALUES (?,?)"
		db = database.DB.Exec(
			query,
			dataint["nomorfaktur"],
			notransaksi,
		)
		if db.Error != nil {
			query = "delete from  penjualan order by TransaksiPenjualan desc limit 1"
			database.DB.Exec(query)
			query = "delete from  profits order by NomorProfit desc limit 1"
			database.DB.Exec(query)
			return c.JSON(fiber.Map{
				"message": "penjualan error 2",
			})
		}
		return nil
	}
	tmpqty := dataint["quantity"]
	stockprocess := func(bmsqty int, stock models.Stock) (int, error) {

		qty := 0
		if bmsqty >= tmpqty {
			qty = bmsqty - tmpqty
			err = process(stock, tmpqty)
			if err != nil {
				return qty, err
			}
			tmpqty = 0
		} else if bmsqty > 0 {
			//stock.BigQty < tmpqty
			tmpqty -= bmsqty
			err = process(stock, bmsqty)
			if err != nil {
				return qty, err
			}
		} else {
			return tmpqty, nil
		}
		return qty, nil
	}
	if qtystock < dataint["quantity"] {
		query := "INSERT INTO demands(KodeBarang, QuantityDemand, QuantityAvailable, TipeQuantity) VALUES (?,?,?,?)"
		database.DB.Exec(query,
			data["kodebarang"],
			dataint["quantity"],
			qtystock,
			data["tipequantity"],
		)
		fmt.Println("stock kurang")
		return c.JSON(fiber.Map{
			"message": "stock barang kurang",
		})
	}
	for _, stock := range stocks {

		qty := 0
		query := ""
		if tipeqty == "BigQty" {
			qty, err = stockprocess(stock.BigQty, stock)
			if err != nil {
				return err
			}
			query = "UPDATE `stock` SET BigQty = ? WHERE `stock`.`NomorStock` = ?"
		} else if tipeqty == "MediumQty" {
			qty, err = stockprocess(stock.MediumQty, stock)
			if err != nil {
				return err
			}
			query = "UPDATE `stock` SET MediumQty = ? WHERE `stock`.`NomorStock` = ?"
		} else if tipeqty == "SmallQty" {
			qty, err = stockprocess(stock.SmallQty, stock)
			if err != nil {
				return err
			}
			query = "UPDATE `stock` SET SmallQty = ? WHERE `stock`.`NomorStock` = ?"
		}
		if qty != tmpqty {
			db = database.DB.Exec(query, qty, stock.NomorStock)
		}
		if db.Error != nil {
			fmt.Println(db.Error)
			return db.Error
		}
		if tmpqty == 0 {
			break
		}
	}
	if tmpqty != 0 {
		msg := "success with " + strconv.Itoa(dataint["quantity"]-tmpqty) + " data proceed"
		return c.JSON(fiber.Map{
			"message": msg,
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

//GET
func GetAllPenjualan(c *fiber.Ctx) error {

	type getjual struct {
		NomorFaktur        int
		TanggalFaktur      time.Time
		KodeBarang         string
		NamaBarang         string
		Expired            time.Time
		Quantity           int
		TipeQuantity       string
		HargaJualKecil     int
		TipePembayaran     string
		DiskontilPenjualan int
		TotalHarga         int
		NamaCustomer       string
	}
	var penjualan []jual
	var gets []getjual
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
	faktur.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga, 
	faktur.NomorEntitas FROM penjualan 
	JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan 
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur 
	JOIN stock ON penjualan.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang`
	db := database.DB.Raw(query).Find(&penjualan)
	if db.Error != nil {
		return db.Error
	}
	for _, p := range penjualan {
		get := getjual{
			NomorFaktur:        p.NomorFaktur,
			TanggalFaktur:      p.TanggalFaktur,
			KodeBarang:         p.KodeBarang,
			NamaBarang:         p.NamaBarang,
			Expired:            p.Expired,
			Quantity:           p.Quantity,
			TipeQuantity:       p.TipeQuantity,
			HargaJualKecil:     p.HargaJualKecil,
			TipePembayaran:     p.TipePembayaran,
			DiskontilPenjualan: p.DiskontilPenjualan,
			TotalHarga:         p.TotalHarga,
			NamaCustomer:       "-",
		}
		if p.NomorCustomer > 0 {
			var tmp string
			database.DB.Table("`customer`").Select("`NamaCustomer`").Where("`NomorUrut`=?", p.NomorCustomer).Find(&tmp)
			get.NamaCustomer = tmp
		}
		gets = append(gets, get)
	}

	return c.JSON(gets)
}

//POST
func GetPenjualanPerFaktur(c *fiber.Ctx) error {

	type getjual struct {
		NomorFaktur        int
		TanggalFaktur      time.Time
		KodeBarang         string
		NamaBarang         string
		Expired            time.Time
		Quantity           int
		TipeQuantity       string
		HargaJualKecil     int
		TipePembayaran     string
		DiskontilPenjualan int
		TotalHarga         int
		NamaCustomer       string
	}
	var penjualan []jual
	var gets []getjual
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
	faktur.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga, 
	faktur.NomorEntitas FROM penjualan 
	JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan 
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur 
	JOIN stock ON penjualan.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE faktur.NomorFaktur = ?`
	db := database.DB.Raw(query, dataint["nomor"]).Find(&penjualan)
	if db.Error != nil {
		return db.Error
	}
	for _, p := range penjualan {
		get := getjual{
			NomorFaktur:        p.NomorFaktur,
			TanggalFaktur:      p.TanggalFaktur,
			KodeBarang:         p.KodeBarang,
			NamaBarang:         p.NamaBarang,
			Expired:            p.Expired,
			Quantity:           p.Quantity,
			TipeQuantity:       p.TipeQuantity,
			HargaJualKecil:     p.HargaJualKecil,
			TipePembayaran:     p.TipePembayaran,
			DiskontilPenjualan: p.DiskontilPenjualan,
			TotalHarga:         p.TotalHarga,
			NamaCustomer:       "-",
		}
		if p.NomorCustomer > 0 {
			var tmp string
			database.DB.Table("`customer`").Select("`NamaCustomer`").Where("`NomorUrut`=?", p.NomorCustomer).Find(&tmp)
			get.NamaCustomer = tmp
		}
		gets = append(gets, get)
	}

	return c.JSON(gets)
}

//POST FAKTUR PEMBELIAN
func FakturPenjualan(c *fiber.Ctx) error {
	var data map[string]string
	/*
		"nomor" : ""
		"tanggal" : ""
		"tipepembayaran" : ""
		"nomorcustomer" : ""
		"giro":""
		"jatuhtempo" : ""
	*/
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	nomorcustomer, _ := strconv.Atoi(data["nomorcustomer"])
	return c.JSON(InsertFaktur(data, nomorcustomer, "PENJUALAN"))
}

//GET FAKTUR
func GetFakturPenjualan(c *fiber.Ctx) error {
	/*
		"nomor" : ""
		"tanggal" : ""
	*/
	var faktur []models.Faktur
	database.DB.Table("faktur").Where("`TipeTransaksi` = \"PENJUALAN\"").Find(&faktur)

	return c.JSON(faktur)
}

//POST PENJUALAN PER BARANG
func PenjualanPerBarang(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	var penjualan []jual
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
	faktur.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga, 
	faktur.NomorEntitas AS NomorCustomer FROM penjualan 
	JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan 
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur 
	JOIN stock ON penjualan.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE stock.KodeBarang = ?`
	db := database.DB.Raw(query, data["kodebarang"]).Find(&penjualan)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(penjualan)
}

//-----SUMMARY------

//GET SUMMARY
func SummaryPenjualan(c *fiber.Ctx) error {
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
		SUM(penjualan.TotalHarga) AS TotalPenjual
		FROM penjualan JOIN fakturpenjualan on
		fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
		GROUP BY faktur.NomorFaktur
	*/
	var sums []sumpenjualan
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjualan
	FROM penjualan JOIN fakturpenjualan on
	fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	GROUP BY faktur.NomorFaktur`

	database.DB.Raw(query).Find(&sums)
	sum := summarypenjualan{
		Details:        sums,
		TotalDiskontil: 0,
		TotalPenjualan: 0,
	}
	for _, s := range sums {
		sum.TotalDiskontil += s.TotalDiskontil
		sum.TotalPenjualan += s.TotalPenjualan
	}
	return c.JSON(sum)
}

//POST SUMMARY PER FAKTUR
func SummaryPenjualanPerFaktur(c *fiber.Ctx) error {
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
		SUM(penjualan.TotalHarga) AS TotalPenjual
		FROM penjualan JOIN fakturpenjualan on
		fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
		GROUP BY faktur.NomorFaktur
	*/
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)

	var sum sumpenjualan
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjualan
	FROM penjualan JOIN fakturpenjualan on
	fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE faktur.NomorFaktur = ? AND faktur.TipeTransaksi = "PENJUALAN"`

	database.DB.Raw(query, dataint["nomorfaktur"]).Find(&sum)

	return c.JSON(sum)
}

//POST SUMMARY PER TANGGAL
func SummaryPenjualanPerTanggal(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}

	tanggalawal, err := utils.ParsingDate(data["tanggalawal"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	tanggalakhir, err := utils.ParsingDate(data["tanggalakhir"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	var sums []sumpenjualan
	query := `SELECT faktur.TanggalFaktur,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjualan
	FROM penjualan JOIN fakturpenjualan on
	fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE (faktur.TanggalFaktur BETWEEN ? AND ?)
	GROUP BY faktur.TanggalFaktur `

	database.DB.Raw(query, tanggalawal, tanggalakhir).Find(&sums)
	sum := summarypenjualan{
		Details:        sums,
		TotalDiskontil: 0,
		TotalPenjualan: 0,
	}
	for _, s := range sums {
		sum.TotalDiskontil += s.TotalDiskontil
		sum.TotalPenjualan += s.TotalPenjualan
	}
	return c.JSON(sum)
}

//GET SUMMARY PER BARANG
func SummaryPenjualanPerBarang(c *fiber.Ctx) error {
	type sumpenjualan struct {
		KodeBarang     string
		TotalDiskontil int
		TotalPenjualan int
	}
	type quantityjual struct {
		TotalQuantity int
		TipeQuantity  string
	}
	type penjualan struct {
		BarangJual     models.Barang
		TotalSmallQty  int
		TotalMediumQty int
		TotalBigQty    int
		TotalDiskontil int
		TotalPenjualan int
	}
	query := `SELECT barang.KodeBarang,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjual FROM penjualan
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	GROUP BY barang.KodeBarang`
	var sums []sumpenjualan
	var pens []penjualan
	database.DB.Raw(query).Find(&sums)
	for _, s := range sums {
		var q []quantityjual
		query = `SELECT SUM(penjualan.Quantity) AS TotalQuantity, 
		LOWER(penjualan.TipeQuantity) AS TipeQuantity FROM penjualan 
		JOIN stock ON penjualan.NomorStock = stock.NomorStock 
		JOIN barang ON barang.KodeBarang = stock.KodeBarang 
		WHERE barang.KodeBarang = ? GROUP BY LOWER(penjualan.TipeQuantity)`
		database.DB.Raw(query, s.KodeBarang).Find(&q)
		barang := GetBarang(s.KodeBarang)
		pen := penjualan{
			BarangJual:     barang,
			TotalSmallQty:  0,
			TotalMediumQty: 0,
			TotalBigQty:    0,
			TotalDiskontil: s.TotalDiskontil,
			TotalPenjualan: s.TotalPenjualan,
		}
		for _, x := range q {
			if strings.EqualFold(barang.TipeBigQty, x.TipeQuantity) {
				pen.TotalBigQty += x.TotalQuantity
			} else if strings.EqualFold(barang.TipeMediumQty, x.TipeQuantity) {
				pen.TotalMediumQty += x.TotalQuantity
			} else if strings.EqualFold(barang.TipeSmallQty, x.TipeQuantity) {
				pen.TotalSmallQty += x.TotalQuantity
			} else {
				return c.JSON(fiber.Map{
					"message": "tipe quantity invalid",
				})
			}
		}
		pens = append(pens, pen)
	}
	return c.JSON(pens)
}

//POST SUMMARY BARANG TANGGAL
func SummaryPenjualanPerBarangTanggal(c *fiber.Ctx) error {
	type sumpenjualan struct {
		KodeBarang     string
		TotalDiskontil int
		TotalPenjualan int
	}
	type quantityjual struct {
		TotalQuantity int
		TipeQuantity  string
	}
	type penjualan struct {
		BarangJual     models.Barang
		TotalSmallQty  int
		TotalMediumQty int
		TotalBigQty    int
		TotalDiskontil int
		TotalPenjualan int
	}
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	tanggalawal, err := utils.ParsingDate(data["tanggalawal"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	tanggalakhir, err := utils.ParsingDate(data["tanggalakhir"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	query := `SELECT barang.KodeBarang,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjualan
	FROM penjualan JOIN fakturpenjualan on
	fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE (faktur.TanggalFaktur BETWEEN ? AND ?)
	GROUP BY barang.KodeBarang`
	var sums []sumpenjualan
	var pens []penjualan
	database.DB.Raw(query, tanggalawal, tanggalakhir).Find(&sums)
	for _, s := range sums {
		var q []quantityjual
		query = `SELECT SUM(penjualan.Quantity) AS TotalQuantity, 
		LOWER(penjualan.TipeQuantity) AS TipeQuantity
		FROM penjualan JOIN fakturpenjualan on
		fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
		WHERE (faktur.TanggalFaktur BETWEEN ? AND ?) AND barang.KodeBarang = ? 
		GROUP BY LOWER(penjualan.TipeQuantity)`
		database.DB.Raw(query, tanggalawal, tanggalakhir, s.KodeBarang).Find(&q)
		barang := GetBarang(s.KodeBarang)
		pen := penjualan{
			BarangJual:     barang,
			TotalSmallQty:  0,
			TotalMediumQty: 0,
			TotalBigQty:    0,
			TotalDiskontil: s.TotalDiskontil,
			TotalPenjualan: s.TotalPenjualan,
		}
		for _, x := range q {
			if strings.EqualFold(barang.TipeBigQty, x.TipeQuantity) {
				pen.TotalBigQty += x.TotalQuantity
			} else if strings.EqualFold(barang.TipeMediumQty, x.TipeQuantity) {
				pen.TotalMediumQty += x.TotalQuantity
			} else if strings.EqualFold(barang.TipeSmallQty, x.TipeQuantity) {
				pen.TotalSmallQty += x.TotalQuantity
			} else {
				return c.JSON(fiber.Map{
					"message": "tipe quantity invalid",
				})
			}
		}
		pens = append(pens, pen)
	}
	if pens == nil {
		return c.JSON(fiber.Map{
			"message": "laporan keuangan tidak ditemukan",
		})
	}
	return c.JSON(pens)
}

//----------------------------

//
func TotalSummaryPenjualan(c *fiber.Ctx) error {

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	template := `SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil, 
	SUM(penjualan.TotalHarga) AS Total FROM penjualan 
	JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan 
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur 
	JOIN stock ON penjualan.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang 
	GROUP BY `

	return c.JSON(TotalSummary(template, data["per"]))
}
