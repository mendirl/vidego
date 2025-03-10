package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"strings"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
)

var sql_request_putback = `select path, name, duration, size
							from videogo.video
							where path like '%dedup%';`

func main() {
	dsn := "host=db.mend.ovh user=fabien password=xxoca306 dbname=videogo port=5434 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	var dedups []datatype.VideoEntity
	db.Raw(sql_request_putback).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))

	for _, dedup := range dedups {
		log.Printf("putback %s\n", dedup.Name)

		source := dedup.Path + "/" + dedup.Name

		dest := strings.ReplaceAll(dedup.Path, "/dedup", "") + "/" + dedup.Name
		result := utils.MoveFile(source, dest)
		if result {
			// update db
			db.Save(&dedup).Update("path", dest)
		} else {
			// remove from db
			db.Delete(&dedup)
		}
	}
}
