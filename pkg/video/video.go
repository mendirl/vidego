package video

import (
	vidio "github.com/AlexEidt/Vidio"
	"log"
	"os"
	"strings"
	"vidego/pkg/datatype"
	"vidego/pkg/utils"
)

func computeDuration(path string) (float64, error) {
	video, err := vidio.NewVideo(path)

	return video.Duration(), err
}

func CreateVideo(path string) datatype.Video {
	defer utils.HandlePanic(path)

	log.Printf("#video path : %s \n", path)
	//defer HandlePanic(path)

	info, err := os.Stat(path)

	if err != nil {
		log.Printf("## ERROR with Stat : %s \n", err)
	} else {
		duration, err := computeDuration(path)
		if err != nil {
			log.Printf("#ERROR with vidio.NewVideo and file %s: %s \n", path, err)
		}

		split := strings.Split(path, "/")
		name := split[len(split)-1]
		sourcePath := utils.TrimSuffix(path, "/"+name)

		return datatype.Video{Name: name, Path: sourcePath, Size: info.Size(), Duration: duration, Complete: duration == 0}
	}

	return datatype.Video{Name: "empty", Path: path}
}
