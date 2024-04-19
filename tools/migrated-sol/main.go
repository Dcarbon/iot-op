package main

import (
	"log"

	"github.com/Dcarbon/go-shared/libs/dbutils"
	"github.com/Dcarbon/go-shared/libs/utils"
	"github.com/Dcarbon/iot-op/internal/models"
	"gorm.io/gorm"
)

var oldDbUrl = utils.StringEnv("OLD_DB_URL", "postgres://admin:244466666@10.60.0.58/iott?sslmode=disable")
var newDbUrl = utils.StringEnv("DB_URL", "postgres://admin:hellosecret@testnet.dcarbon.org/iot_op?sslmode=disable")

var dbOld = dbutils.MustNewDB(oldDbUrl)
var dbNew = dbutils.MustNewDB(newDbUrl)

// 0x19adf96848504a06383b47aaa9bbbc6638e81afd
func main() {

	err := dbNew.AutoMigrate(
		&models.MintSign{},
		&models.Minted{},
	)
	utils.PanicError("Auto migrate new database", err)
	// TransformMintSign()
	TransformMinted()
}

func TransformMinted() {
	var transMinted = &dbutils.Transform[OldMinted, models.Minted]{
		TblOld:    OldTableNameMinted,
		TblNew:    models.TableNameMinted,
		OnConvert: ConvertMinted,
	}
	transMinted.OnLoaded = func(d *gorm.DB, m *models.Minted) error {
		if m.Carbon == 0 {
			return nil
		}
		return transMinted.Insert(d, m)
	}
	err := transMinted.LoadAll(dbOld, dbNew)
	log.Println("Transform minted error: ", err)
}

func TransformMintSign() {
	log.Println()
	var transMinted = &dbutils.Transform[OldMintSign, models.MintSign]{
		TblOld:    OldTableNameMintSign,
		TblNew:    models.TableNameMintSign,
		OnConvert: ConvertMintSign,
	}
	transMinted.OnLoaded = func(d *gorm.DB, m *models.MintSign) error {
		return transMinted.Insert(d, m)
	}
	err := transMinted.LoadAll(dbOld, dbNew)
	log.Println("Transform mintsign error: ", err)
}
