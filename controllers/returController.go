package controllers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/API-Obat/database"
	"github.com/kamil5b/API-Obat/models"
	"github.com/kamil5b/API-Obat/utils"
)

type returquery struct {
	NomorRetur     int
	NomorFaktur    int
	Status         string
	KodeBarang     string
	Quantity       int
	TipeQuantity   string
	DiskontilRetur int
	TotalNominal   int
	Description    string
}

//GET
func GetAllRetur(c *fiber.Ctx) error {
	var rets []returquery
	var returs []models.Retur

	database.DB.Table("retur").Find(&rets)
	for _, ret := range rets {
		var faktur models.Faktur
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", ret.NomorFaktur).Find(&faktur)
		barang := GetBarang(ret.KodeBarang)

		retur := models.Retur{
			FakturBarang:   faktur,
			Status:         ret.Status,
			BarangRetur:    barang,
			Quantity:       ret.Quantity,
			TipeQuantity:   ret.TipeQuantity,
			DiskontilRetur: ret.DiskontilRetur,
			TotalNominal:   ret.TotalNominal,
			Description:    ret.Description,
		}
		returs = append(returs, retur)
	}
	if len(returs) == 0 {
		return c.JSON(fiber.Map{
			"message": "laporan keuangan tidak ditemukan",
		})

	}

	return c.JSON(returs)
}

//POST
func ReturBarang(c *fiber.Ctx) error {
	/*
		{
			nomorfaktur:
			quantity:
			tipequantity:
			kodebarang:
			diskontil:
			desc:
		}
	*/
	var data map[string]string
	var faktur models.Faktur
	//var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}

	dataint := utils.MapStringToInt(data)
	db := database.DB.Table("faktur").Where("`NomorFaktur` = ?", data["nomorfaktur"]).Scan(&faktur)
	if db.Error != nil {
		fmt.Println(db.Error)
		return c.JSON(fiber.Map{
			"message": "nomor faktur belum terdaftar",
		})
	}
	barang := GetBarang(data["kodebarang"])
	total := 0
	faktor := 0
	if faktur.TipeTransaksi == "PENJUALAN" {
		faktor = barang.HargaJualKecil
	} else if faktur.TipeTransaksi == "PEMBELIAN" {
		stock := GetStockByFakturBarang(faktur.NomorFaktur, data["kodebarang"])
		faktor = stock.HargaBeliKecil
	}
	if barang.ConvertQty(data["tipequantity"], "small", dataint["quantity"]) != 0 {
		total = (barang.ConvertQty(
			data["tipequantity"],
			"small",
			dataint["quantity"],
		) * faktor) - dataint["diskontil"]
	}
	retur := models.Retur{
		FakturBarang:   faktur,
		Status:         "RETUR",
		BarangRetur:    barang,
		Quantity:       dataint["quantity"],
		TipeQuantity:   data["tipequantity"],
		DiskontilRetur: dataint["diskontil"],
		TotalNominal:   total,
		Description:    data["desc"],
	}
	query := `INSERT INTO retur(NomorFaktur, TipeTransaksi, Status, KodeBarang, Quantity, TipeQuantity, 
		DiskontilRetur, TotalNominal, Description) VALUES (?,?,?,?,?,?,?,?,?)`
	db = database.DB.Exec(query,
		retur.FakturBarang.NomorFaktur,
		faktur.TipeTransaksi,
		retur.Status,
		barang.KodeBarang,
		dataint["quantity"],
		retur.TipeQuantity,
		retur.DiskontilRetur,
		retur.TotalNominal,
		retur.Description,
	)
	if db.Error != nil {
		fmt.Println(db.Error)
		return c.JSON(fiber.Map{
			"message": "error insert retur",
		})
	}
	tipeqty := ""
	if strings.EqualFold(data["tipequantity"], retur.BarangRetur.TipeBigQty) {
		//qty = barang.BigQty
		tipeqty = "BigQty"
	} else if strings.EqualFold(data["tipequantity"], retur.BarangRetur.TipeMediumQty) {
		//qty = barang.MediumQty
		tipeqty = "MediumQty"
	} else if strings.EqualFold(data["tipequantity"], retur.BarangRetur.TipeSmallQty) {
		//qty = barang.SmallQty
		tipeqty = "SmallQty"
	} else {
		return c.JSON(fiber.Map{
			"message": "tipe quantity tidak ditemukan pada barang",
		})
	}

	/*
		SELECT stock.Expired, stock.HargaBeliKecil FROM stock
		JOIN pembelian ON pembelian.NomorStock = stock.NomorStock
		JOIN fakturpembelian ON pembelian.TransaksiPembelian = fakturpembelian.TransaksiPembelian

		SELECT stock.Expired, stock.HargaBeliKecil FROM stock
		JOIN penjualan ON penjualan.NomorStock = stock.NomorStock
		JOIN fakturpenjualan ON penjualan.TransaksiPenjualan = fakturpenjualan.TransaksiPenjualan
		JOIN pembelian ON pembelian.NomorStock = stock.NomorStock
		JOIN fakturpembelian ON pembelian.TransaksiPembelian = fakturpembelian.TransaksiPembelian
	*/
	//ambil stock
	//var retur returquery
	//PEMBELIAN -2 PENJUALAN -1
	query = `SELECT stock.Expired, stock.HargaBeliKecil FROM stock
	JOIN pembelian ON pembelian.NomorStock = stock.NomorStock
	JOIN fakturpembelian ON 
	pembelian.TransaksiPembelian = fakturpembelian.TransaksiPembelian `
	if faktur.TipeTransaksi == "PEMBELIAN" {
		query += ` WHERE fakturpembelian.NomorFaktur = ? AND stock.KodeBarang = ?`
	} else if faktur.TipeTransaksi == "PENJUALAN" {
		query += ` JOIN penjualan ON penjualan.NomorStock = stock.NomorStock
			JOIN fakturpenjualan ON penjualan.TransaksiPenjualan = fakturpenjualan.TransaksiPenjualan
			WHERE fakturpenjualan.NomorFaktur = ? AND stock.KodeBarang = ?`
	}

	type s struct {
		Expired        time.Time
		HargaBeliKecil int
	}
	var expharga s
	database.DB.Raw(query, faktur.NomorFaktur, barang.KodeBarang).Find(&expharga)

	var nostock int
	query = "SELECT `NomorStock` FROM stock ORDER BY `NomorStock` DESC LIMIT 1"
	database.DB.Raw(query).Find(&nostock)
	nostock++
	query = `INSERT INTO stock(NomorStock, KodeBarang, Expired, 
		` + tipeqty + `, HargaBeliKecil) 
		VALUES (?,?,?,?,?)`
	database.DB.Exec(query,
		nostock,
		barang.KodeBarang,
		expharga.Expired,
		dataint["quantity"],
		expharga.HargaBeliKecil,
	)
	query = `INSERT INTO pembelian(Quantity,
		TipeQuantity, DiskontilPembelian,
		TotalHargaBeliNomorStock) VALUES (?,?,?,?,?)`
	db = database.DB.Exec(query,
		dataint["quantity"],
		data["tipequantity"],
		dataint["diskontil"],
		total,
		nostock,
	)
	if db.Error != nil {
		fmt.Println(db.Error)
		return c.JSON(fiber.Map{
			"message": "error insert pembelian retur",
		})
	}
	//query = "UPDATE `stock` SET " + tipeqty + " = ? WHERE `stock`.`KodeBarang` = ?"
	//retur.Quantity = qty
	//database.DB.Exec(query, qty, barang.KodeBarang)
	/*
		INSERT INTO retur(Status, KodeBarang, Quantity, TipeQuantity,
			DiskontilRetur, TotalNominal, Description) VALUES (?,?,?,?,?,?,?)
	*/
	var notransaksi int
	var transaksi string
	if faktur.TipeTransaksi == "PENJUALAN" {
		transaksi = "TransaksiPenjualan"
		query = "SELECT `TransaksiPenjualan` FROM penjualan ORDER BY `TransaksiPenjualan` DESC LIMIT 1"
		database.DB.Raw(query).Scan(&notransaksi)
		query = "INSERT INTO `fakturpenjualan`(`NomorFaktur`, `TransaksiPenjualan`) VALUES (-1,?)"
		db = database.DB.Exec(
			query,
			notransaksi,
		)
		if db.Error != nil {
			fmt.Println(db.Error)
			return c.JSON(fiber.Map{
				"message": "error insert pembelian retur",
			})
		}

		query = `UPDATE hutang SET NominalRetur = NominalRetur + ?, 
		Diskontil = Diskontil - ?, SisaHutang = SisaHutang - ? 
		WHERE NomorFaktur = ?`
	} else if faktur.TipeTransaksi == "PEMBELIAN" {
		transaksi = "TransaksiPembelian"
		query = "SELECT `TransaksiPembelian` FROM pembelian ORDER BY `TransaksiPembelian` DESC LIMIT 1"
		database.DB.Raw(query).Scan(&notransaksi)
		query = "INSERT INTO `fakturpembelian`(`NomorFaktur`, `TransaksiPembelian`) VALUES (-2,?)"
		db = database.DB.Exec(
			query,
			notransaksi,
		)
		if db.Error != nil {
			fmt.Println(db.Error)
			return c.JSON(fiber.Map{
				"message": "error insert pembelian retur",
			})
		}
		query = `UPDATE piutang SET NominalRetur = NominalRetur + ?, 
		Diskontil = Diskontil - ?, SisaPiutang = SisaPiutang - ? 
		WHERE NomorFaktur = ?`
	}
	database.DB.Exec(query,
		retur.TotalNominal,
		retur.DiskontilRetur,
		retur.TotalNominal,
		faktur.NomorFaktur,
	)
	query = `UPDATE ` + strings.ToLower(faktur.TipeTransaksi) + ` SET ` + tipeqty + ` = ` + tipeqty + ` - ? 
			WHERE ` + transaksi + ` = ? `
	database.DB.Exec(query,
		dataint["quantity"],
		notransaksi,
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func GetReturFaktur(nomorfaktur int) []models.Retur {

	var returs []models.Retur
	var r []returquery
	var faktur models.Faktur
	database.DB.Table("faktur").Where("`NomorFaktur` = ?", nomorfaktur).Find(&faktur)
	database.DB.Table("retur").Where("'NomorFaktur' = ?", nomorfaktur).Find(&r)
	for _, ret := range r {
		barang := GetBarang(ret.KodeBarang)
		retur := models.Retur{
			FakturBarang:   faktur,
			Status:         ret.Status,
			BarangRetur:    barang,
			Quantity:       ret.Quantity,
			TipeQuantity:   ret.TipeQuantity,
			DiskontilRetur: ret.DiskontilRetur,
			TotalNominal:   ret.TotalNominal,
			Description:    ret.Description,
		}
		returs = append(returs, retur)
	}
	return returs
}

/*
func GetReturStock(KodeBarang int) (models.Retur, error) {
	var retur models.Retur
	var faktur models.Faktur
	var r returquery
	r.NomorRetur = 0
	db := database.DB.Table("retur").Where("'KodeBarang' = ?", KodeBarang).Find(&r)
	if db.Error != nil {
		return models.Retur{}, db.Error
	}
	if r.NomorRetur == 0 {
		return models.Retur{}, errors.New("no retur")
	}
	database.DB.Table("faktur").Where("'NomorFaktur' = ?", r.NomorFaktur).Find(&faktur)
	stock := GetStock(r.KodeBarang)
	retur = models.Retur{
		FakturBarang:   faktur,
		Status:         r.Status,
		BarangRetur:    barang,
		Quantity:       r.Quantity,
		TipeQuantity:   r.TipeQuantity,
		DiskontilRetur: r.DiskontilRetur,
		TotalNominal:   r.TotalNominal,
		Description:    r.Description,
	}
	return retur, nil
}
*/
