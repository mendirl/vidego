package utils

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

func MoveFile(source string, dest string) bool {
	if Exists(source) {
		if Exists(dest) {
			log.Printf("file dst already exists %s\n", dest)
			return false
		}

		log.Printf("move file %s to %s\n", source, dest)
		err := moveFileInner(source, dest)
		if err != nil {
			log.Println(err)
			return false
		}
		return true

	} else {
		log.Printf("file src doesnt exist anymore %s\n", dest)
	}

	return false
}

func moveFileInner(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}

func DeleteFile(path string) bool {
	if Exists(path) {
		err := os.Remove(path)
		if err != nil {
			log.Fatal(err)
			return false
		}
		return true
	}
	return false
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return false
}

func TrimSuffix(s string, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}
