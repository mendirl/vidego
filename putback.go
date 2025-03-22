package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
)

var sql_request_putback = `select *
							from videogo.video
								where duration in (select duration from videogo.video group by duration having count(1) > 1)`

/*
`select path, name, duration, size
from videogo.video
where path like '%dedup%';`
*/
func main() {
	dsn := "host=db.mend.ovh user=fabien password=xxoca306 dbname=fabien port=5434 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	var dedups []datatype.VideoEntity
	db.Raw(sql_request_putback).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))

	for _, dedup := range dedups {
		log.Printf("putback %s\n", dedup.Name)

		source := dedup.Path + "/dedup/" + dedup.Name
		dest := dedup.Path + "/" + dedup.Name

		//dest := strings.ReplaceAll(dedup.Path, "/dedup", "") + "/" + dedup.Name
		result := utils.MoveFile(source, dest)
		existsInDest := utils.Exists(dest)
		if !result && !existsInDest {
			// remove from db
			db.Delete(&dedup)
		}
	}
}
