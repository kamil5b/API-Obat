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

//POST
func BeliBarang(c *fiber.Ctx) error {
	/*
		{
			nomorfaktur:
			quantity:
			tipequantity:
			tipepembayaran:
			nomortoko:
			kodebarang:
			expired:
			hargabeli:
			diskontil:
			jatuhtempo:
		}

		Alur :
		1. Toko udah diregister
		2. Barang udah diregister
		3. Faktur sudah dibuat
		3a. Nomor giro udah di register
		4. Buat stock dulu!
		4. Buat pembelian

	*/
	var data map[string]string
	var nostock int
	var faktur models.Faktur
	var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)

	db := database.DB.Table("faktur").Where("`NomorFaktur` = ?", dataint["nomorfaktur"]).Scan(&faktur)
	if db.Error != nil {
		fmt.Println(db.Error)
		return c.JSON(fiber.Map{
			"message": "nomor faktur belum terdaftar",
		})
	}

	expired, _ := utils.ParsingDate(data["expired"])
	query := "SELECT `NomorStock` FROM stock ORDER BY `NomorStock` DESC LIMIT 1"
	database.DB.Raw(query).Find(&nostock)
	nostock++
	db = database.DB.Table("barang").Where("`KodeBarang` = ?", data["kodebarang"]).Find(&barang)
	if db.Error != nil {
		return c.JSON(fiber.Map{
			"message": "barang belum terdaftar",
		})
	}
	totalhargakecil := 0
	bigqty, medqty, smallqty := 0, 0, 0
	if strings.EqualFold(data["tipequantity"], strings.ToLower(barang.TipeBigQty)) {
		bigqty = dataint["quantity"]

	} else if strings.EqualFold(data["tipequantity"], strings.ToLower(barang.TipeMediumQty)) {
		medqty = dataint["quantity"]
	} else if strings.EqualFold(data["tipequantity"], strings.ToLower(barang.TipeSmallQty)) {
		smallqty = dataint["quantity"]
	} else {
		return c.JSON(fiber.Map{
			"message": "tipe quantity tidak ditemukan pada barang",
		})
	}
	totalhargakecil = (dataint["hargabeli"] - dataint["diskontil"]) / (barang.ConvertQty(
		data["tipequantity"],
		"small",
		dataint["quantity"],
	))
	query = `INSERT INTO stock(NomorStock, KodeBarang, Expired, 
		BigQty, MediumQty, SmallQty, HargaBeliKecil) 
		VALUES (?,?,?,?,?,?,?)`
	db = database.DB.Exec(query,
		nostock,
		barang.KodeBarang,
		expired,
		bigqty,
		medqty,
		smallqty,
		totalhargakecil,
	)
	if db.Error != nil {
		return c.JSON(fiber.Map{
			"message": "error restock barang",
		})
	}

	query = `INSERT INTO pembelian(Quantity,
		TipeQuantity, DiskontilPembelian,
		TotalHargaBeli, NomorStock) VALUES (?,?,?,?,?)`
	db = database.DB.Exec(
		query,
		dataint["quantity"],
		data["tipequantity"],
		dataint["diskontil"],
		dataint["hargabeli"]-dataint["diskontil"],
		nostock,
	)
	var tipetransaksi string
	query = "SELECT `TipePembayaran` FROM faktur WHERE `NomorFaktur`=?"
	database.DB.Raw(query, dataint["nomorfaktur"]).Find(&tipetransaksi)
	if tipetransaksi == "KREDIT" {
		HutangBarang(dataint["nomorfaktur"], dataint["diskontil"], dataint["hargabeli"]-dataint["diskontil"])
	}
	if db.Error != nil {
		query = "delete from stock order by NomorStock desc limit 1"
		database.DB.Exec(query)
		return c.JSON(fiber.Map{
			"message": "pembelian error",
		})
	}
	var notransaksi int
	query = "SELECT `TransaksiPembelian` FROM pembelian ORDER BY `TransaksiPembelian` DESC LIMIT 1"
	database.DB.Raw(query).Scan(&notransaksi)

	query = "INSERT INTO `fakturpembelian`(`NomorFaktur`, `TransaksiPembelian`) VALUES (?,?)"
	db = database.DB.Exec(
		query,
		dataint["nomorfaktur"],
		notransaksi,
	)
	if db.Error != nil {
		query = "delete from pembelian order by TransaksiPembelian desc limit 1"
		database.DB.Exec(query)
		query = "delete from stock order by NomorStock desc limit 1"
		database.DB.Exec(query)
		return c.JSON(fiber.Map{
			"message": "pembelian error",
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

type beli struct {
	NomorFaktur        int
	TanggalFaktur      time.Time
	KodeBarang         string
	NamaBarang         string
	Expired            time.Time
	Quantity           int
	TipeQuantity       string
	HargaBeliKecil     int
	TipePembayaran     string
	NamaToko           string
	DiskontilPembelian int
	TotalHargaBeli     int
}

//GET HISTORY PEMBELIAN
func GetAllPembelian(c *fiber.Ctx) error {
	var pembelian []beli
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	pembelian.Quantity,pembelian.TipeQuantity, stock.HargaBeliKecil,
	faktur.TipePembayaran, toko.NamaToko,pembelian.DiskontilPembelian, 
	pembelian.TotalHargaBeli FROM pembelian 
	JOIN fakturpembelian on fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian 
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur 
	JOIN stock ON pembelian.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang 
	JOIN toko ON faktur.NomorEntitas = toko.NomorToko`
	db := database.DB.Raw(query).Find(&pembelian)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(pembelian)
}

//POST FAKTUR PEMBELIAN
func FakturPembelian(c *fiber.Ctx) error {
	var data map[string]string
	/*
		"nomor" : ""
		"tanggal" : ""
		"tipepembayaran" : ""
		"giro":""
		"jatuhtempo" : ""
	*/
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	nomortoko, _ := strconv.Atoi(data["nomortoko"])
	return c.JSON(InsertFaktur(data, nomortoko, "PEMBELIAN"))
}

//GET FAKTUR
func GetFakturPembelian(c *fiber.Ctx) error {
	/*
		"nomor" : ""
		"tanggal" : ""
	*/
	var faktur []models.Faktur
	database.DB.Table("faktur").Where("`TipeTransaksi` = \"PEMBELIAN\"").Find(&faktur)

	return c.JSON(faktur)
}

//-----NEW-----

type sumpembelian struct {
	NomorFaktur    int
	TanggalFaktur  time.Time
	TotalDiskontil int
	TotalPembelian int
}
type summarypembelian struct {
	Details        []sumpembelian
	TotalDiskontil int
	TotalPembelian int
}

//POST PEMBELIAN PER FAKTUR
func PembelianPerFaktur(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	var pembelian []beli
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	faktur.TipePembayaran,faktur.NomorEntitas AS NomorToko , stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	pembelian.Quantity,pembelian.TipeQuantity, barang.HargaJualKecil,
	pembelian.DiskontilPembelian, pembelian.TotalHargaBeli FROM pembelian 
	JOIN fakturpembelian on fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian 
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur 
	JOIN stock ON pembelian.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE faktur.NomorFaktur = ?`
	db := database.DB.Raw(query, dataint["nomorfaktur"]).Find(&pembelian)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(pembelian)
}

//POST PEMBELIAN PER BARANG
func PembelianPerBarang(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	var pembelian []beli
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	faktur.TipePembayaran, stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	pembelian.Quantity,pembelian.TipeQuantity, barang.HargaJualKecil,
	pembelian.DiskontilPembelian, pembelian.TotalHargaBeli, 
	faktur.NomorEntitas AS NomorToko FROM pembelian 
	JOIN fakturpembelian on fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian 
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur 
	JOIN stock ON pembelian.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE stock.KodeBarang = ?`
	db := database.DB.Raw(query, data["kodebarang"]).Find(&pembelian)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(pembelian)
}

//-----SUMMARY-----

//GET SUMMARY
func SummaryPembelian(c *fiber.Ctx) error {
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		SUM(pembelian.DiskontilPembelian) AS TotalDiskontil,
		SUM(pembelian.TotalHargaBeli) AS TotalPembelian
		FROM pembelian JOIN fakturpembelian on
		fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian
		JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur
		JOIN stock ON pembelian.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
		GROUP BY faktur.NomorFaktur
	*/
	var sums []sumpembelian
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
	SUM(pembelian.DiskontilPembelian) AS TotalDiskontil,
	SUM(pembelian.TotalHargaBeli) AS TotalPembelian
	FROM pembelian JOIN fakturpembelian on
	fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur
	JOIN stock ON pembelian.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	GROUP BY faktur.NomorFaktur AND faktur.TipeTransaksi = "PEMBELIAN"`

	database.DB.Raw(query).Find(&sums)
	sum := summarypembelian{
		Details:        sums,
		TotalDiskontil: 0,
		TotalPembelian: 0,
	}
	for _, s := range sums {
		sum.TotalDiskontil += s.TotalDiskontil
		sum.TotalPembelian += s.TotalPembelian
	}
	return c.JSON(sum)
}

//POST SUMMARY PER FAKTUR
func SummaryPembelianPerFaktur(c *fiber.Ctx) error {
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		SUM(pembelian.DiskontilPembelian) AS TotalDiskontil,
		SUM(pembelian.TotalHargaBeli) AS TotalPembelian
		FROM pembelian JOIN fakturpembelian on
		fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian
		JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur
		JOIN stock ON pembelian.NomorStock = stock.NomorStock
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

	var sum sumpembelian
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
	SUM(pembelian.DiskontilPembelian) AS TotalDiskontil,
	SUM(pembelian.TotalHargaBeli) AS TotalPembelian
	FROM pembelian JOIN fakturpembelian on
	fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur
	JOIN stock ON pembelian.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE faktur.NomorFaktur = ? AND faktur.TipeTransaksi = "PEMBELIAN"`

	database.DB.Raw(query, dataint["nomorfaktur"]).Find(&sum)

	return c.JSON(sum)
}

//POST SUMMARY PER TANGGAL
func SummaryPembelianPerTanggal(c *fiber.Ctx) error {
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
	var sums []sumpembelian
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
	SUM(pembelian.DiskontilPembelian) AS TotalDiskontil,
	SUM(pembelian.TotalHargaBeli) AS TotalPembelian
	FROM pembelian JOIN fakturpembelian on
	fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur
	JOIN stock ON pembelian.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE (faktur.TanggalFaktur BETWEEN ? AND ?)
	GROUP BY faktur.NomorFaktur `

	database.DB.Raw(query, tanggalawal, tanggalakhir).Find(&sums)
	sum := summarypembelian{
		Details:        sums,
		TotalDiskontil: 0,
		TotalPembelian: 0,
	}
	for _, s := range sums {
		sum.TotalDiskontil += s.TotalDiskontil
		sum.TotalPembelian += s.TotalPembelian
	}
	return c.JSON(sum)
}

//GET SUMMARY PER BARANG
func SummaryPembelianPerBarang(c *fiber.Ctx) error {
	type sumpembelian struct {
		KodeBarang     string
		TotalDiskontil int
		TotalPembelian int
	}
	type quantitybeli struct {
		TotalQuantity int
		TipeQuantity  string
	}
	type pembelian struct {
		BarangJual     models.Barang
		TotalSmallQty  int
		TotalMediumQty int
		TotalBigQty    int
		TotalDiskontil int
		TotalPembelian int
	}
	query := `SELECT barang.KodeBarang,
	SUM(pembelian.DiskontilPembelian) AS TotalDiskontil,
	SUM(pembelian.TotalHargaBeli) AS TotalPembelian FROM pembelian
	JOIN stock ON pembelian.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	GROUP BY barang.KodeBarang`
	var sums []sumpembelian
	var pens []pembelian
	database.DB.Raw(query).Find(&sums)
	for _, s := range sums {
		var q []quantitybeli
		query = `SELECT SUM(pembelian.Quantity) AS TotalQuantity, 
		LOWER(pembelian.TipeQuantity) AS TipeQuantity FROM pembelian 
		JOIN stock ON pembelian.NomorStock = stock.NomorStock 
		JOIN barang ON barang.KodeBarang = stock.KodeBarang 
		WHERE barang.KodeBarang = ? GROUP BY LOWER(pembelian.TipeQuantity)`
		database.DB.Raw(query, s.KodeBarang).Find(&q)
		barang := GetBarang(s.KodeBarang)
		pen := pembelian{
			BarangJual:     barang,
			TotalSmallQty:  0,
			TotalMediumQty: 0,
			TotalBigQty:    0,
			TotalDiskontil: s.TotalDiskontil,
			TotalPembelian: s.TotalPembelian,
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

func TotalSummaryPembelian(c *fiber.Ctx) error {

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	template := `SUM(pembelian.DiskontilPembelian) AS TotalDiskontil,
	SUM(pembelian.TotalHargaBeli) AS Total
	FROM pembelian JOIN fakturpembelian on
	fakturpembelian.TransaksiPembelian=pembelian.TransaksiPembelian
	JOIN faktur on faktur.NomorFaktur = fakturpembelian.NomorFaktur
	JOIN stock ON pembelian.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	GROUP BY `

	return c.JSON(TotalSummary(template, data["per"]))

}
