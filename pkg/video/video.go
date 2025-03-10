package video

import (
	"fmt"
	vidio "github.com/AlexEidt/Vidio"
)

func ComputeDuration(path string) (float64, error) {
	video, err := vidio.NewVideo(path)

	return video.Duration(), err
}
