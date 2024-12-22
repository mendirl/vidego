package main

//
//import (
//	"fmt"
//	vidio "github.com/AlexEidt/Vidio"
//	"gorm.io/driver/postgres"
//	"gorm.io/gorm"
//	"log"
//	"os"
//	"path/filepath"
//	"slices"
//	"strings"
//	"sync"
//)
//
//type CMap = struct {
//	sync.RWMutex
//	value map[uint][]string
//}
//type CSet = struct {
//	sync.RWMutex
//	value map[uint]void
//}
//
//type CStringList = struct {
//	sync.RWMutex
//	value []string
//}
//
//type Video = struct {
//	Name     string
//	Path     string
//	Size     int64
//	Duration uint
//}
//
//type VideoEntity struct {
//	gorm.Model
//	Name     string
//	Path     string
//	Size     int64
//	Duration uint
//}
//
//type Tabler interface {
//	TableName() string
//}
//
//func (VideoEntity) TableName() string {
//	return "video"
//}
//
//type void struct{}
//
//var member void
//
//func main() {
//	//	base := "/media/fabien/exdata/A1_over60"
//	//  base := "/home/fabien/Videos"
//	//	base := "/run/media/fabien/exdata/O"
//	//	base := "/run/media/fabien/data/O"
//	dest := "/mnt/share/misc/P/"
//
//	bases := []string{"/run/media/fabien/exdata/O", "/run/media/fabien/exdata/A" /*, "/mnt/share/misc/P/O"*/}
//
//	process(bases, dest)
//
//	fmt.Printf("#### C'est fini #####")
//}
//
//func process(bases []string, dest string) {
//	pathsByDurations := CMap{value: make(map[uint][]string)}
//	durations := CSet{value: make(map[uint]void)}
//	files := CStringList{value: make([]string, 0)}
//
//	dsn := "host=localhost user=myuser password=secret dbname=trygo port=5432 sslmode=disable"
//	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ops := 0
//	var wg sync.WaitGroup
//
//	// list all files present in folders
//	for _, base := range bases {
//		wg.Add(1)
//		go listFiles(base, &files, &wg)
//	}
//	wg.Wait()
//
//	// split this list into chuncks to parallize computation
//	filesSlices := chunkSlice(files.value, 50)
//
//	// parallize treatment for each chunck
//	for _, filesSlice := range filesSlices {
//		ops++
//		wg.Add(1)
//		go reads(filesSlice, ops, &wg, &pathsByDurations, &durations, db)
//	}
//	wg.Wait()
//
//	move(&pathsByDurations, &durations, dest)
//}
//
//func move(result *CMap, keys *CSet, base string) {
//	result.RLock()
//	keys.RLock()
//
//	durations := make([]uint, 0, len(keys.value))
//	for k := range keys.value {
//		durations = append(durations, k)
//	}
//
//	slices.Sort(durations)
//
//	for _, duration := range durations {
//		nb := len(result.value[duration])
//		if nb > 1 {
//			fmt.Printf("duration %d has multiples files %d :\n", duration, nb)
//			for _, path := range result.value[duration] {
//				fmt.Printf("%v\n", path)
//				split := strings.Split(path, "/")
//				destPath := base + "/verify/" + split[len(split)-1]
//				fmt.Printf("will move to %v\n", destPath)
//				//				err := os.Rename(path, destPath)
//				//				if err != nil {
//				//					log.Fatal(err)
//				//				}
//			}
//		}
//	}
//
//	keys.RUnlock()
//	result.RUnlock()
//}
//func HandlePanic() {
//	r := recover()
//
//	if r != nil {
//		fmt.Println("RECOVER", r)
//	}
//}
//
//// for each file, compute its size as int
//// and group them by the size
//func reads(files []string, ops int, wg *sync.WaitGroup, pathsByDurations *CMap, durations *CSet, db *gorm.DB) {
//	for _, file := range files {
//		video := createVideo(file)
//		entity := VideoEntity{Name: video.Name, Path: video.Path, Duration: video.Duration, Size: video.Size}
//		db.Create(&entity)
//		read(video, ops, pathsByDurations, durations)
//	}
//	defer wg.Done()
//}
//
//func read(video Video, ops int, pathsByDurations *CMap, durations *CSet) {
//	durations.Lock()
//	pathsByDurations.Lock()
//
//	fmt.Printf("#%d - %s/%s, %db, %ds\n", ops, video.Path, video.Name, video.Size, video.Duration)
//	durations.value[video.Duration] = member
//	pathsByDurations.value[video.Duration] = append(pathsByDurations.value[video.Duration], video.Path)
//
//	pathsByDurations.Unlock()
//	durations.Unlock()
//}
//
//func chunkSlice(files []string, chunkSize int) [][]string {
//	var chunks [][]string
//	for i := 0; i < len(files); i += chunkSize {
//		end := i + chunkSize
//
//		// necessary check to avoid slicing beyond files capacity
//		if end > len(files) {
//			end = len(files)
//		}
//
//		chunks = append(chunks, files[i:end])
//	}
//
//	return chunks
//}
//
//func listFiles(base string, files *CStringList, wg *sync.WaitGroup) {
//	defer HandlePanic()
//	defer wg.Done()
//
//	err := filepath.Walk(base,
//		func(path string, info os.FileInfo, err error) error {
//			if err != nil {
//				return err
//			}
//
//			if !info.IsDir() && strings.HasSuffix(path, ".mp4") {
//				files.Lock()
//				files.value = append(files.value, path)
//				files.Unlock()
//			}
//			return nil
//		})
//	if err != nil {
//		log.Println(err)
//	}
//}
//
//func createVideo(path string) Video {
//	video, err := vidio.NewVideo(path)
//	info, err := os.Stat(path)
//
//	if err != nil {
//		fmt.Printf("ERROR : %s", err)
//	}
//
//	defer func() {
//		if r := recover(); r != nil {
//			video.Duration()
//			fmt.Println("Recovered in f", r)
//		}
//	}()
//
//	duration := uint(video.Duration())
//
//	split := strings.Split(path, "/")
//	name := split[len(split)-1]
//	sourcePath := TrimSuffix(path, "/"+name)
//
//	return Video{name, sourcePath, info.Size(), duration}
//}
//
//func TrimSuffix(s, suffix string) string {
//	if strings.HasSuffix(s, suffix) {
//		s = s[:len(s)-len(suffix)]
//	}
//	return s
//}
