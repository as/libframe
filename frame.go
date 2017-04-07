package main

import (
	"image"
	"image/draw"

	"golang.org/x/image/font"
)

const (
	TEXT  = 0
	BACK  = 1
	HTEXT = 2
	HIGH  = 3
)

type Font struct {
	font.Face
	height int
}

type Frame struct {
	font         Font
	disp         draw.Image
	b            draw.Image
	cols         []image.Image
	r            image.Rectangle
	entire       image.Rectangle
	maxtab       int
	nbox         int
	nalloc       int
	nchars       int64
	nlines       int
	p0           int64
	p1           int64
	box          []Box
	lastlinefull int
	tick         draw.Image
	tickback     draw.Image
	ticked       bool
	tickscale    int
	maxlines     int
	modified     bool
	noredraw     bool
}

func NewFrame(r image.Rectangle, ft Font, b draw.Image, cols []image.Image) *Frame {
	f := &Frame{
		font:   ft,
		maxtab: 8 * stringwidth(ft, "0"),
		cols:   append([]image.Image{}, cols...),
	}
	f.setrects(r, b)
	f.inittick()
	return f
}

const FRTICKW = 4

func (f *Frame) inittick() {
	f.tickscale = 1 // TODO implement scalesize
	f.tick = image.NewRGBA(image.Rect(0, 0, FRTICKW, f.font.height))
	f.tickback = image.NewRGBA(image.Rect(0, 0, FRTICKW, f.font.height))
	Draw(f.tick, f.tick.Bounds(), f.cols[TEXT], image.ZP)
	Draw(f.tick, f.tick.Bounds().Inset(1), f.cols[HIGH], image.ZP)
}

func (f *Frame) setrects(r image.Rectangle, b draw.Image) {
	f.b = b
	f.entire = r
	f.r = r
	f.r.Max.Y -= f.r.Dy() % f.font.height
	f.maxlines = f.r.Dy() / f.font.height
}

func (f *Frame) measure(p []byte) int {
	return int(font.MeasureBytes(f.font.Face, p) >> 6)
}

func (f *Frame) clear(freeall bool) {
	if f.nbox != 0 {
		f.Delete(0, f.nbox-1)
	}
	if f.box != nil {
		free(f.box)
	}
	if freeall {
		// TODO: unnecessary
		freeimage(f.tick)
		freeimage(f.tickback)
		f.tick = nil
		f.tickback = nil
	}
	f.box = nil
	f.ticked = false

}

func free(i interface{}) {
}
func freeimage(i image.Image) {
}

func stringwidth(ft Font, s string) int {
	return int(font.MeasureString(ft.Face, s) >> 6)
}

func (f *Frame) ChopFrame(pt image.Point, p int64, bn int) {
	bn, nbox := 0, f.nbox
	for ; bn < f.nbox; bn++ {
		b := &f.box[bn]
		pt := f.LineWrap(pt, b)
		if pt.Y >= f.r.Max.Y {
			break
		}
		p += int64(NRUNE(b))
		pt = f.Advance(pt, b)
	}
	f.nchars = p
	f.nlines = f.maxlines
	if bn < nbox {
		f.Delete(bn, f.nbox-1)
	}
}
