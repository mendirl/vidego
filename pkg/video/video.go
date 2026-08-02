package video

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"vidego/pkg/datatype"

	vidio "github.com/AlexEidt/Vidio"
)

func computeDuration(path string) (float64, error) {
	video, err := vidio.NewVideo(path)
	if err != nil {
		log.Printf("#ERROR with vidio.NewVideo and file %s: %s \n", path, err)
		return 0, err
	}

	return video.Duration(), nil
}

func CreateVideo(path string) (v datatype.Video, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("## something is panicking with file %s : %s\n", path, r)
			err = fmt.Errorf("panic: %v", r)
			v = datatype.Video{Name: filepath.Base(path), Path: filepath.ToSlash(filepath.Dir(path)), Duration: 0, Complete: false}
		}
	}()

	log.Printf("#create Video with path : %s \n", path)

	info, statErr := os.Stat(path)

	if statErr != nil {
		log.Printf("## ERROR with Stat : %s \n", statErr)
		return datatype.Video{Name: filepath.Base(path), Path: filepath.ToSlash(filepath.Dir(path))}, statErr
	} else {
		duration, durErr := computeDuration(path)
		if durErr != nil {
			return datatype.Video{Name: filepath.Base(path), Path: filepath.ToSlash(filepath.Dir(path)), Size: info.Size(), Duration: 0, Complete: false}, durErr
		}

		name := filepath.Base(path)
		sourcePath := filepath.ToSlash(filepath.Dir(path))

		var complete = duration == 0
		if strings.Contains(sourcePath, "ALL") {
			complete = true
		}

		return datatype.Video{Name: name, Path: sourcePath, Size: info.Size(), Duration: duration, Complete: complete}, nil
	}
}
