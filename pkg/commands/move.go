package commands

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"sync"
	"vidego/pkg/utils"
)

func newMoveCommand() *cobra.Command {

	var (
		src, dst string
	)

	c := &cobra.Command{
		Use:  "move",
		Long: "move files from one folder to another",
		Run: func(cmd *cobra.Command, args []string) {
			moveFiles(src, dst)
		},
	}

	c.PersistentFlags().StringVar(&src, "source", "", "")
	c.PersistentFlags().StringVar(&dst, "destination", "", "")

	return c
}

func moveFiles(src string, dst string) {
	files, err := os.ReadDir(src)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Move files from %s to %s\n", src, dst)

	const maxGoroutines = 5
	semaphore := make(chan struct{}, maxGoroutines)

	var counterMutex sync.Mutex

	var wg sync.WaitGroup
	for _, file := range files {
		semaphore <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-semaphore
				counterMutex.Lock()
				counterMutex.Unlock()
				log.Printf("Processed video: %s\n", file.Name)
				wg.Done()
			}()
			if utils.MoveFile(src+"/"+file.Name(), dst+"/"+file.Name()) {

			}
		}()
	}
	wg.Wait()

	log.Printf("All files have been moved from %s to %s\n", src, dst)

}
