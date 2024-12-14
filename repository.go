package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type VideoEntity struct {
	gorm.Model
	Name     string
	Path     string
	Size     int64
	Duration uint
}

type Tabler interface {
	TableName() string
}

func (VideoEntity) TableName() string {
	return "video"
}

func connection() *gorm.DB {
	dsn := "host=localhost user=myuser password=secret dbname=trygo port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}
	return db
}
