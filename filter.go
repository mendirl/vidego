package main

import (
	"log"
	"os"
)

func main() {
	source := "/run/media/fabien/exdata/O_T/"

	files, err := os.ReadDir(source)
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, file := range files {
		// define size of the file
		fileInfo, err := file.Info()
		//define its duration for a video

	}

	// list all files present in folders

	//dsn := "host=db.mend.ovh user=fabien password=xxoca306 dbname=videogo port=5434 sslmode=disable"
	//db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	//
	//if err != nil {
	//	log.Fatal(err)
	//}

	//var dedups []datatype.VideoEntity
	//db.Raw(sql_request_putback).Scan(&dedups)
	//
	//log.Printf("dedup size list %d\n", len(dedups))
	//
	//for _, dedup := range dedups {
	//	log.Printf("putback %s\n", dedup.Name)
	//
	//	source := dedup.Path + "/" + dedup.Name
	//
	//	dest := strings.ReplaceAll(dedup.Path, "/dedup", "") + "/" + dedup.Name
	//	result := utils.MoveFile(source, dest)
	//	if result {
	//		// update db
	//		db.Save(&dedup).Update("path", dest)
	//	} else {
	//		// remove from db
	//		db.Delete(&dedup)
	//	}
	//}
}
