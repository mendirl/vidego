package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
)

func MoveAndCheckFile(source string, dest string, file string) bool {
	if !Exists(dest) {
		log.Printf("destination folder %s does not exist, creating it\n", dest)
		err := os.MkdirAll(dest, 0755)
		if err != nil {
			log.Printf("error creating destination folder %s: %v\n", dest, err)
			return false
		}
	}

	return MoveFile(source+"/"+file, dest+"/"+file)
}

func MoveFile(source string, dest string) bool {
	if Exists(source) {
		if Exists(dest) {
			log.Printf("file dst already exists %s\n", dest)
			DeleteFile(dest)
		}

		log.Printf("move file %s to %s\n", source, dest)
		err := moveFileInner(source, dest)
		if err != nil {
			log.Println(err)
			return false
		}
		return true

	} else {
		log.Printf("file src doesnt exist anymore %s\n", source)
	}

	return false
}

func moveFileInner(sourcePath, destPath string) error {
	// Déplacer le fichier sans copie (rename atomique sur le même filesystem)
	if err := os.Rename(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}
	return nil

	//inputFile, err := os.Open(sourcePath)
	//if err != nil {
	//	return fmt.Errorf("couldn't open source file: %s", err)
	//}
	//outputFile, err := os.Create(destPath)
	//if err != nil {
	//	inputFile.Close()
	//	return fmt.Errorf("couldn't open dest file: %s", err)
	//}
	//defer outputFile.Close()
	//_, err = io.Copy(outputFile, inputFile)
	//inputFile.Close()
	//if err != nil {
	//	return fmt.Errorf("writing to output file failed: %s", err)
	//}
	//// The copy was successful, so now delete the original file
	//err = os.Remove(sourcePath)
	//if err != nil {
	//	return fmt.Errorf("failed removing original file: %s", err)
	//}
	//return nil
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
