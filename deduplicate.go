package main

import (
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io/fs"
	"log"
	"os"
)

type Dedup struct {
	Id       uint
	Name     string
	Path     string
	Duration uint
}

var sql_request_dedup = `select path, name, duration from video where duration in (select duration from video group by duration having count(1) > 1) and path like '%/P/%'`

//var sql_update_dedup_id = `update video set deduplicate = true where id = ?`
//
//var sql_update_dedup = `update video set deduplicate = true where duration > 3600 and deduplicate = false`

func main() {
	dsn := "host=localhost user=videogo password=videogo dbname=videogo port=5431 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	//dest_base := "/run/media/fabien/exdata/D/D1_over60/"
	//dest_base := "/run/media/fabien/exdata/D/D2_under60/"
	//dest_base := "/run/media/fabien/exdata/D/D3_under40/"
	//dest_base := "/run/media/fabien/exdata/D/D4_under20/"

	if err != nil {
		log.Fatal(err)
	}

	var dedups []Dedup
	db.Raw(sql_request_dedup).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))

	nbfiles := 0

	for _, dedup := range dedups {
		if nbfiles == 500 {
			break
		}
		log.Printf("dedup %s\n", dedup.Name)
		source := dedup.Path + "/" + dedup.Name
		dest := dedup.Path + "/dedup/" + dedup.Name

		if moveFile(source, dest) {
			nbfiles++
		}
	}
}

func moveFile(source string, dest string) bool {
	if exists(source) {
		log.Println("move file ", source, " to ", dest)
		err := os.Rename(source, dest)
		if err != nil {
			log.Fatal(err)
			return false
		}
	}

	return true
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return false
}
