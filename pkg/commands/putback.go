package commands

import (
	"github.com/spf13/cobra"
	"log"
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
								where duration in (select duration from videogo.video group by duration having count(1) > 1)`

func processPutback() {
	db := database.Connect()

	var dedups []datatype.VideoEntity
	db.Raw(sqlRequestPutback).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))

	for _, dedup := range dedups {
		log.Printf("putback %s\n", dedup.Name)

		source := dedup.Path + "/dedup/" + dedup.Name
		dest := dedup.Path + "/" + dedup.Name

		//dest := strings.ReplaceAll(dedup.Path, "/dedup", "") + "/" + dedup.Name
		result := utils.MoveFile(source, dest)
		existsInDest := utils.Exists(dest)
		if !result && !existsInDest {
			// remove from db
			db.Delete(&dedup)
		}
	}
}
