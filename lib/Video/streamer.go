package Video

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"github.com/gen2brain/x264-go"
	"github.com/kbinani/screenshot"
	"image"
	"image/draw"
	"os"
	"os/signal"
	"remote_linux/lib/WebRTC"
	"syscall"
	"time"
)

var cursor_serial int = 0

type Cursor struct {
	RelativeX      float32
	RelativeY      float32
	Visible        bool
	Width          int
	Height         int
	ImgHeader      string
	Base64Img      string
	PngImageStream []byte
}

func (cursor *Cursor) Write(data []byte) (n int, err error) {
	cursor.PngImageStream = append(cursor.PngImageStream, data...)
	return 0, nil
}
func (cursor *Cursor) Close() {
	cursor.ImgHeader = "data:image/png;base64,"
	cursor.Base64Img = b64.StdEncoding.EncodeToString(cursor.PngImageStream)
}

func Stream(webRtcEndpoint *WebRTC.WebRtcEndpoint, ctx context.Context, dx int, dy int, dwidth int, dheight int) {

	<-ctx.Done()

	Init()

	opts := Options{
		Width:     dwidth,
		Height:    dheight,
		FrameRate: 30,
		Tune:      "zerolatency",
		Preset:    "ultrafast",
		Profile:   "baseline",
		LogLevel:  x264.LogDebug,
		BitRate:   100_000,
	}
	enc, err := NewEncoder(webRtcEndpoint, &opts)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	defer enc.Close()
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Second / time.Duration(opts.FrameRate))

	//start := time.Now()
	frame := 0

	for range ticker.C {
		select {
		case <-s:
			enc.Flush()
			os.Exit(0)
		default:
			frame++
			img, err := screenshot.Capture(dx, dy, dwidth, dheight)

			x, y, data, width, height, _ := Get()
			upLeft := image.Point{0, 0}
			lowRight := image.Point{width, height}
			i := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})
			i.Pix = data

			if x > dx && x < dx+dwidth && y > dy && y < dy+dheight {
				offset := image.Pt(x-dx, y-dy)
				b := img.Bounds()
				final := image.NewRGBA(b)
				draw.Draw(final, b, img, image.ZP, draw.Src)
				draw.Draw(final, i.Bounds().Add(offset), i, image.ZP, draw.Over)

				img = final
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				continue
			}
			err = enc.Encode(img)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			}
		}

	}
}
