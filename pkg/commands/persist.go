package commands

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"vidego/pkg/database"
	"vidego/pkg/datatype"
	"vidego/pkg/panic"
	"vidego/pkg/video"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func newPersistCommand() *cobra.Command {

	var (
		path    []string
		cfgFile string
	)

	c := &cobra.Command{
		Use:  "persist",
		Long: "from input folders, put video info in db",
		Run: func(cmd *cobra.Command, args []string) {
			path = viper.GetStringSlice("path")
			processPersist(path)
		},
	}

	cobra.OnInitialize(func() { initConfig(cfgFile) })

	c.PersistentFlags().StringSliceVar(&path, "path", []string{}, "")
	err := viper.BindPFlag("path", c.PersistentFlags().Lookup("path"))
	if err != nil {
		return nil
	}

	c.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vidego.yaml)")
	viper.BindPFlag("config", c.PersistentFlags().Lookup("config"))

	return c
}

func processPersist(bases []string) {
	files := datatype.CStringList{Value: make([]string, 0)}

	db := database.Connect()

	const maxGoroutines = 10
	semaphore := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	// list all files present in folders
	for _, base := range bases {
		log.Printf("# let's analyze folder : %s\n", base)
		semaphore <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			listFiles(base, &files)
		}()
	}
	wg.Wait()

	// split this list into chunks to paralyze computation
	//filesSlices := chunkSlice(files.Value, 50)

	// paralyze treatment for each chunk
	for idx, file := range files.Value {
		log.Printf("# let's analyze a new slice %d:\n", idx)
		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			funcName(file, db)
		}()
	}
	wg.Wait()
}

func funcName(file string, db *gorm.DB) {
	newVideo := video.CreateVideo(file)
	if newVideo.Name != "empty" {
		log.Printf("# persist video with path : %s/%s \n", newVideo.Path, newVideo.Name)
		entity := datatype.VideoEntity{Name: newVideo.Name, Path: newVideo.Path, Duration: newVideo.Duration, Size: newVideo.Size, Complete: newVideo.Complete}
		db.Create(&entity)
	}
}

func listFiles(base string, files *datatype.CStringList) {
	defer panic.HandlePanic("")

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
