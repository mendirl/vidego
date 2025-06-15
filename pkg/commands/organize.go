package commands

import (
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"log"
	"strings"
	"sync"
	"vidego/pkg/database"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
)

func newOrganizeCommand() *cobra.Command {
	c := &cobra.Command{
		Use:  "organize",
		Long: "Organize videos into their respective folders based on metadata.",
		Run: func(cmd *cobra.Command, args []string) {
			processOrganize()
		},
	}

	return c
}

var sqlRequestConfig = `select * from videogo.config order by position`
var sqlRequestOrganize = `select * from videogo.video where complete is false and deleted_at is null;`

func processOrganize() {
	db := database.Connect()

	var configs []datatype.ConfigEntity
	db.Raw(sqlRequestConfig).Scan(&configs)

	var videos []datatype.VideoEntity
	db.Raw(sqlRequestOrganize).Scan(&videos)

	const maxGoroutines = 20
	semaphore := make(chan struct{}, maxGoroutines)

	var size = len(videos)
	var counter = size
	var counterMutex sync.Mutex

	log.Printf("Starting to process %d videos\n", size)

	var wg sync.WaitGroup
	for _, video := range videos {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(v datatype.VideoEntity) {
			defer wg.Done()
			defer func() {
				<-semaphore
				counterMutex.Lock()
				counter--
				remaining := counter
				counterMutex.Unlock()
				log.Printf("Processed video: %s. Remaining: %d/%d\n", v.Name, remaining, size)
			}()
			filter(configs, v, db)
		}(video)
	}

	wg.Wait()
	log.Printf("All %d videos have been processed\n", size)
}

func filter(configs []datatype.ConfigEntity, video datatype.VideoEntity, db *gorm.DB) {

	match := false
	for _, config := range configs {
		for _, value := range config.Values {
			if !containsWord(video.Name, value) {
				match = false
				break
			} else {
				match = true
				goto matchOk
			}
		}

	matchOk:
		if match {
			srcFile := video.Path
			dstFile := findBase(video.Path) + config.Name
			if utils.MoveAndCheckFile(srcFile, dstFile, video.Name) {
				db.Save(&video).Update("path", dstFile).Update("complete", true)
				log.Printf("Moved %s to %s for %s\n", srcFile, dstFile, video.Name)
			} else {
				log.Printf("Failed to move %s to %s for %s\n", srcFile, dstFile, video.Name)
			}
			goto nextVideo
		}
	}
nextVideo:
}

func findBase(path string) string {
	if strings.Contains(path, "nas") {
		return "/mnt/nas/misc/P/ALL/"
	} else {
		return "/run/media/fabien/exdata/O/ALL/"
	}
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
