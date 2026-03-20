package commands

import (
	"log"
	"strings"
	"sync"
	"vidego/pkg/database"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var sqlRequestDedup = `WITH doublons AS (
    SELECT
        REPLACE(path, '/_dedup', '') as path_normalise,
        FLOOR(duration) as duration_entiere
		FROM vidego.video
		WHERE path IS NOT NULL
		  AND duration IS NOT NULL
		GROUP BY REPLACE(path, '/_dedup', ''), FLOOR(duration)
		HAVING COUNT(*) > 1
	),
	all_dedup as (
	SELECT v.*
	FROM vidego.video v
	INNER JOIN doublons d
		ON REPLACE(v.path, '/_dedup', '') = d.path_normalise
		AND FLOOR(v.duration) = d.duration_entiere
	ORDER BY REPLACE(v.path, '/_dedup', ''), FLOOR(v.duration), v.id)
	select * from all_dedup where
    path not like '%_dedup%'`

func newDedupCommand() *cobra.Command {
	c := &cobra.Command{
		Use:  "dedup",
		Long: "from db, move duplicate video to _dedup folder",
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
	var dest string
	if strings.Contains(dedup.Path, "_dedup/_dedup") {
		dest = strings.Replace(dedup.Path, "_dedup/_dedup", "_dedup", 1)
	} else if !strings.Contains(dedup.Path, "_dedup") {
		dest = dedup.Path + "/_dedup"
	} else {
		dest = dedup.Path
	}

	if utils.MoveAndCheckFile(source, dest, dedup.Name) {
		db.Save(&dedup).Update("path", dest).Update("deduplicate", true)
	}
}
