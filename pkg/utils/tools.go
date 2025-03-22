package utils

import (
	"log"
)

func HandlePanic(path string) {
	r := recover()

	if r != nil {
		log.Printf("## something is panicking with file %s : %s\n", path, r)
	}

}
