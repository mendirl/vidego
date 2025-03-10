package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"vidego/pkg/datatype"
	"vidego/pkg/video"
)

// var bases = []string{"/mnt/nas/misc/P"}
// var bases = []string{"/run/media/fabien/exdata/O/"}
var bases = []string{"/mnt/nas/misc/P", "/run/media/fabien/exdata/O/"}

func main() {
	fmt.Printf("#### Let's go #####\n")

	process(bases)

	fmt.Printf("#### C'est fini #####")
}

func process(bases []string) {
	files := datatype.CStringList{Value: make([]string, 0)}

	dsn := "host=db.mend.ovh user=fabien password=xxoca306 dbname=fabien port=5434 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		//NamingStrategy: schema.NamingStrategy{
		//	TablePrefix: "videogo.",
		//}
	})

	if err != nil {
		fmt.Printf("error %s", err)
	}

	ops := 0
	var wg sync.WaitGroup

	// list all files present in folders
	for _, base := range bases {
		wg.Add(1)
		go listFiles(base, &files, &wg, ops)
	}
	wg.Wait()

	// split this list into chuncks to parallize computation
	filesSlices := chunkSlice(files.Value, 50)

	// parallize treatment for each chunck
	for _, filesSlice := range filesSlices {
		ops++
		wg.Add(1)
		go reads(filesSlice, ops, &wg, db)
	}
	wg.Wait()
}

func HandlePanic(path string) {
	r := recover()

	if r != nil {
		fmt.Printf("## something is panicking with file %s : %s\n", path, r)
	}

}

// for each file, compute its size as int
// and group them by the size
func reads(files []string, ops int, wg *sync.WaitGroup, db *gorm.DB) {
	defer wg.Done()
	for _, file := range files {
		wg.Add(1)
		funcName(file, ops, db, wg)
	}
}

func funcName(file string, ops int, db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	video := createVideo(file)
	if video.Name != "empty" {
		fmt.Printf("#%d persist video with path : %s/%s \n", ops, video.Path, video.Name)
		entity := datatype.VideoEntity{Name: video.Name, Path: video.Path, Duration: video.Duration, Size: video.Size, Complete: video.Complete}
		db.Create(&entity)
	}
}

func chunkSlice(files []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(files); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond files capacity
		if end > len(files) {
			end = len(files)
		}

		chunks = append(chunks, files[i:end])
	}

	return chunks
}

func listFiles(base string, files *datatype.CStringList, wg *sync.WaitGroup, ops int) {
	defer HandlePanic("")
	defer wg.Done()

	err := filepath.Walk(base,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".mp4") {
				files.Lock()
				files.Value = append(files.Value, path)
				files.Unlock()
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
}

func createVideo(path string) datatype.Video {
	fmt.Printf("#%d video path : %s \n", path)
	defer HandlePanic(path)

	info, err := os.Stat(path)

	if err != nil {
		fmt.Printf("## ERROR with Stat : %s \n", err)
	} else {
		duration, err := video.ComputeDuration(path)
		if err != nil {
			fmt.Printf("#ERROR with vidio.NewVideo and file %s: %s \n", path, err)
		}

		split := strings.Split(path, "/")
		name := split[len(split)-1]
		sourcePath := trimSuffix(path, "/"+name)

		return datatype.Video{name, sourcePath, info.Size(), duration, duration == 0}
	}

	return datatype.Video{"empty", path, 0, 0, false}
}

func trimSuffix(s string, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}
