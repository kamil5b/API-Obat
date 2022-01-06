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

type Pengajuan struct {
	NomorPengajuan int
	Before         models.Barang
	After          models.Barang
}

//GET
func GetPengajuan(c *fiber.Ctx) error {
	var updates []models.Barang
	var offerings []Pengajuan
	var arrNomor []int
	database.DB.Table("updatepengajuan").Find(&updates)
	database.DB.Table("updatepengajuan").Select("`NomorPengajuan`").Find(&arrNomor)
	i := 0
	for _, update := range updates {
		var barang models.Barang
		database.DB.Table("barang").Where("`KodeBarang` = ?", update.KodeBarang).Find(&barang)
		offering := Pengajuan{
			NomorPengajuan: arrNomor[i],
			Before:         barang,
			After:          update,
		}
		offerings = append(offerings, offering)
		i++
	}
	return c.JSON(offerings)
}

func GetTargetSales(c *fiber.Ctx) error {
	type Target struct {
		NomorTarget   int
		Karyawan      models.User
		TanggalTarget time.Time
		NominalTarget int
		Status        string
	}
	type targetquery struct {
		NomorTarget   int
		NIK           string
		TanggalTarget time.Time
		NominalTarget int
		Status        string
	}
	var targets []Target
	var qtar []targetquery
	database.DB.Table("targetsales").Find(&qtar)
	for _, tar := range qtar {
		var user models.User
		database.DB.Where("nik = ?", tar.NIK).First(&user)
		target := Target{
			NomorTarget:   tar.NominalTarget,
			Karyawan:      user,
			TanggalTarget: tar.TanggalTarget,
			NominalTarget: tar.NominalTarget,
			Status:        tar.Status,
		}
		targets = append(targets, target)
	}
	return c.JSON(targets)
}

//POST
func PengajuanUpdateBarang(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
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
	query := `INSERT INTO updatepengajuan(KodeBarang, NamaBarang, 
		TipeBigQty, BigToMedium, TipeMediumQty,
		MediumToSmall, TipeSmallQty, HargaJualKecil, 
		TipeBarang) VALUES (?,?,?,?,?,?,?,?,?)`
	database.DB.Exec(query,
		data["kode"],
		data["nama"],
		data["tipebig"],
		dataint["btm"],
		data["tipemedium"],
		dataint["mts"],
		data["tipesmall"],
		dataint["hargakecil"],
		data["tipebarang"],
	)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func PostTargetSales(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	tanggal, _ := utils.ParsingDate(data["tanggal"])
	/*
		nik :
		tanggal :
		nominal :
	*/
	query := "INSERT INTO `targetsales`( `NIK`, `TanggalTarget`, `NominalTarget`) VALUES (?,?,?)"
	database.DB.Exec(query,
		data["nik"],
		tanggal,
		dataint["nominal"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

//PUT
func UpdateBarang(c *fiber.Ctx) error {
	/*
		nomorpengajuan:
	*/
	var data map[string]string
	var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		barang := models.Barang{
			KodeBarang:     data["kode"],
			NamaBarang:     data["nama"],
			TipeBigQty:     data["tipebig"],
			BigToMedium:    btm,
			TipeMediumQty:  data["tipemedium"],
			MediumToSmall:  mts,
			TipeSmallQty:   data["tipesmall"],
			HargaJualKecil: hargakecil,
			TipeBarang:     data["tipebarang"],
		}
	*/
	database.DB.Table("updatepengajuan").Where("`NomorPengajuan` = ?", dataint["nomorpengajuan"]).Find(&barang)
	err := DeletePengajuan(dataint["nomorpengajuan"])
	if err != nil {
		return c.JSON(fiber.Map{
			"message": err,
		})
	}
	query := `UPDATE barang SET NamaBarang= ? ,TipeBigQty= ? ,
	BigToMedium= ? ,TipeMediumQty= ? ,MediumToSmall= ? ,TipeSmallQty= ? ,
	HargaJualKecil= ? ,TipeBarang= ? WHERE KodeBarang = ? `
	db := database.DB.Exec(query,
		barang.NamaBarang,
		barang.TipeBigQty,
		barang.BigToMedium,
		barang.TipeMediumQty,
		barang.MediumToSmall,
		barang.TipeSmallQty,
		barang.HargaJualKecil,
		barang.TipeBarang,
		barang.KodeBarang,
	)
	if db.Error != nil {
		return c.JSON(fiber.Map{
			"message": "gagal update barang",
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func GantiStatus(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		nomor :
		status :
	*/
	query := "UPDATE `targetsales` SET `Status`= ? WHERE `NomorTarget` = ?"
	database.DB.Exec(query,
		data["status"],
		dataint["nomor"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

//DELETE
func TolakPengajuan(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	err := DeletePengajuan(dataint["nomorpengajuan"])
	if err != nil {
		return c.JSON(fiber.Map{
			"message": err,
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func DeletePengajuan(nomorpengajuan int) error {
	query := `DELETE FROM updatepengajuan WHERE NomorPengajuan = ?`
	db := database.DB.Exec(query,
		nomorpengajuan,
	)

	if db.Error != nil {
		return errors.New("gagal menghapus pengajuan")
	}
	return nil
}
