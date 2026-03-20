package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
		move    bool
		cfgFile string
	)

	c := &cobra.Command{
		Use:  "sort",
		Long: "sort videos into folders",
		Run: func(cmd *cobra.Command, args []string) {
			paths = viper.GetStringSlice("paths")
			move = viper.GetBool("move")
			processSort(paths, move)
		},
	}

	cobra.OnInitialize(func() { initConfig(cfgFile) })

	c.PersistentFlags().StringSliceVar(&paths, "paths", []string{}, "")
	viper.BindPFlag("paths", c.PersistentFlags().Lookup("paths"))
	c.PersistentFlags().BoolVar(&move, "move", true, "")
	viper.BindPFlag("move", c.PersistentFlags().Lookup("move"))

	c.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vidego.yaml)")
	viper.BindPFlag("config", c.PersistentFlags().Lookup("config"))

	return c
}

var sqlRequestConfig = `select * from vidego.config order by position`

func processSort(paths []string, move bool) {
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
			sortFolder(path, configs, db, move)
		}()
	}

	wg.Wait()
}

func sortFolder(path string, configs []datatype.ConfigEntity, db *gorm.DB, move bool) {
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
				handleFile(filePath, configs, db, move)
			}(path)

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	wg.Wait()
}

func handleFile(path string, configs []datatype.ConfigEntity, db *gorm.DB, move bool) {
	newVideo := video.CreateVideo(path)

	if newVideo.Duration == 0 {
		return
	}

	var dst, src string

	var match, config = findInConfigs(newVideo.Name, configs)

	src = newVideo.Path
	if match {
		dst = computeNamedNaseFolder(path, config)
	} else if !move {
		return
	}

	dst = computeOtherNameFolder(path, newVideo)

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

	if duration > 7200 {
		return base + "/O/O17_over120"
	}
	if duration > 5400 {
		return base + "/O/O16_over90"
	}
	if duration > 4200 {
		return base + "/O/O15_over70"
	}

	index := (duration / 300) + 1
	limit := index * 5

	return fmt.Sprintf("%s/O/O%d_under%02d", base, index, limit)
}

func findBase(path string) string {
	re := regexp.MustCompile(`/([cdefghnx])/`)
	match := re.FindStringSubmatch(path)
	if len(match) > 1 {
		return "/mnt/" + match[1]
	}
	return "/mnt/n/T"
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
