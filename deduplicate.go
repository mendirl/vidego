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
	Id     uint
	Name   string
	Path   string
	Length uint
}

var sql_request_dedup = `select id, path, name, video.duration as length from video
         join (select count(id) as nb, duration from video group by duration having count(id) > 1) dedup
              on video.duration = dedup.duration
-- where video.duration > 3600
-- where video.duration < 3600 and video.duration >= 2400
-- where video.duration < 2400 and video.duration >= 1200
where video.duration < 1200
order by video.duration desc`

var sql_update_dedup_id = `update video set deduplicate = true where id = ?`

var sql_update_dedup = `update video set deduplicate = true where duration > 3600 and deduplicate = false`

func main() {
	dsn := "host=localhost user=myuser password=secret dbname=trygo port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	//dest_base := "/run/media/fabien/exdata/D/D1_over60/"
	//dest_base := "/run/media/fabien/exdata/D/D2_under60/"
	//dest_base := "/run/media/fabien/exdata/D/D3_under40/"
	dest_base := "/run/media/fabien/exdata/D/D4_under20/"

	if err != nil {
		log.Fatal(err)
	}

	var dedups []Dedup
	db.Raw(sql_request_dedup).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))
	idx := 0
	for _, dedup := range dedups {
		log.Printf("dedup %s\n", dedup.Name)
		source := dedup.Path + "/" + dedup.Name
		dest := dest_base + dedup.Name

		if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
			stat, err := os.Stat(source)
			log.Println("sourceInfo : ", stat)
			if err != nil {
				log.Fatal(err)
			}

			err = os.Rename(source, dest)
			if err != nil {
				log.Fatal(err)
			}

		} else {
			split := strings.Split(dest, ".")
			lastPart := split[len(split)-1]
			dest, _ = strings.CutSuffix(dest, lastPart)
			dest = dest + fmt.Sprint("_", idx) + lastPart
		}

		db.Raw(sql_update_dedup_id, dedup.Id)
		log.Printf("will move %s to %s\n", source, dest)
	}

	// move all not duplicate

}
