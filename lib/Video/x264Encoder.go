package Video

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"github.com/gen2brain/x264-go"
	"image"
	"io"

	"github.com/gen2brain/x264-go/x264c"
)

// Logging constants.
const (
	LogNone int32 = iota - 1
	LogError
	LogWarning
	LogInfo
	LogDebug
)

// Options represent encoding options.
type Options struct {
	// Frame width.
	Width int
	// Frame height.
	Height int
	// Frame rate.
	FrameRate int
	// Tunings: film, animation, grain, stillimage, psnr, ssim, fastdecode, zerolatency.
	Tune string
	// Presets: ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow, placebo.
	Preset string
	// Profiles: baseline, main, high, high10, high422, high444.
	Profile string
	// Log level.
	LogLevel int32
	// bit rate
	BitRate int32
}

// Encoder type.
type Encoder struct {
	e *x264c.T
	w io.Writer

	img  *x264.YCbCr
	opts *Options

	csp int32
	pts int64

	nnals int32
	nals  []*x264c.Nal

	picIn x264c.Picture

	tpf int64
}

// NewX264Encoder returns new x264 encoder.
func NewX264Encoder(w io.Writer, opts *Options) (e *Encoder, err error) {
	e = &Encoder{}

	e.w = w
	e.pts = 0
	e.opts = opts

	e.csp = x264c.CspI420

	e.nals = make([]*x264c.Nal, 3)
	e.img = x264.NewYCbCr(image.Rect(0, 0, e.opts.Width, e.opts.Height))

	param := x264c.Param{}

	if e.opts.Preset != "" && e.opts.Profile != "" {
		ret := x264c.ParamDefaultPreset(&param, e.opts.Preset, e.opts.Tune)
		if ret < 0 {
			err = fmt.Errorf("x264: invalid preset/tune name")
			return
		}
	} else {
		x264c.ParamDefault(&param)
	}

	param.IWidth = int32(e.opts.Width)
	param.IHeight = int32(e.opts.Height)
	param.ICsp = e.csp
	param.ILogLevel = e.opts.LogLevel
	param.IBitdepth = 8
	param.BSlicedThreads = 4
	param.ISyncLookahead = 0
	param.Rc.IVbvMaxBitrate = opts.BitRate / 1000
	param.Rc.IVbvBufferSize = (opts.BitRate / int32(opts.FrameRate)) / 1000
	param.BVfrInput = 0
	param.BRepeatHeaders = 1
	param.BAnnexb = 1
	param.IBframe = 0
	param.BIntraRefresh = 1
	param.IKeyintMax = int32(e.opts.FrameRate)
	param.IFpsNum = uint32(e.opts.FrameRate)
	param.IFpsDen = 1

	if e.opts.Profile != "" {
		ret := x264c.ParamApplyProfile(&param, e.opts.Profile)
		if ret < 0 {
			err = fmt.Errorf("x264: invalid profile name")
			return
		}
	}

	var picIn x264c.Picture
	x264c.PictureInit(&picIn)
	e.picIn = picIn
	e.e = x264c.EncoderOpen(&param)
	if e.e == nil {
		err = fmt.Errorf("x264: cannot open the encoder")
		return
	}

	ret := x264c.EncoderHeaders(e.e, e.nals, &e.nnals)
	if ret < 0 {
		err = fmt.Errorf("x264: cannot encode headers")
		return
	}

	if ret > 0 {
		b := C.GoBytes(e.nals[0].PPayload, C.int(ret))
		n, er := e.w.Write(b)
		if er != nil {
			err = er
			return
		}

		if int(ret) != n {
			err = fmt.Errorf("x264: error writing headers, size=%d, n=%d", ret, n)
		}
	}

	return
}

// Encode encodes image.
func (e *Encoder) Encode(im image.Image) (err error) {
	var picOut x264c.Picture

	_, rgba := im.(*image.RGBA)
	if rgba {
		e.img.ToYCbCr(im)
	} else {
		e.img.ToYCbCrDraw(im)
	}

	picIn := e.picIn

	picIn.Img.ICsp = e.csp

	picIn.Img.IPlane = 3
	picIn.Img.IStride[0] = int32(e.opts.Width)
	picIn.Img.IStride[1] = int32(e.opts.Width) / 2
	picIn.Img.IStride[2] = int32(e.opts.Width) / 2

	picIn.Img.Plane[0] = C.CBytes(e.img.Y)
	picIn.Img.Plane[1] = C.CBytes(e.img.Cb)
	picIn.Img.Plane[2] = C.CBytes(e.img.Cr)

	picIn.IPts = e.pts
	e.pts++

	defer func() {
		picIn.FreePlane(0)
		picIn.FreePlane(1)
		picIn.FreePlane(2)
	}()

	ret := x264c.EncoderEncode(e.e, e.nals, &e.nnals, &picIn, &picOut)
	if ret < 0 {
		err = fmt.Errorf("x264: cannot encode picture")
		return
	}

	if ret > 0 {
		b := C.GoBytes(e.nals[0].PPayload, C.int(ret))

		n, er := e.w.Write(b)
		if er != nil {
			err = er
			return
		}

		if int(ret) != n {
			err = fmt.Errorf("x264: error writing payload, size=%d, n=%d", ret, n)
		}
	}

	return
}

// Flush flushes encoder.
func (e *Encoder) Flush() (err error) {
	var picOut x264c.Picture

	for x264c.EncoderDelayedFrames(e.e) > 0 {
		ret := x264c.EncoderEncode(e.e, e.nals, &e.nnals, nil, &picOut)
		if ret < 0 {
			err = fmt.Errorf("x264: cannot encode picture")
			return
		}

		if ret > 0 {
			b := C.GoBytes(e.nals[0].PPayload, C.int(ret))

			n, er := e.w.Write(b)
			if er != nil {
				err = er
				return
			}

			if int(ret) != n {
				err = fmt.Errorf("x264: error writing payload, size=%d, n=%d", ret, n)
			}
		}
	}

	return
}

// Close closes encoder.
func (e *Encoder) Close() error {
	picIn := e.picIn
	x264c.PictureClean(&picIn)
	x264c.EncoderClose(e.e)
	return nil
}
