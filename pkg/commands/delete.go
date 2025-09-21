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

func newDeleteCommand() *cobra.Command {
	c := &cobra.Command{
		Use:  "delete",
		Long: "delete all videos marked as to_delete in db",
		Run: func(cmd *cobra.Command, args []string) {
			processDelete()
		},
	}

	return c
}

var sqlRequestDelete = `select *
							from vidego.video
							where to_delete is true`

func processDelete() {
	db := database.Connect()

	var toDeletes []datatype.VideoEntity
	db.Raw(sqlRequestDelete).Scan(&toDeletes)

	const maxGoroutines = 5
	semaphore := make(chan struct{}, maxGoroutines)

	var size = len(toDeletes)
	var counter = size
	var counterMutex sync.Mutex

	log.Printf("Starting to delete %d videos\n", size)

	var wg sync.WaitGroup
	for _, toDelete := range toDeletes {
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
				log.Printf("Delete video: %s. Remaining: %d/%d\n", toDelete.Name, remaining, size)
			}()
			deleteVideo(toDelete, db)
		}()
	}
	wg.Wait()
	log.Printf("All %d videos have been deleted\n", size)
}

func deleteVideo(toDelete datatype.VideoEntity, db *gorm.DB) {
	log.Printf("Delete %s\n", toDelete.Name)

	src := toDelete.Path + "/" + toDelete.Name

	if utils.DeleteFile(src) {
		db.Delete(&toDelete)
	}

}
