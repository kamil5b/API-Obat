package controllers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/API-Obat/database"
	"github.com/kamil5b/API-Obat/models"
	"github.com/kamil5b/API-Obat/utils"
)

func GetBank(c *fiber.Ctx) error {
	var bank []models.Bank
	database.DB.Table("bank").Find(&bank)
	return c.JSON(bank)
}

func GetCustomer(c *fiber.Ctx) error {
	var customer []models.Customer
	database.DB.Table("customer").Find(&customer)
	return c.JSON(customer)
}

func PostCustomer(c *fiber.Ctx) error {
	var data map[string]string
	//var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		{
			nama:
			nomor:
			alamat:
		}
	*/
	query := "INSERT INTO `customer`(`NamaCustomer`, `Alamat`, `NomorHP`) VALUES (?,?,?)"
	database.DB.Exec(query, data["nama"], data["alamat"], data["nomor"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func UpdateCustomer(c *fiber.Ctx) error {
	var data map[string]string
	//var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		{
			nomorcustomer:
			nama:
			nomor:
			alamat:
		}
	*/
	query := "UPDATE customer SET NamaCustomer = ?, NomorHP = ?, Alamat = ? WHERE NomorUrut = ?"
	database.DB.Exec(query, data["nama"], data["nomor"], data["alamat"], dataint["nomorcustomer"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func GetToko(c *fiber.Ctx) error {
	var toko []models.Toko
	database.DB.Table("toko").Find(&toko)
	return c.JSON(toko)
}

func PostToko(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		{
			nama:
			nomor:
			alamat:
		}
	*/
	query := "INSERT INTO toko(NamaToko, NomorTelepon, Alamat) VALUES (?,?,?)"
	database.DB.Exec(query, data["nama"], data["nomor"], data["alamat"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func UpdateToko(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		{
			nomortoko:
			nama:
			nomor:
			alamat:
		}
	*/
	query := "UPDATE toko SET NamaToko = ?, NomorTelepon = ?, Alamat = ? WHERE NomorToko = ?"
	database.DB.Exec(query, data["nama"], data["nomor"], data["alamat"], dataint["nomortoko"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func GetGiro(c *fiber.Ctx) error {
	var giros []models.Giro
	var nobank int
	rows, err := database.DB.Table("giro").Rows()
	if err != nil {
		return err
	}
	for rows.Next() {
		var giro models.Giro
		var bank models.Bank
		rows.Scan(
			&giro.NomorGiro,
			&giro.Nominal,
			&giro.TanggalGiro,
			nobank,
		)
		database.DB.Table("bank").Find(&bank)
		giro.BankGiro = bank
		giros = append(giros, giro)
	}

	return c.JSON(giros)
}

func PostGiro(c *fiber.Ctx) error {
	var data map[string]string
	//var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		{
			nomorgiro:
			nominal:
			tanggal:
			nomorbank:
		}
	*/
	query := "INSERT INTO giro(NomorGiro, Nominal, TanggalGiro, NomorBank) VALUES (?,?,?,?)"
	tanggal, err := utils.ParsingDate(data["tanggal"])
	if err != nil {
		return err
	}
	database.DB.Exec(query, data["nomorgiro"], data["nominal"], tanggal, dataint["nomorbank"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func TotalSummary(template, per string) interface{} {
	type perhari struct {
		Tahun          int
		Bulan          int
		Tanggal        int
		TotalDiskontil int
		Total          int
	}
	type perminggu struct {
		Tahun          int
		Bulan          int
		Minggu         int
		TotalDiskontil int
		Total          int
	}
	type perbulan struct {
		Tahun          int
		Bulan          int
		TotalDiskontil int
		Total          int
	}
	type pertahun struct {
		Tahun          int
		TotalDiskontil int
		Total          int
	}

	query := ""
	/*
		{
			per:, //hari,minggu,bulan,tahun  SELECT
		}

		query = `
		SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
		SUM(penjualan.TotalHarga) AS TotalPenjualan FROM penjualan
		JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
		GROUP BY
		`

		//PER HARI
		SELECT YEAR(faktur.TanggalFaktur) AS Tahun,
		MONTH(faktur.TanggalFaktur) AS Bulan,
		DAY(faktur.TanggalFaktur) AS Tanggal,
		query
		YEAR(faktur.TanggalFaktur), MONTH(faktur.TanggalFaktur),DAY(faktur.TanggalFaktur)

		//PER MINGGU
		SELECT YEAR(faktur.TanggalFaktur) AS Tahun,
		MONTH(faktur.TanggalFaktur) AS Bulan,
		WEEK(faktur.TanggalFaktur) AS Minggu,
		query
		YEAR(faktur.TanggalFaktur), MONTH(faktur.TanggalFaktur), WEEK(faktur.TanggalFaktur)

		//PER BULAN
		SELECT YEAR(faktur.TanggalFaktur) AS Tahun,
		MONTH(faktur.TanggalFaktur) AS Bulan,
		query
		YEAR(faktur.TanggalFaktur), MONTH(faktur.TanggalFaktur)

		//PER TAHUN
		SELECT YEAR(faktur.TanggalFaktur) AS Tahun,
		query
		BY YEAR(faktur.TanggalFaktur)
	*/
	if per == "hari" {
		query = `SELECT YEAR(faktur.TanggalFaktur) AS Tahun,
		MONTH(faktur.TanggalFaktur) AS Bulan,
		DAY(faktur.TanggalFaktur) AS Tanggal,` + template + `
		YEAR(faktur.TanggalFaktur), 
		MONTH(faktur.TanggalFaktur),
		DAY(faktur.TanggalFaktur)
		`
		var summary []perhari
		database.DB.Raw(query).Find(&summary)
		return (summary)
	}
	if per == "minggu" {
		query = `SELECT YEAR(faktur.TanggalFaktur) AS Tahun,
		MONTH(faktur.TanggalFaktur) AS Bulan,
		WEEK(faktur.TanggalFaktur) AS Minggu,` + template + `
		YEAR(faktur.TanggalFaktur), 
		MONTH(faktur.TanggalFaktur),
		WEEK(faktur.TanggalFaktur)
		`
		var summary []perminggu
		database.DB.Raw(query).Find(&summary)
		return summary
	}
	if per == "bulan" {
		query = `SELECT YEAR(faktur.TanggalFaktur) AS Tahun,
		MONTH(faktur.TanggalFaktur) AS Bulan,` + template + `
		YEAR(faktur.TanggalFaktur), 
		MONTH(faktur.TanggalFaktur)
		`
		var summary []perbulan
		database.DB.Raw(query).Find(&summary)
		return summary
	}
	if per == "tahun" {
		query = `SELECT YEAR(faktur.TanggalFaktur) AS Tahun,` + template + `
		YEAR(faktur.TanggalFaktur)
		`
		var summary []pertahun
		database.DB.Raw(query).Find(&summary)
		return summary
	}
	return fiber.Map{
		"message": "per invalid",
	}

}

func GetProfits(c *fiber.Ctx) error {
	type queryprofit struct {
		NomorProfit    int
		NomorFaktur    int
		TanggalFaktur  time.Time
		KodeBarang     string
		TotalPembelian int
		TotalPenjualan int
		TotalProfit    int
	}
	type out struct {
		NomorProfit    int
		Faktur         models.Faktur
		Barang         models.Barang
		TotalPembelian int
		TotalPenjualan int
		TotalProfit    int
	}
	type totalprofit struct {
		Profit         []out
		TotalPembelian int
		TotalPenjualan int
		TotalProfit    int
	}
	var profits []queryprofit
	var total totalprofit
	total.TotalPembelian = 0
	total.TotalPenjualan = 0
	total.TotalProfit = 0
	database.DB.Table("profits").Find(&profits)
	for _, profit := range profits {
		var faktur models.Faktur
		var barang models.Barang
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", profit.NomorFaktur).Find(&faktur)
		database.DB.Table("barang").Where("`KodeBarang` = ?", profit.KodeBarang).Find(&barang)
		total.TotalPembelian++
		total.TotalPenjualan++
		total.TotalProfit++
		outs := out{
			NomorProfit:    profit.NomorProfit,
			Faktur:         faktur,
			Barang:         barang,
			TotalPembelian: profit.TotalPembelian,
			TotalPenjualan: profit.TotalPenjualan,
			TotalProfit:    profit.TotalProfit,
		}
		total.Profit = append(total.Profit, outs)
	}
	return c.JSON(total)
}

func InsertFaktur(data map[string]string, nomorentitas int, tipetransaksi string) interface{} {
	dataint := utils.MapStringToInt(data)
	tanggal, err := utils.ParsingDate(data["tanggal"])
	if err != nil {
		return fiber.Map{
			"message": "error parsing date",
		}
	}
	jatuhtempo, err := utils.ParsingDate(data["jatuhtempo"])
	if err != nil {
		return fiber.Map{
			"message": "error parsing date",
		}
	}
	/*
		type Faktur struct {
			NomorFaktur    int
			TanggalFaktur  time.Time
			TipeTransaksi  string
			TipePembayaran string
			JatuhTempo     time.Time
			NomorGiro      string
		}
	*/
	fmt.Println(dataint["nomor"],
		tanggal,
		tipetransaksi,
		nomorentitas,
		data["tipepembayaran"],
		data["nomorgiro"],
		jatuhtempo)
	query := `INSERT INTO faktur(NomorFaktur, TanggalFaktur, 
		TipeTransaksi,NomorEntitas,TipePembayaran,
		NomorGiro,JatuhTempo) VALUES (?,?,?,?,?,?,?)`
	database.DB.Exec(query,
		dataint["nomor"],
		tanggal,
		tipetransaksi,
		nomorentitas,
		data["tipepembayaran"],
		data["nomorgiro"],
		jatuhtempo,
	)
	return fiber.Map{
		"message": "success",
	}
}
