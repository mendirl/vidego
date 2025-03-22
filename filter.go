package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
	"vidego/pkg/video"
)

var sql_request_all = `select * from videogo.video`

func main() {
	source := "/run/media/fabien/exdata/T"

	files, err := os.ReadDir(source)
	if err != nil {
		log.Fatal(err)
		return
	}
	videos := make([]datatype.VideoEntity, 0)

	for _, file := range files {
		// define size of the file
		fileInfo, err := file.Info()
		if err != nil {
			log.Printf("impossible to open %s with error %v \n", fileInfo, err)
			return
		}
		//define its duration for a video
		newVideo := video.CreateVideo(source + "/" + fileInfo.Name())

		if newVideo.Name == "" {
			log.Printf("## ERROR with file %s : %v \n", fileInfo.Name(), err)
			continue
		}
		log.Printf("video : %v \n", newVideo)
		entity := datatype.VideoEntity{Name: newVideo.Name, Path: newVideo.Path, Duration: newVideo.Duration, Size: newVideo.Size, Complete: newVideo.Complete}
		videos = append(videos, entity)

	}

	dsn := "host=db.mend.ovh user=fabien password=xxoca306 dbname=fabien port=5434 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var datas []datatype.VideoEntity
	db.Raw(sql_request_all).Scan(&datas)

	dataInDbByDuration := transform(datas)

	for _, itVideo := range videos {
		// check if the video is already in dataInDbByDuration
		// if not, copy the file into a new folder
		if _, videoOrderers := dataInDbByDuration[itVideo.Duration]; !videoOrderers {
			folder := findFolder(itVideo.Duration)

			src := itVideo.Path + "/" + itVideo.Name
			dst := folder + "/" + itVideo.Name

			log.Printf("%f move file %s to %s \n", itVideo.Duration, src, dst)
			if utils.MoveFile(src, dst) {
				itVideo.Path = folder
				db.Create(&itVideo)
			}

		} else {
			// if yes, copy both files to a new folder to dedup them
			folder := findFolder(itVideo.Duration)
			src := itVideo.Path + "/" + itVideo.Name
			dst := folder + "/dedup/" + itVideo.Name

			log.Printf("%f move new file %s to %s \n", itVideo.Duration, src, dst)
			if utils.MoveFile(src, dst) {
				itVideo.Path = folder
				db.Create(&itVideo)

				oldVideos := dataInDbByDuration[itVideo.Duration]
				for _, oldVideo := range oldVideos {

					src := oldVideo.Path + "/" + oldVideo.Name
					dst := oldVideo.Path + "/dedup/" + oldVideo.Name

					log.Printf("%f move old file %s to %s \n", oldVideo.Duration, src, dst)
					if utils.MoveFile(src, dst) {
						db.Save(&oldVideo).Update("path", oldVideo.Path+"/dedup")
					}
				}
			}

		}

	}

}

func findFolder(duration float64) string {
	if duration < 1200 {
		return "/run/media/fabien/exdata/O/O5_under20"
	} else if duration < 1800 {
		return "/run/media/fabien/exdata/O/O4_under30"
	} else if duration < 2400 {
		return "/mnt/nas/misc/P/A3_under40"
	} else if duration < 3600 {
		return "/mnt/nas/misc/P/A2_under60"
	} else {
		return "/mnt/nas/misc/P/A1_over60"
	}
}

func transform(videos []datatype.VideoEntity) map[float64][]datatype.VideoEntity {
	videoMap := make(map[float64][]datatype.VideoEntity)

	for _, itVideo := range videos {
		if _, ok := videoMap[itVideo.Duration]; !ok {
			tab := make([]datatype.VideoEntity, 0)
			videoMap[itVideo.Duration] = append(tab, itVideo)
		} else {
			videoMap[itVideo.Duration] = append(videoMap[itVideo.Duration], itVideo)
		}
	}

	return videoMap

}
