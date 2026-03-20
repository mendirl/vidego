package commands

import (
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
	"vidego/pkg/utils"
	"vidego/pkg/video"

	"github.com/spf13/cobra"
)

func newDedupFileCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "dedupFile [directory]",
		Short: "Move duplicate video files (by duration) to a single _dedup folder at root directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir := args[0]
			err := processDedupLocal(dir)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		},
	}

	return c
}

func processDedupLocal(root string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	dedupRoot := filepath.Join(absRoot, "_dedup")

	var wg sync.WaitGroup
	const maxGoroutines = 10
	semaphore := make(chan struct{}, maxGoroutines)

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip '_dedup' folders themselves to avoid infinite loops or processing already deduped files
			if d.Name() == "_dedup" {
				return filepath.SkipDir
			}

			wg.Add(1)
			semaphore <- struct{}{}
			go func(dirPath string) {
				defer wg.Done()
				defer func() { <-semaphore }()

				if err := processDirectory(dirPath, dedupRoot); err != nil {
					log.Printf("Error processing directory %s: %v\n", dirPath, err)
				}
			}(path)

			return nil
		}
		return nil
	})

	wg.Wait()
	return err
}

func processDirectory(dir string, dedupRoot string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Group files by duration
	filesByDuration := make(map[int64][]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		v := video.CreateVideo(filePath)

		if v.Duration == 0 {
			log.Printf("Warning: duration is 0 for %s, skipping\n", entry.Name())
			continue
		}

		duration := int64(math.Floor(v.Duration))
		filesByDuration[duration] = append(filesByDuration[duration], entry.Name())
	}

	// For each duration group with more than 1 file, move duplicates
	for duration, files := range filesByDuration {
		if len(files) <= 1 {
			continue
		}

		log.Printf("Found %d files with duration %d seconds in %s\n", len(files), duration, dir)

		// Move all files in the group
		for i := 0; i < len(files); i++ {
			fileName := files[i]
			source := filepath.Join(dir, fileName)

			log.Printf("Moving %s to %s\n", fileName, dedupRoot)
			if !utils.MoveAndCheckFile(dir, dedupRoot, fileName) {
				log.Printf("Failed to move %s to %s\n", source, dedupRoot)
			}
		}
	}

	return nil
}
