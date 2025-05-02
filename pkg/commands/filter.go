package commands

import (
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"io/fs"
	"log"
	"os"
	"strings"
	"sync"
	"vidego/pkg/database"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
	"vidego/pkg/video"
)

func newFilterCommand() *cobra.Command {

	var (
		path string
	)

	c := &cobra.Command{
		Use:  "filter",
		Long: "from a temporary folder, put them in the right folder or dedup folder, compare from db",
		Run: func(cmd *cobra.Command, args []string) {
			processFilter(path)
		},
	}

	c.PersistentFlags().StringVar(&path, "path", "", "")

	return c
}

var sqlRequestAll = `select * from videogo.video`

func processFilter(source string) {
	var wg sync.WaitGroup

	videos := datatype.CVideoEntityList{Value: make([]datatype.VideoEntity, 0)}

	wg.Add(1)
	go computeVideos(source, &videos, &wg)

	dataInDbByDuration := datatype.CVideoEntityMap{Value: make(map[float64][]datatype.VideoEntity)}
	db := database.Connect()
	wg.Add(1)
	go computeDb(db, &dataInDbByDuration, &wg)

	wg.Wait()

	toto(&videos, &dataInDbByDuration, db)
}

func computeVideos(source string, videos *datatype.CVideoEntityList, wg *sync.WaitGroup) {
	defer wg.Done()

	files, err := os.ReadDir(source)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, file := range files {
		// define size of the file
		fileInfo, err := file.Info()
		if err != nil {
			log.Printf("impossible to open %s with error %v \n", fileInfo, err)
			continue
		}
		if fileInfo.Name() == "" {
			log.Printf("## ERROR with file %s : %v \n", fileInfo.Name(), err)
			continue
		}
		computeVideo(source, videos, fileInfo)

	}
}

func computeVideo(source string, videos *datatype.CVideoEntityList, fileInfo fs.FileInfo) {
	//define its duration for a video
	newVideo := video.CreateVideo(source + "/" + fileInfo.Name())

	log.Printf("video : %v \n", newVideo)
	entity := datatype.VideoEntity{Name: newVideo.Name, Path: newVideo.Path, Duration: newVideo.Duration, Size: newVideo.Size, Complete: newVideo.Complete}

	videos.Lock()
	videos.Value = append(videos.Value, entity)
	videos.Unlock()
}

func computeDb(db *gorm.DB, dataInDbByDuration *datatype.CVideoEntityMap, wg *sync.WaitGroup) {
	defer wg.Done()

	var videos []datatype.VideoEntity
	db.Raw(sqlRequestAll).Scan(&videos)

	for _, itVideo := range videos {
		if _, ok := dataInDbByDuration.Value[itVideo.Duration]; !ok {
			tab := make([]datatype.VideoEntity, 0)
			dataInDbByDuration.Value[itVideo.Duration] = append(tab, itVideo)
		} else {
			dataInDbByDuration.Value[itVideo.Duration] = append(dataInDbByDuration.Value[itVideo.Duration], itVideo)
		}
	}
}

func toto(videos *datatype.CVideoEntityList, dataInDbByDuration *datatype.CVideoEntityMap, db *gorm.DB) {
	for _, itVideo := range videos.Value {
		// check if the video is already in dataInDbByDuration
		// if not, copy the file into a new folder
		if _, videoOrderers := dataInDbByDuration.Value[itVideo.Duration]; !videoOrderers {
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

				oldVideos := dataInDbByDuration.Value[itVideo.Duration]
				for _, oldVideo := range oldVideos {

					src := oldVideo.Path + "/" + oldVideo.Name

					if !strings.Contains(oldVideo.Path, "dedup") {

						dstPath := oldVideo.Path + "/dedup"
						dst := dstPath + "/" + oldVideo.Name

						log.Printf("%f move old file %s to %s \n", oldVideo.Duration, src, dst)
						if utils.MoveFile(src, dst) {
							db.Save(&oldVideo).Update("path", dstPath)
						}
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
