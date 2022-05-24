package Video

import (
	"fmt"
	"github.com/kbinani/screenshot"
	"image"
	"image/draw"
	"os"
	"time"
)

type DesktopCapture struct {
	x            int
	y            int
	width        int
	height       int
	outputStream chan *image.RGBA
	ticker       *time.Ticker
}

func (capture DesktopCapture) Start() {
	go func() {
		for range capture.ticker.C {
			img, err := screenshot.Capture(capture.x, capture.y, capture.width, capture.height)
			x, y, data, width, height, _ := CursorGet()
			upLeft := image.Point{0, 0}
			lowRight := image.Point{width, height}
			i := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})
			i.Pix = data

			if x > capture.x && x < capture.x+capture.width && y > capture.y && y < capture.y+capture.height {
				offset := image.Pt(x-capture.x, y-capture.y)
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
			capture.outputStream <- img
		}

	}()
}

func NewCapture(x int, y int, width int, height int, frameRate int, outputStream chan *image.RGBA) DesktopCapture {
	capture := DesktopCapture{
		x:            x,
		y:            y,
		width:        width,
		height:       height,
		outputStream: outputStream,
		ticker:       time.NewTicker(time.Second / time.Duration(frameRate)),
	}

	CursorInit()

	return capture
}
