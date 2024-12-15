package main

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
)

type Dedup struct {
	Name   string
	Path   string
	Length uint
}

var sql_request = `select path, name, video.duration as length from video
						join (select count(id) as nb, duration from video 
									group by duration having count(id) > 1) dedup
						on video.duration = dedup.duration 
						where video.duration >3600
						order by video.duration desc`

func main() {
	dsn := "host=localhost user=myuser password=secret dbname=trygo port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	dest_base := "/run/media/fabien/exdata/dedup"

	if err != nil {
		log.Fatal(err)
	}

	var dedups []Dedup
	db.Raw(sql_request).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))
	idx := 0
	for _, dedup := range dedups {
		log.Printf("dedup %s\n", dedup.Name)
		source := dedup.Path + "/" + dedup.Name
		dest := dest_base + dedup.Name

		if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
			err := os.Rename(source, dest)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			split := strings.Split(dest, ".")
			lastPart := split[len(split)-1]
			dest, _ = strings.CutSuffix(dest, lastPart)
			dest = dest + fmt.Sprint("_", idx) + lastPart
		}

		log.Printf("will move %s to %s\n", source, dest)

	}

}
