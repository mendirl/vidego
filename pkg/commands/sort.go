package commands

import (
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
		paths     []string
		move      bool
		search    bool
		toUniqueT bool
		cfgFile   string
	)

	c := &cobra.Command{
		Use:  "sort",
		Long: "sort videos into folders",
		Run: func(cmd *cobra.Command, args []string) {
			paths = viper.GetStringSlice("paths")
			move = viper.GetBool("move")
			search = viper.GetBool("search")
			toUniqueT = viper.GetBool("uniqueSearch")
			processSort(paths, move, search, toUniqueT)
		},
	}

	cobra.OnInitialize(func() { initConfig(cfgFile) })

	c.PersistentFlags().StringSliceVar(&paths, "paths", []string{}, "")
	viper.BindPFlag("paths", c.PersistentFlags().Lookup("paths"))

	c.PersistentFlags().BoolVar(&move, "move", true, "move to O folder")
	viper.BindPFlag("move", c.PersistentFlags().Lookup("move"))

	c.PersistentFlags().BoolVar(&search, "search", true, "search in config to move to named folders")
	viper.BindPFlag("search", c.PersistentFlags().Lookup("search"))

	c.PersistentFlags().BoolVar(&toUniqueT, "uniqueSearch", false, "move all found in config to unique folder _ALL")
	viper.BindPFlag("uniqueSearch", c.PersistentFlags().Lookup("uniqueSearch"))

	c.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vidego.yaml)")
	viper.BindPFlag("config", c.PersistentFlags().Lookup("config"))

	return c
}

var sqlRequestConfig = `select * from vidego.config order by position`

func processSort(paths []string, move bool, search bool, toUniqueT bool) {
	log.Printf("Sort command parameters: paths=%v, move=%v, search=%v, uniqueSearch=%v\n", paths, move, search, toUniqueT)
	db := database.Connect()

	var configs []datatype.ConfigEntity
	if search {
		db.Raw(sqlRequestConfig).Scan(&configs)
	}

	log.Printf("nb of config : %d\n", len(configs))

	var wg sync.WaitGroup

	for _, path := range paths {
		log.Printf("# let's analyze folder : %s\n", path)
		wg.Add(1)
		go func(p string) {
			defer func() {
				wg.Done()
			}()
			sortFolder(p, configs, db, move, search, toUniqueT)
		}(path)
	}

	wg.Wait()
}

func sortFolder(path string, configs []datatype.ConfigEntity, db *gorm.DB, move bool, search bool, toUniqueT bool) {
	const maxGoroutines = 10
	semaphore := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("error accessing path %q: %v\n", path, err)
				return nil
			}
			if info.IsDir() {
				return nil
			}
			lowerPath := strings.ToLower(path)
			if !(strings.HasSuffix(lowerPath, ".mp4") || strings.HasSuffix(lowerPath, ".mkv") || strings.HasSuffix(lowerPath, ".avi") || strings.HasSuffix(lowerPath, ".wmv") || strings.HasSuffix(lowerPath, ".mov")) {
				return nil
			}

			semaphore <- struct{}{}
			wg.Add(1)
			go func(filePath string) {
				defer func() {
					<-semaphore
					wg.Done()
				}()
				handleFile(filePath, configs, db, move, search, toUniqueT)
			}(path)

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	wg.Wait()
}

func handleFile(path string, configs []datatype.ConfigEntity, db *gorm.DB, move bool, search bool, toUniqueT bool) {
	newVideo, err := video.CreateVideo(path)
	var dst, src string

	if err != nil || (newVideo.Duration == 0 && !newVideo.Complete) {
		if !move {
			return
		}
		var name = filepath.Base(path)
		var match, _ = findInConfigs(name, configs)
		if match {
			dst = findBase(path) + "/P"
		} else {
			dst = findBase(path) + "/O/O"
		}
		utils.MoveAndCheckFile(filepath.ToSlash(filepath.Dir(path)), dst, name)
		return
	}

	src = newVideo.Path

	if search {
		var match, config = findInConfigs(newVideo.Name, configs)

		if match {
			dst = computeNamedNaseFolder(path, config, toUniqueT)
			if utils.MoveAndCheckFile(src, dst, newVideo.Name) {
				newVideo.Path = dst
				persistVideo(newVideo, db)
				return
			}
		}
	}

	if !move {
		return
	}
	dst = computeOtherNameFolder(path, newVideo)

	if dst == "" {
		return
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

func computeNamedNaseFolder(path string, config string, toUniqueT bool) string {
	base := findBase(path) + "/N/"
	if toUniqueT {
		return base + "_ALL"
	}
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
	path = filepath.ToSlash(path)
	re := regexp.MustCompile(`/([cdefghjnx])/`)
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

func containsWord(videoName, subword string) bool {
	videoName = strings.ToLower(videoName)
	subword = strings.ToLower(subword)

	if strings.Contains(subword, " ") {
		split := strings.Split(subword, " ")
		for _, w := range split {
			if !strings.Contains(videoName, w) {
				return false
			}
		}
		return true
	} else {
		return strings.Contains(videoName, subword)
	}
}
