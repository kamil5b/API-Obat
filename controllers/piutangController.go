package controllers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/API-Obat/database"
	"github.com/kamil5b/API-Obat/models"
	"github.com/kamil5b/API-Obat/utils"
)

type getpiutang struct {
	NomorPiutang  int
	NomorFaktur   int
	NomorCustomer int
	NominalRetur  int
	Diskontil     int
	SisaPiutang   int
	JatuhTempo    time.Time
}

//GET
func GetPiutang(c *fiber.Ctx) error {

	/*

		type Piutang struct {
			FakturPiutang Faktur
			CustomerDipiutang Customer
			StockBarang  Stock
			ReturBarang  Retur
			SisaPiutang   int
			JatuhTempo   time.Time
		}

	*/
	var piutangs []models.Piutang
	var htng []getpiutang
	database.DB.Table("piutang").Where("`SisaPiutang` > 0").Group("`NomorCustomer`").Find(&htng)
	for _, tmp := range htng {
		var faktur models.Faktur
		var customer models.Customer
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", tmp.NomorFaktur).Find(&faktur)
		database.DB.Table("customer").Where("`NomorUrut` = ?", tmp.NomorCustomer).Find(&customer)
		piutang := models.Piutang{
			NomorPiutang:    tmp.NomorPiutang,
			FakturPiutang:   faktur,
			CustomerPiutang: customer,
			NominalRetur:    tmp.NominalRetur,
			SisaPiutang:     tmp.SisaPiutang,
			JatuhTempo:      tmp.JatuhTempo,
		}
		piutangs = append(piutangs, piutang)
	}
	return c.JSON(piutangs)
}

//POST
func GetPiutangPerCustomer(c *fiber.Ctx) error {
	type piutangcustomer struct {
		Piutang      []models.Piutang
		TotalPiutang int
		TotalSisa    int
	}
	var piutangs piutangcustomer
	var ptng []getpiutang
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		NOMOR CUSTOMER
	*/
	database.DB.Table("piutang").Where("`SisaPiutang` > 0 AND `NomorCustomer` = ?", data["nomorcustomer"]).Find(&ptng)
	for _, tmp := range ptng {
		var faktur models.Faktur
		var customer models.Customer
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", tmp.NomorFaktur).Find(&faktur)
		database.DB.Table("customer").Where("`NomorCustomer` = ?", tmp.NomorCustomer).Find(&customer)
		piutang := models.Piutang{
			NomorPiutang:    tmp.NomorPiutang,
			FakturPiutang:   faktur,
			CustomerPiutang: customer,
			NominalRetur:    tmp.NominalRetur,
			SisaPiutang:     tmp.SisaPiutang,
			JatuhTempo:      tmp.JatuhTempo,
		}
		piutangs.Piutang = append(piutangs.Piutang, piutang)
	}
	return c.JSON(piutangs)
}

func PiutangBarang(nomorfaktur, diskontil, nominal int) {
	/*
		var customer models.Customer
		var faktur models.Faktur
		query := "SELECT `TransaksiPembelian` FROM pembelian WHERE `TransaksiPembelian` = ?"
		database.DB.Raw(query, transaksipenjualan).Scan(&notransaksi)
		retur := GetReturFaktur(nomorfaktur)
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", nomorfaktur).Find(&faktur)
		database.DB.Table("customer").Where("`NomorCustomer` = ?", nomorcustomer).Find(&customer)
		/*
			piutang := models.Piutang{
				FakturPiutang: faktur,
				CustomerDipiutang: customer,
				ReturBarang:  retur,
				SisaPiutang:   nominal,
				JatuhTempo:   jatuhtempo,
			}
			/*
			INSERT INTO piutang(NomorFaktur, NomorCustomer, NominalRetur,
				Diskontil, SisaPiutang, JatuhTempo) VALUES (?,?,?,?,?,?)

		database.DB.Table("fakturpembelian").Select("`DiskontilPembelian`").Joins("join pembelian on ").Where("`NomorCustomer` = ?", nomorcustomer).Find(&customer)
	*/
	var faktur models.Faktur
	query := `SELECT * FROM faktur WHERE NomorFaktur=?`
	database.DB.Raw(query,
		nomorfaktur,
	).Find(&faktur)
	var piutang getpiutang
	query = `SELECT * FROM piutang WHERE NomorFaktur=? AND NomorCustomer=?`
	database.DB.Raw(query,
		nomorfaktur,
		faktur.NomorEntitas,
	).Find(&piutang)
	if piutang.NomorPiutang == 0 {
		query = `INSERT INTO piutang(NomorFaktur, NomorCustomer,
			NominalRetur, Diskontil, SisaPiutang, JatuhTempo) VALUES (?,?,?,?,?,?)`
		database.DB.Exec(query,
			nomorfaktur,
			faktur.NomorEntitas,
			0,
			diskontil,
			nominal,
			faktur.JatuhTempo,
		)
	} else {
		query = `UPDATE piutang SET 
			Diskontil=Diskontil+?,SisaPiutang=SisaPiutang+? 
			WHERE NomorFaktur=? AND NomorCustomer=?`
		database.DB.Exec(query,
			diskontil,
			nominal,
			nomorfaktur,
			faktur.NomorEntitas,
		)
	}
}
func GetRecordPiutang(c *fiber.Ctx) error {
	type record struct {
		NomorUrut     int
		NomorFaktur   int
		NomorCustomer int
		NomorPiutang  int
		Nominal       int
		TanggalBayar  time.Time
	}
	type RecordPiutang struct {
		NomorUrut    int
		NomorPiutang int
		Faktur       models.Faktur
		Customer     models.Customer
		Nominal      int
		TanggalBayar time.Time
	}
	var records []record
	var recordpiutangs []RecordPiutang
	query := "SELECT * FROM recordpiutang join piutang on piutang.NomorPiutang = recordpiutang.NomorPiutang"
	database.DB.Raw(query).Find(&records)
	for _, tmp := range records {
		var faktur models.Faktur
		var customer models.Customer
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", tmp.NomorFaktur).Find(&faktur)
		database.DB.Table("customer").Where("`NomorUrut` = ?", tmp.NomorCustomer).Find(&customer)
		recordpiutang := RecordPiutang{
			NomorUrut:    tmp.NomorPiutang,
			NomorPiutang: tmp.NomorPiutang,
			Faktur:       faktur,
			Customer:     customer,
			Nominal:      tmp.Nominal,
			TanggalBayar: tmp.TanggalBayar,
		}
		recordpiutangs = append(recordpiutangs, recordpiutang)
	}
	return c.JSON(recordpiutangs)
}

//PUT
func BayarPiutang(c *fiber.Ctx) error {
	/*
		{
			nomorpiutang:
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
	query := "INSERT INTO `recordpiutang`(`NomorPiutang`, `Nominal`, `TanggalBayar`) VALUES (?,?,?)"
	database.DB.Exec(query,
		dataint["nomorpiutang"],
		-dataint["bayar"],
		tanggal,
	)
	query = "UPDATE piutang SET `SisaPiutang` = `SisaPiutang` - ? WHERE `NomorPiutang` = ?"
	database.DB.Exec(query,
		dataint["bayar"],
		dataint["nomorpiutang"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func PiutangNaik(c *fiber.Ctx) error {
	/*
		{
			nomorpiutang:
			naik:
		}
	*/
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	query := "INSERT INTO `recordpiutang`(`NomorPiutang`, `Nominal`) VALUES (?,?)"
	database.DB.Exec(query,
		dataint["nomorpiutang"],
		dataint["bayar"],
	)
	query = "UPDATE piutang SET `SisaPiutang` = `SisaPiutang` + ? WHERE `NomorPiutang` = ?"
	database.DB.Exec(query,
		dataint["naik"],
		dataint["nomorpiutang"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
