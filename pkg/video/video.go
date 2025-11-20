package video

import (
	"log"
	"os"
	"strings"
	"vidego/pkg/datatype"
	"vidego/pkg/panic"
	"vidego/pkg/tools"

	vidio "github.com/AlexEidt/Vidio"
)

func computeDuration(path string) float64 {
	video, err := vidio.NewVideo(path)
	if err != nil {
		log.Printf("#ERROR with vidio.NewVideo and file %s: %s \n", path, err)
	}

	return video.Duration()
}

func CreateVideo(path string) datatype.Video {
	defer panic.HandlePanic(path)

	log.Printf("#create Video with path : %s \n", path)
	//defer HandlePanic(path)

	info, err := os.Stat(path)

	if err != nil {
		log.Printf("## ERROR with Stat : %s \n", err)
	} else {
		duration := computeDuration(path)

		split := strings.Split(path, "/")
		name := split[len(split)-1]
		sourcePath := tools.TrimSuffix(path, "/"+name)

		var complete = duration == 0
		if strings.Contains(sourcePath, "ALL") {
			complete = true
		}

		return datatype.Video{Name: name, Path: sourcePath, Size: info.Size(), Duration: duration, Complete: complete}
	}

	return datatype.Video{Name: "empty", Path: path}
}
