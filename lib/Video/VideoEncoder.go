package Video

import (
	"fmt"
	"github.com/gen2brain/x264-go"
	"image"
	"os"
)

type VideoEncoder struct {
	Encoder      *Encoder
	inputStream  chan *image.RGBA
	outputStream chan *[]byte
}

func (encoder VideoEncoder) Write(data []byte) (n int, err error) {
	fmt.Println(encoder.outputStream)
	encoder.outputStream <- &data
	return len(data), nil
}

func (encoder VideoEncoder) Start() {
	go func() {
		for {
			image := <-encoder.inputStream
			encoder.Encoder.Encode(image)
		}

	}()
}

func NewEncoder(width int, height int, frameRate int, bitRate int, inputStream chan *image.RGBA, outputStream chan *[]byte) VideoEncoder {
	encoder := VideoEncoder{
		outputStream: outputStream,
		inputStream:  inputStream,
	}

	opts := Options{
		Width:     width,
		Height:    height,
		FrameRate: frameRate,
		Tune:      "zerolatency",
		Preset:    "ultrafast",
		Profile:   "baseline",
		LogLevel:  x264.LogDebug,
		BitRate:   int32(bitRate),
	}
	fmt.Println("creating x264 encoder")
	enc, err := NewX264Encoder(encoder, &opts)
	encoder.Encoder = enc

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("encoder is ok")

	return encoder
}
