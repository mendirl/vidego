package commands

import (
	"log"
	"sync"
	"vidego/pkg/database"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var sqlRequestDedup = `select *
						from vidego.video
						where path not like '%dedup%'
						  and duration in (select duration from vidego.video where path not like '%dedup' group by duration having count(1) > 1)`

func newDedupCommand() *cobra.Command {
	c := &cobra.Command{
		Use:  "dedup",
		Long: "from db, move duplicate video to dedup folder",
		Run: func(cmd *cobra.Command, args []string) {
			processDedup()
		},
	}

	return c
}

func processDedup() {
	db := database.Connect()

	var dedups []datatype.VideoEntity
	db.Raw(sqlRequestDedup).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))

	var size = len(dedups)
	var counter = size
	var counterMutex sync.Mutex

	const maxGoroutines = 10
	semaphore := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	log.Printf("Starting to process %d videos\n", size)

	for _, dedup := range dedups {

		semaphore <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-semaphore
				counterMutex.Lock()
				counter--
				remaining := counter
				counterMutex.Unlock()
				wg.Done()
				log.Printf("Processed video: %s. Remaining: %d/%d\n", dedup.Name, remaining, size)
			}()
			move(dedup, db)
		}()
	}

	wg.Wait()
	log.Printf("All %d videos have been processed\n", size)
}

func move(dedup datatype.VideoEntity, db *gorm.DB) {

	log.Printf("dedup %s\n", dedup.Name)
	source := dedup.Path
	dest := dedup.Path + "/dedup"

	if utils.MoveAndCheckFile(source, dest, dedup.Name) {
		db.Save(&dedup).Update("path", dedup.Path+"/dedup").Update("deduplicate", true)
	}
}
