package controllers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/API-Obat/database"
	"github.com/kamil5b/API-Obat/models"
	"github.com/kamil5b/API-Obat/utils"
)

type gethutang struct {
	NomorHutang  int
	NomorFaktur  int
	NomorToko    int
	NominalRetur int
	Diskontil    int
	SisaHutang   int
	JatuhTempo   time.Time
}

//GET
func GetHutang(c *fiber.Ctx) error {

	/*

		type Hutang struct {
			FakturHutang Faktur
			TokoDihutang Toko
			StockBarang  Stock
			ReturBarang  Retur
			SisaHutang   int
			JatuhTempo   time.Time
		}

	*/
	var hutangs []models.Hutang
	var htng []gethutang
	database.DB.Table("hutang").Where("`SisaHutang` > 0").Group("`NomorToko`").Find(&htng)
	for _, tmp := range htng {
		var faktur models.Faktur
		var toko models.Toko
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", tmp.NomorFaktur).Find(&faktur)
		database.DB.Table("toko").Where("`NomorToko` = ?", tmp.NomorToko).Find(&toko)
		hutang := models.Hutang{
			NomorHutang:  tmp.NomorHutang,
			FakturHutang: faktur,
			TokoDihutang: toko,
			NominalRetur: tmp.NominalRetur,
			SisaHutang:   tmp.SisaHutang,
			JatuhTempo:   tmp.JatuhTempo,
		}
		hutangs = append(hutangs, hutang)
	}
	return c.JSON(hutangs)
}
func GetRecordHutang(c *fiber.Ctx) error {
	type record struct {
		NomorUrut    int
		NomorFaktur  int
		NomorToko    int
		NomorHutang  int
		Nominal      int
		TanggalBayar time.Time
	}
	type RecordHutang struct {
		NomorUrut    int
		NomorHutang  int
		Faktur       models.Faktur
		Toko         models.Toko
		Nominal      int
		TanggalBayar time.Time
	}
	var records []record
	var recordhutangs []RecordHutang
	query := "SELECT * FROM recordhutang join hutang on hutang.NomorHutang = recordhutang.NomorHutang"
	database.DB.Raw(query).Find(&records)
	for _, tmp := range records {
		var faktur models.Faktur
		var toko models.Toko
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", tmp.NomorFaktur).Find(&faktur)
		database.DB.Table("toko").Where("`NomorToko` = ?", tmp.NomorToko).Find(&toko)
		recordhutang := RecordHutang{
			NomorUrut:    tmp.NomorHutang,
			NomorHutang:  tmp.NomorHutang,
			Faktur:       faktur,
			Toko:         toko,
			Nominal:      tmp.Nominal,
			TanggalBayar: tmp.TanggalBayar,
		}
		recordhutangs = append(recordhutangs, recordhutang)
	}
	return c.JSON(recordhutangs)
}

//POST
func GetHutangPerToko(c *fiber.Ctx) error {
	type hutangtoko struct {
		Hutang      []models.Hutang
		TotalHutang int
		TotalSisa   int
	}
	var hutangs hutangtoko
	var htng []gethutang
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		NOMOR TOKO
	*/
	database.DB.Table("hutang").Where("`SisaHutang` > 0 AND `NomorToko` = ?", data["nomortoko"]).Find(&htng)
	for _, tmp := range htng {
		var faktur models.Faktur
		var toko models.Toko
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", tmp.NomorFaktur).Find(&faktur)
		database.DB.Table("toko").Where("`NomorToko` = ?", tmp.NomorToko).Find(&toko)
		hutang := models.Hutang{
			NomorHutang:  tmp.NomorHutang,
			FakturHutang: faktur,
			TokoDihutang: toko,
			NominalRetur: tmp.NominalRetur,
			SisaHutang:   tmp.SisaHutang,
			JatuhTempo:   tmp.JatuhTempo,
		}
		hutangs.Hutang = append(hutangs.Hutang, hutang)
	}
	return c.JSON(hutangs)
}

func HutangBarang(nomorfaktur, diskontil, nominal int) {
	/*
		var toko models.Toko
		var faktur models.Faktur
		query := "SELECT `TransaksiPembelian` FROM pembelian WHERE `TransaksiPembelian` = ?"
		database.DB.Raw(query, transaksipenjualan).Scan(&notransaksi)
		retur := GetReturFaktur(nomorfaktur)
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", nomorfaktur).Find(&faktur)
		database.DB.Table("toko").Where("`NomorToko` = ?", nomortoko).Find(&toko)
		/*
			hutang := models.Hutang{
				FakturHutang: faktur,
				TokoDihutang: toko,
				ReturBarang:  retur,
				SisaHutang:   nominal,
				JatuhTempo:   jatuhtempo,
			}
			/*
			INSERT INTO hutang(NomorFaktur, NomorToko, NominalRetur,
				Diskontil, SisaHutang, JatuhTempo) VALUES (?,?,?,?,?,?)

		database.DB.Table("fakturpembelian").Select("`DiskontilPembelian`").Joins("join pembelian on ").Where("`NomorToko` = ?", nomortoko).Find(&toko)
	*/
	var faktur models.Faktur
	query := `SELECT * FROM faktur WHERE NomorFaktur=?`
	database.DB.Raw(query,
		nomorfaktur,
	).Find(&faktur)
	var hutang gethutang
	query = `SELECT * FROM hutang WHERE NomorFaktur=? AND NomorToko=?`
	database.DB.Raw(query,
		nomorfaktur,
		faktur.NomorEntitas,
	).Find(&hutang)
	if hutang.NomorHutang == 0 {
		query = `INSERT INTO hutang(NomorFaktur, NomorToko,
			NominalRetur, Diskontil, SisaHutang, JatuhTempo) VALUES (?,?,?,?,?,?)`
		database.DB.Exec(query,
			nomorfaktur,
			faktur.NomorEntitas,
			0,
			diskontil,
			nominal,
			faktur.JatuhTempo,
		)
	} else {
		query = `UPDATE hutang SET 
			Diskontil=Diskontil+?,SisaHutang=SisaHutang+? 
			WHERE NomorFaktur=? AND NomorToko=?`
		database.DB.Exec(query,
			diskontil,
			nominal,
			nomorfaktur,
			faktur.NomorEntitas,
		)
	}

}

//PUT
func BayarHutang(c *fiber.Ctx) error {
	/*
		{
			nomorhutang:
			bayar:
			tanggal:
		}
	*/
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	tanggal, _ := utils.ParsingDate(data["tanggal"])
	query := "INSERT INTO `recordhutang`(`NomorHutang`, `Nominal`, `TanggalBayar`) VALUES (?,?,?)"
	database.DB.Exec(query,
		dataint["nomorhutang"],
		-dataint["bayar"],
		tanggal,
	)
	query = "UPDATE hutang SET `SisaHutang` = `SisaHutang` - ? WHERE `NomorHutang` = ?"
	database.DB.Exec(query,
		dataint["bayar"],
		dataint["nomorhutang"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func HutangNaik(c *fiber.Ctx) error {
	/*
		{
			nomorhutang:
			naik:
			tanggal:
		}
	*/
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	tanggal, _ := utils.ParsingDate(data["tanggal"])
	query := "INSERT INTO `recordhutang`(`NomorHutang`, `Nominal`, `TanggalBayar`) VALUES (?,?,?)"
	database.DB.Exec(query,
		dataint["nomorhutang"],
		dataint["bayar"],
		tanggal,
	)
	query = "UPDATE hutang SET `SisaHutang` = `SisaHutang` + ? WHERE `NomorHutang` = ?"
	database.DB.Exec(query,
		dataint["naik"],
		dataint["nomorhutang"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
