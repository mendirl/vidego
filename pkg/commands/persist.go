package commands

import (
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"vidego/pkg/database"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
	"vidego/pkg/video"
)

func newPersistCommand() *cobra.Command {

	var (
		path []string
	)

	c := &cobra.Command{
		Use: "persist",
		Run: func(cmd *cobra.Command, args []string) {
			processPersist(path)
		},
	}

	c.PersistentFlags().StringSliceVar(&path, "path", []string{}, "")

	return c
}

func processPersist(bases []string) {
	files := datatype.CStringList{Value: make([]string, 0)}

	db := database.Connect()

	var wg sync.WaitGroup

	// list all files present in folders
	for _, base := range bases {
		wg.Add(1)
		go listFiles(base, &files, &wg)
	}
	wg.Wait()

	// split this list into chunks to paralyze computation
	filesSlices := chunkSlice(files.Value, 50)

	// paralyze treatment for each chunk
	for _, filesSlice := range filesSlices {
		wg.Add(1)
		go reads(filesSlice, &wg, db)
	}
	wg.Wait()
}

// for each file, compute its size as int
// and group them by the size
func reads(files []string, wg *sync.WaitGroup, db *gorm.DB) {
	defer wg.Done()
	for _, file := range files {
		wg.Add(1)
		funcName(file, db, wg)
	}
}

func funcName(file string, db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	newVideo := video.CreateVideo(file)
	if newVideo.Name != "empty" {
		log.Printf("# persist video with path : %s/%s \n", newVideo.Path, newVideo.Name)
		entity := datatype.VideoEntity{Name: newVideo.Name, Path: newVideo.Path, Duration: newVideo.Duration, Size: newVideo.Size, Complete: newVideo.Complete}
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

func listFiles(base string, files *datatype.CStringList, wg *sync.WaitGroup) {
	defer utils.HandlePanic("")
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
		log.Println(err)
	}
}
