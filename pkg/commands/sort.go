package commands

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"vidego/pkg/database"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
	"vidego/pkg/video"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func newSortCommand() *cobra.Command {

	var (
		paths   []string
		cfgFile string
	)

	c := &cobra.Command{
		Use:  "sort",
		Long: "sort videos into folders",
		Run: func(cmd *cobra.Command, args []string) {
			paths = viper.GetStringSlice("paths")
			processSort(paths)
		},
	}

	cobra.OnInitialize(func() { initConfig(cfgFile) })

	c.PersistentFlags().StringSliceVar(&paths, "paths", []string{}, "")
	err := viper.BindPFlag("paths", c.PersistentFlags().Lookup("paths"))
	if err != nil {
		return nil
	}

	c.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vidego.yaml)")
	viper.BindPFlag("config", c.PersistentFlags().Lookup("config"))

	return c
}

var sqlRequestConfig = `select * from vidego.config order by position`

func processSort(paths []string) {
	db := database.Connect()

	var configs []datatype.ConfigEntity
	db.Raw(sqlRequestConfig).Scan(&configs)

	var wg sync.WaitGroup

	for _, path := range paths {
		log.Printf("# let's analyze folder : %s\n", path)
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
			}()
			sortFolder(path, configs, db)
		}()
	}

	wg.Wait()
}

func sortFolder(path string, configs []datatype.ConfigEntity, db *gorm.DB) {
	const maxGoroutines = 10
	semaphore := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".mp4") {
				return nil
			}

			semaphore <- struct{}{}
			wg.Add(1)
			go func(filePath string) {
				defer func() {
					<-semaphore
					wg.Done()
				}()
				handleFile(filePath, configs, db)
			}(path)

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	wg.Wait()
}

func handleFile(path string, configs []datatype.ConfigEntity, db *gorm.DB) {
	newVideo := video.CreateVideo(path)

	if newVideo.Duration == 0 {
		return
	}

	var dst, src string

	var match, config = findInConfigs(newVideo.Name, configs)

	src = newVideo.Path
	if match {
		dst = computeNamedNaseFolder(path, config)
	} else {
		dst = computeOtherNameFolder(path, newVideo)
	}

	if utils.MoveAndCheckFile(src, dst, newVideo.Name) {
		newVideo.Path = dst
		persistVideo(newVideo, db)
	}
}

func persistVideo(newVideo datatype.Video, db *gorm.DB) {
	entity := datatype.VideoEntity{Name: newVideo.Name, Path: newVideo.Path, Duration: newVideo.Duration, Size: newVideo.Size, Complete: newVideo.Complete}
	db.Create(&entity)
}

func computeNamedNaseFolder(path string, config string) string {
	base := findBase(path) + "/N/"
	return base + config
}

func computeOtherNameFolder(path string, video datatype.Video) string {
	duration := video.Duration
	base := findBase(path)

	if duration < 300 {
		return base + "/O/O1_under05"
	} else if duration < 600 {
		return base + "/O/O2_under10"
	} else if duration < 900 {
		return base + "/O/O3_under15"
	} else if duration < 1200 {
		return base + "/O/O4_under20"
	} else if duration < 1500 {
		return base + "/O/O5_under25"
	} else if duration < 1800 {
		return base + "/O/O6_under30"
	} else if duration < 2100 {
		return base + "/O/O7_under35"
	} else if duration < 2400 {
		return base + "/O/O8_under40"
	} else if duration < 2700 {
		return base + "/O/O9_under45"
	} else if duration < 3000 {
		return base + "/O/O10_under50"
	} else if duration < 3300 {
		return base + "/O/O11_under55"
	} else if duration < 3600 {
		return base + "/O/O12_under60"
	} else if duration < 3900 {
		return base + "/O/O13_under65"
	} else if duration < 4200 {
		return base + "/O/O14_under70"
	} else if duration > 7200 {
		return base + "/O/O17_over120"
	} else if duration > 5400 {
		return base + "/O/O16_over90"
	} else {
		return base + "/O/O15_over70"
	}
}

func findBase(path string) string {
	if strings.Contains(path, "/d/") {
		return "/mnt/d"
	} else if strings.Contains(path, "/e/") {
		return "/mnt/e"
	} else if strings.Contains(path, "/f/") {
		return "/mnt/f"
	} else if strings.Contains(path, "/g/") {
		return "/mnt/g"
	} else if strings.Contains(path, "/h/") {
		return "/mnt/h"
	} else if strings.Contains(path, "/n/") {
		return "/mnt/n"
	} else {
		return "/mnt/n/T"
	}
}

func findInConfigs(videoName string, configs []datatype.ConfigEntity) (bool, string) {
	for _, config := range configs {
		for _, value := range config.Values {
			if containsWord(videoName, value) {
				return true, config.Name
			}
		}
	}
	return false, ""
}

func containsWord(word, subword string) bool {
	word = strings.ToLower(word)
	subword = strings.ToLower(subword)

	if strings.Contains(subword, " ") {
		split := strings.Split(subword, " ")
		for _, w := range split {
			if !strings.Contains(word, w) {
				return false
			}
		}
		return true
	} else {
		return strings.Contains(word, subword)
	}
}
