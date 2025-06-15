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

func newPutbackCommand() *cobra.Command {
	c := &cobra.Command{
		Use:  "putback",
		Long: "from the dedup folder, put video back to its original folder",
		Run: func(cmd *cobra.Command, args []string) {
			processPutback()
		},
	}

	return c
}

var sqlRequestPutback = `select *
							from videogo.video
								where deduplicate is true`

func processPutback() {
	db := database.Connect()

	var dedups []datatype.VideoEntity
	db.Raw(sqlRequestPutback).Scan(&dedups)

	const maxGoroutines = 5
	semaphore := make(chan struct{}, maxGoroutines)

	var size = len(dedups)
	var counter = size
	var counterMutex sync.Mutex

	log.Printf("Starting to process %d videos\n", size)

	var wg sync.WaitGroup
	for _, dedup := range dedups {
		semaphore <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				<-semaphore
				counterMutex.Lock()
				counter--
				remaining := counter
				counterMutex.Unlock()
				log.Printf("Putback video: %s. Remaining: %d/%d\n", dedup.Name, remaining, size)
			}()
			moveBack(dedup, db)
		}()
	}
	wg.Wait()
	log.Printf("All %d videos have been processed\n", size)
}

func moveBack(dedup datatype.VideoEntity, db *gorm.DB) {
	log.Printf("putback %s\n", dedup.Name)

	src := dedup.Path + "/" + dedup.Name
	newDstPath := strings.ReplaceAll(dedup.Path, "/dedup", "")
	dst := newDstPath + "/" + dedup.Name

	if utils.MoveFile(src, dst) {
		db.Save(&dedup).Update("path", newDstPath).Update("deduplicate", false)
	} else {
		db.Delete(&dedup)
	}
}
