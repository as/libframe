package main

import (
	"image"
)

// LineWrap checks whether the box would wrap across a line boundary
// if it were inserted at pt. If it wraps, the line-wrapped point is
// returned.
func (f *Frame) LineWrap(pt image.Point, b *Box) image.Point {
	width := b.width
	if b.nrune < 0 {
		width = b.minwidth
	}
	if width > f.r.Max.X-pt.X {
		pt.X = f.r.Min.X
		pt.Y += f.font.height
	}
	return pt
}

// LineWrap0 returns the line-wrapped point if the box doesn't
// fix on the line
func (f *Frame) LineWrap0(pt image.Point, b *Box) image.Point {
	if f.CanFit(pt, b) == 0 {
		pt.X = f.r.Min.X
		pt.Y += f.font.height
	}
	return pt
}

func (f *Frame) CanFit(pt image.Point, b *Box) int {
	left := f.r.Max.X - pt.X
	w := 0
	if b.nrune < 0 {
		if b.minwidth <= left {
			return 1
		}
		return 0
	}
	if left >= b.width {
		return b.nrune
	}
	p := b.ptr
	for nr := 0; len(p) > 0; p, nr = p[w:], nr+1 {
		// TODO: need to measure actual rune width
		// r := p[0]
		w = 1
		left -= stringwidth(f.font, string(p[:1]))
		if left < 0 {
			return nr
		}
	}
	panic("CanFit Can't Fit Shit")
}

func (f *Frame) Advance(pt image.Point, b *Box) (x image.Point) {
	//	pt0 := pt
	//	defer func(){fmt.Printf("Advance: pt=%d -> %d\n",pt0,x)}()
	//	fmt.Println("boxes width: %d", b.width)
	if b.nrune < 0 && b.bc == '\n' {
		pt.X = f.r.Min.X
		pt.Y += f.font.height
	} else {
		pt.X += b.width
	}
	return pt
}

// TODO: Naming
func (f *Frame) NewWid(pt image.Point, b *Box) int {
	b.width = f.NewWid0(pt, b)
	return b.width
}
func (f *Frame) NewWid0(pt image.Point, b *Box) int {
	c := f.r.Max.X
	x := pt.X
	if b.nrune >= 0 || b.bc != '\t' {
		return b.width
	}
	if x+b.minwidth > c {
		pt.X = f.r.Min.X
		x = pt.X
	}
	x += f.maxtab
	x -= (x - f.r.Min.X) % f.maxtab
	if x-pt.X < b.minwidth || x > c {
		x = pt.X + b.minwidth
	}
	return x - pt.X
}
