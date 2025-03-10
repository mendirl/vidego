package utils

import (
	"errors"
	"io/fs"
	"log"
	"os"
)

func MoveFile(source string, dest string) bool {
	if exists(source) {
		log.Println("move file ", source, " to ", dest)
		err := os.Rename(source, dest)
		if err != nil {
			log.Fatal(err)
			return false
		}
		return true

	}

	return false
}

func DeleteFile(path string) bool {
	if exists(path) {
		err := os.Remove(path)
		if err != nil {
			log.Fatal(err)
			return false
		}
		return true
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return false
}
