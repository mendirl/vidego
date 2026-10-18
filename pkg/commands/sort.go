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
		base      string
	)

	c := &cobra.Command{
		Use:  "sort",
		Long: "sort videos into folders",
		Run: func(cmd *cobra.Command, args []string) {
			paths = viper.GetStringSlice("paths")
			move = viper.GetBool("move")
			search = viper.GetBool("search")
			toUniqueT = viper.GetBool("uniqueSearch")
			base = viper.GetString("base")
			processSort(paths, move, search, toUniqueT, base)
		},
	}

	cobra.OnInitialize(func() { initConfig(cfgFile) })

	c.PersistentFlags().StringSliceVar(&paths, "paths", []string{}, "")
	viper.BindPFlag("paths", c.PersistentFlags().Lookup("paths"))

	c.PersistentFlags().StringVar(&base, "base", "", "base directory for sorting")
	viper.BindPFlag("base", c.PersistentFlags().Lookup("base"))

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

func processSort(paths []string, move bool, search bool, toUniqueT bool, base string) {
	log.Printf("Sort command parameters: paths=%v, move=%v, search=%v, uniqueSearch=%v, base=%v\n", paths, move, search, toUniqueT, base)
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
			sortFolder(p, configs, db, move, search, toUniqueT, base)
		}(path)
	}

	wg.Wait()
}

func sortFolder(path string, configs []datatype.ConfigEntity, db *gorm.DB, move bool, search bool, toUniqueT bool, base string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("error reading directory %s: %v\n", path, err)
	} else {
		log.Printf("nb of elements in %s: %d\n", path, len(entries))
	}

	const maxGoroutines = 10
	semaphore := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	err = filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("error accessing path %q: %v\n", path, err)
				return nil
			}
			if info.IsDir() {
				return nil
			}
			lowerPath := strings.ToLower(path)
			if !(strings.HasSuffix(lowerPath, ".mp4") ||
				strings.HasSuffix(lowerPath, ".mkv") ||
				strings.HasSuffix(lowerPath, ".avi") ||
				strings.HasSuffix(lowerPath, ".wmv") ||
				strings.HasSuffix(lowerPath, ".m4v") ||
				strings.HasSuffix(lowerPath, ".mov")) {
				return nil
			}

			semaphore <- struct{}{}
			wg.Add(1)
			go func(filePath string) {
				defer func() {
					<-semaphore
					wg.Done()
				}()
				handleFile(filePath, configs, db, move, search, toUniqueT, base)
			}(path)

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	wg.Wait()
}

func handleFile(path string, configs []datatype.ConfigEntity, db *gorm.DB, move bool, search bool, toUniqueT bool, base string) {
	newVideo, err := video.CreateVideo(path)
	var dst, src string

	if err != nil || (newVideo.Duration == 0 && !newVideo.Complete) {
		if !move {
			return
		}
		var name = filepath.Base(path)
		var match, _ = findInConfigs(name, configs)
		if match {
			dst = findBase(path, base) + "/P"
		} else {
			dst = findBase(path, base) + "/O/O"
		}
		utils.MoveAndCheckFile(filepath.ToSlash(filepath.Dir(path)), dst, name)
		return
	}

	src = newVideo.Path

	if search {
		var match, config = findInConfigs(newVideo.Name, configs)

		if match {
			dst = computeNamedNaseFolder(path, config, toUniqueT, base)
			if utils.MoveAndCheckFile(src, dst, newVideo.Name) {
				newVideo.Path = dst
				persistVideo(newVideo, db)
			}
			return
		}
	}

	if !move {
		return
	}
	dst = computeOtherNameFolder(path, newVideo, base)

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

func computeNamedNaseFolder(path string, config string, toUniqueT bool, base string) string {
	b := findBase(path, base) + "/N/"
	if toUniqueT {
		return b + "_ALL"
	}
	return b + config
}

func computeOtherNameFolder(path string, video datatype.Video, base string) string {
	duration := video.Duration
	b := findBase(path, base)

	if duration < 300 {
		return b + "/O/O1_under05"
	} else if duration < 600 {
		return b + "/O/O2_under10"
	} else if duration < 900 {
		return b + "/O/O3_under15"
	} else if duration < 1200 {
		return b + "/O/O4_under20"
	} else if duration < 1500 {
		return b + "/O/O5_under25"
	} else if duration < 1800 {
		return b + "/O/O6_under30"
	} else if duration < 2100 {
		return b + "/O/O7_under35"
	} else if duration < 2400 {
		return b + "/O/O8_under40"
	} else if duration < 2700 {
		return b + "/O/O9_under45"
	} else if duration < 3000 {
		return b + "/O/O10_under50"
	} else if duration < 3300 {
		return b + "/O/O11_under55"
	} else if duration < 3600 {
		return b + "/O/O12_under60"
	} else if duration < 3900 {
		return b + "/O/O13_under65"
	} else if duration < 4200 {
		return b + "/O/O14_under70"
	} else if duration > 7200 {
		return b + "/O/O17_over120"
	} else if duration > 5400 {
		return b + "/O/O16_over90"
	} else {
		return b + "/O/O15_over70"
	}
}

func findBase(path string, base string) string {
	if base != "" {
		return strings.TrimRight(filepath.ToSlash(base), "/")
	}
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
