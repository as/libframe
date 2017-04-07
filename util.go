package main

import (
	"image"
	"image/draw"
	//	 "fmt"
)

var cache []image.Rectangle

func init() {
	cache = make([]image.Rectangle, 0, 1024)
}
func flushcache() {
	cache = cache[:0]
}

func Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
	cache = append(cache, r)
	draw.Draw(dst, r, src, sp, draw.Src)
}

func (f *Frame) strlen(nb int) int64 {
	n := int64(0)
	for ; nb < f.nbox; nb++ {
		n += int64(NRUNE(&f.box[nb]))
	}
	return n
}

func (f *Frame) Clean(pt image.Point, n0, n1 int) {
	var b *Box
	c := f.r.Max.X
	nb := 0
	for ; nb < n1-1; nb++ {
		b = &f.box[nb]
		b1 := &f.box[nb+1]
		pt = f.LineWrap(pt, b)
		for b.nrune >= 0 && nb < n1-1 && b1.nrune >= 0 && pt.X+b.width+b1.width < c {
			f.Merge(nb)
			n1--
			b = &f.box[nb]
		}
		pt = f.Advance(pt, &f.box[nb])
	}

	for ; nb < f.nbox; nb++ {
		b = &f.box[nb]
		pt = f.LineWrap(pt, b)
		pt = f.Advance(pt, &f.box[nb])
	}
	f.lastlinefull = 0
	if pt.Y >= f.r.Max.Y {
		f.lastlinefull = 1
	}
}
