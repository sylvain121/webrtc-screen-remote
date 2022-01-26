package main

import (
	"flag"
	"remote_linux/lib"
)

func main() {
	height := flag.Int("height", 600, "screen captured height")
	width := flag.Int("width", 800, "screen capture width")
	x := flag.Int("x", 0, "screen capture x offset")
	y := flag.Int("y", 0, "screen capture y offset")

	flag.Parse()

	lib.Init(*x, *y, *width, *height)
}
