package database

import (
	"fmt"

	"github.com/kamil5b/API-Obat/models"
	"github.com/kamil5b/API-Obat/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	user := utils.GoDotEnvVariable("DATABASE_USER")
	password := utils.GoDotEnvVariable("DATABASE_PASSWORD")
	url := utils.GoDotEnvVariable("DATABASE_URL")
	protocol := utils.GoDotEnvVariable("DSN_PROTOCOL")
	database := utils.GoDotEnvVariable("DATABASE_NAME")

	dsn := user + ":" + password + "@" + protocol + "(" + url + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
	connection, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("could not connect to the database")
	}

	DB = connection
	connection.AutoMigrate(&models.User{})
	var result int
	DB.Table("hutang").Select("`SisaHutang`").Where("`NomorHutang` = 1").Find(&result)
	fmt.Println(result)
}
