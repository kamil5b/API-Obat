package controllers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/API-Obat/database"
	"github.com/kamil5b/API-Obat/models"
	"github.com/kamil5b/API-Obat/utils"
)

type querydemand struct {
	ID                int
	KodeBarang        string
	QuantityDemand    int
	QuantityAvailable int
	TipeQuantity      string
}

//GET DEMAND
func GetAllDemand(c *fiber.Ctx) error {

	var qds []querydemand
	var demands []models.Demand

	database.DB.Table("demands").Find(&qds)
	for _, qd := range qds {
		var barang models.Barang
		var tipeqty string
		var quantityavailable int
		database.DB.Table("barang").Where("`KodeBarang` = ?", qd.KodeBarang).Find(&barang)

		if strings.EqualFold(qd.TipeQuantity, barang.TipeBigQty) {
			tipeqty = "BigQty"
		} else if strings.EqualFold(qd.TipeQuantity, barang.TipeMediumQty) {
			tipeqty = "MediumQty"
		} else if strings.EqualFold(qd.TipeQuantity, barang.TipeSmallQty) {
			tipeqty = "SmallQty"
		} else {
			c.JSON(fiber.Map{
				"message": "invalID type",
			})
		}
		where := "SUM(`" + tipeqty + "`)"
		database.DB.Table("stock").Select(where).Where("`KodeBarang` = ?", qd.KodeBarang).Find(&quantityavailable)

		demand := models.Demand{
			ID:               qd.ID,
			Barang:           barang,
			QuantityDemand:   qd.QuantityDemand,
			QuantityThen:     qd.QuantityAvailable,
			QuantityRightNow: quantityavailable,
			TipeQuantity:     qd.TipeQuantity,
		}
		demands = append(demands, demand)
	}

	return c.JSON(demands)
}

//DELETE DEMAND
func DeleteDemand(c *fiber.Ctx) error {
	/*
		{
			ID: //IDdemand
		}
	*/
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	var qd querydemand

	database.DB.Table("demands").Where("`ID` = ?", dataint["ID"]).Find(&qd)
	var barang models.Barang
	var tipeqty string
	var quantityavailable int
	database.DB.Table("barang").Where("`KodeBarang` = ?", qd.KodeBarang).Find(&barang)

	if strings.EqualFold(qd.TipeQuantity, barang.TipeBigQty) {
		tipeqty = "BigQty"
	} else if strings.EqualFold(qd.TipeQuantity, barang.TipeMediumQty) {
		tipeqty = "MediumQty"
	} else if strings.EqualFold(qd.TipeQuantity, barang.TipeSmallQty) {
		tipeqty = "SmallQty"
	} else {
		c.JSON(fiber.Map{
			"message": "invalID type",
		})
	}
	where := "SUM(`" + tipeqty + "`)"
	database.DB.Table("stock").Select(where).Where("`KodeBarang` = ?", qd.KodeBarang).Find(&quantityavailable)

	if quantityavailable >= qd.QuantityDemand {
		query := "DELETE FROM `demands` WHERE `demands`.`ID` = ?"
		database.DB.Exec(query, dataint["ID"])
	} else {
		fmt.Println("demand belum terpenuhi")
		return c.JSON(fiber.Map{
			"message": "demand belum terpenuhi",
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
