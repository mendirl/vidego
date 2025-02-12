package main

import (
	"fmt"
	vidio "github.com/AlexEidt/Vidio"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type CMap = struct {
	sync.RWMutex
	value map[uint][]string
}
type CSet = struct {
	sync.RWMutex
	value map[uint]void
}

type CStringList = struct {
	sync.RWMutex
	value []string
}

type Video = struct {
	Name     string
	Path     string
	Size     int64
	Duration uint
	Complete bool
}

type VideoEntity struct {
	gorm.Model
	Name     string
	Path     string
	Size     int64
	Duration uint
	Complete bool
}

type Tabler interface {
	TableName() string
}

func (VideoEntity) TableName() string {
	return "video"
}

type void struct{}

var member void

// var bases = []string{"/mnt/nas/misc/P"}
// var bases = []string{"/run/media/fabien/exdata/O/"}
var bases = []string{"/mnt/nas/misc/P", "/run/media/fabien/exdata/O/"}

//var finalPath = "/run/media/fabien/exdata/O/"

func main() {
	fmt.Printf("#### Let's go #####\n")

	process(bases)

	fmt.Printf("#### C'est fini #####")
}

func process(bases []string) {
	files := CStringList{value: make([]string, 0)}

	dsn := "host=localhost user=videogo password=videogo dbname=videogo port=5431 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

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
	filesSlices := chunkSlice(files.value, 50)

	// parallize treatment for each chunck
	for _, filesSlice := range filesSlices {
		ops++
		wg.Add(1)
		go reads(filesSlice, ops, &wg, db)
	}
	wg.Wait()
}

func HandlePanic(ops int, path string) {
	r := recover()

	if r != nil {
		fmt.Printf("## %d something is panicking with file %s : %s\n", ops, path, r)
	}

}

// for each file, compute its size as int
// and group them by the size
func reads(files []string, ops int, wg *sync.WaitGroup, db *gorm.DB) {
	for _, file := range files {
		video := createVideo(file, ops)
		if video.Name != "empty" {
			fmt.Printf("#%d persist video with path : %s/%s \n", ops, video.Path, video.Name)
			entity := VideoEntity{Name: video.Name, Path: video.Path, Duration: video.Duration, Size: video.Size, Complete: video.Complete}
			db.Create(&entity)
		}
	}
	defer wg.Done()
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

func listFiles(base string, files *CStringList, wg *sync.WaitGroup, ops int) {
	defer HandlePanic(ops, "")
	defer wg.Done()

	err := filepath.Walk(base,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".mp4") {
				files.Lock()
				files.value = append(files.value, path)
				files.Unlock()
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
}

func createVideo(path string, ops int) Video {
	fmt.Printf("#%d video path : %s \n", ops, path)
	defer HandlePanic(ops, path)

	info, err := os.Stat(path)

	if err != nil {
		fmt.Printf("##%d ERROR with Stat : %s \n", ops, err)
	} else {
		duration := computeDuration(path, ops)

		split := strings.Split(path, "/")
		name := split[len(split)-1]
		//finalPath := moveFile(path, name, duration)
		sourcePath := trimSuffix(path, "/"+name)

		return Video{name, sourcePath, info.Size(), duration, duration == 0}
	}

	return Video{"empty", path, 0, 0, false}
}

//func moveFile(path string, name string, duration uint) string {
//	var destPath string
//
//	if duration == 0 {
//		destPath = finalPath + "O5_error/" + name
//	} else if duration < 1200 {
//		destPath = finalPath + "O4_under20/" + name
//	} else if duration < 2400 && duration >= 1200 {
//		destPath = finalPath + "O3_under40/" + name
//	} else if duration < 3600 && duration >= 2400 {
//		destPath = finalPath + "O2_under60/" + name
//	} else if duration >= 3600 {
//		destPath = finalPath + "O1_over60/" + name
//	} else {
//
//	}
//
//	if destPath != "" {
//		fmt.Printf("## MOVE file %s to %s \n", path, destPath)
//		err := os.Rename(path, destPath)
//		if err != nil {
//			fmt.Printf("## ERROR with os.Rename : %s \n", err)
//		}
//	}
//
//	return destPath
//}

func computeDuration(path string, ops int) uint {
	defer HandlePanic(ops, path)

	video, err := vidio.NewVideo(path)
	if err != nil {
		fmt.Printf("#%d ERROR with vidio.NewVideo and file %s: %s \n", ops, path, err)
	}

	return uint(video.Duration())
}

func trimSuffix(s string, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}
