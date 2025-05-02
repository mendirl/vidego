package commands

import (
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"log"
	"sync"
	"vidego/pkg/database"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
)

var sqlRequestDedup = `select *
						from videogo.video
						where path not like '%dedup%'
						  and duration in (select duration from videogo.video where path not like '%dedup' group by duration having count(1) > 1)`

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

	var wg sync.WaitGroup

	for _, dedup := range dedups {
		//wg.Add(1)
		move(dedup, db, &wg)
	}

	//wg.Wait()
}

func move(dedup datatype.VideoEntity, db *gorm.DB, wg *sync.WaitGroup) {
	//defer wg.Done()

	log.Printf("dedup %s\n", dedup.Name)
	source := dedup.Path + "/" + dedup.Name
	dest := dedup.Path + "/dedup/" + dedup.Name

	if utils.MoveFile(source, dest) {
		db.Save(&dedup).Update("path", dedup.Path+"/dedup")
	}
}
