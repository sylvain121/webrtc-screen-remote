package Video

import (
	"context"
	"fmt"
	"github.com/pion/webrtc/v3/pkg/media"
	"image"
	"remote_linux/lib/WebRTC"
	"time"
)

func Stream(webRtcEndpoint *WebRTC.WebRtcEndpoint, ctx context.Context, x int, y int, width int, height int) {
	fmt.Println("waiting done")
	<-ctx.Done()

	fmt.Println("configuring")
	frameRate := 30
	captureChan := make(chan *image.RGBA, 10)
	encodedChan := make(chan *[]byte, 10)
	/**
	Go routine handle video data
	*/

	go func() {
		for {
			data := <-encodedChan
			_ = webRtcEndpoint.VideoTrack.WriteSample(media.Sample{Data: *data, Duration: time.Second / time.Duration(frameRate)})
		}
	}()

	fmt.Println("configuring desktop capture")
	capture := NewCapture(x, y, width, height, frameRate, captureChan)
	fmt.Println("configuring encoder")
	encoder := NewEncoder(width, height, frameRate, 1_000_000, captureChan, encodedChan)
	fmt.Println("Init ok")

	fmt.Println("starting")
	encoder.Start()
	capture.Start()

}
