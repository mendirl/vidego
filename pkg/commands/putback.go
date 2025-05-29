package commands

import (
	"github.com/spf13/cobra"
	"log"
	"strings"
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
								where deduplicate is false`

func processPutback() {
	db := database.Connect()

	var dedups []datatype.VideoEntity
	db.Raw(sqlRequestPutback).Scan(&dedups)

	log.Printf("dedup size list %d\n", len(dedups))

	for _, dedup := range dedups {
		log.Printf("putback %s\n", dedup.Name)

		src := dedup.Path + "/" + dedup.Name
		newDstPath := strings.ReplaceAll(dedup.Path, "/dedup", "")
		dst := newDstPath + "/" + dedup.Name

		if utils.MoveFile(src, dst) {
			db.Save(&dedup).Update("path", newDstPath).Update("deduplicate", true)
		} else {
			db.Delete(&dedup)
		}

	}
}
