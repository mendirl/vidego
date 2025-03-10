package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
)

var sql_request_dedup = `select *
							from videogo.video
								where duration in (select duration from videogo.video group by duration having count(1) > 1)`

func main() {
	dsn := "host=db.mend.ovh user=fabien password=xxoca306 dbname=fabien port=5434 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	var dedups []datatype.VideoEntity
	db.Raw(sql_request_dedup).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))

	nbfiles := 0

	for _, dedup := range dedups {
		log.Printf("dedup %s\n", dedup.Name)
		source := dedup.Path + "/" + dedup.Name
		dest := dedup.Path + "/dedup/" + dedup.Name

		if utils.MoveFile(source, dest) {
			nbfiles++
		}
	}
}
