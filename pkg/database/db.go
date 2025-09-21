package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	dsn := "host=localhost user=fabien password=xxoca306 dbname=fabien port=5432 sslmode=disable"
	//dsn := "host=db.mend.ovh user=fabien password=xxoca306 dbname=fabien port=5434 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	return db
}
