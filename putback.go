package main

import (
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io/fs"
	"log"
	"os"
	"strings"
)

type VideoEntity struct {
	gorm.Model
	Name     string
	Path     string
	Size     int64
	Duration float64
	Complete bool
}

type Tabler interface {
	TableName() string
}

func (VideoEntity) TableName() string {
	return "video"
}

var sql_request_putback = `select path, name, duration, size
							from video
							where path like '%dedup%';`

func main() {
	dsn := "host=localhost user=videogo password=videogo dbname=videogo port=5431 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	var dedups []VideoEntity
	db.Raw(sql_request_putback).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))

	for _, dedup := range dedups {
		log.Printf("putback %s\n", dedup.Name)

		source := dedup.Path + "/" + dedup.Name

		dest := strings.ReplaceAll(dedup.Path, "/dedup", "") + "/" + dedup.Name
		result := MoveFile(source, dest)
		if result {
			// update db
			db.Save(&dedup).Update("path", dest)
		} else {
			// remove from db
			db.Delete(&dedup)
		}
	}
}

func DeleteFile(path string) bool {
	if exists(path) {
		err := os.Remove(path)
		if err != nil {
			log.Fatal(err)
			return false
		}
		return true
	}
	return false
}
