package main

import (
	"image"
	//	"log"
)

var frame Frame

const (
	DELTA   = 25
	TMPSIZE = 256
)

func (f *Frame) bxscan(s string, ppt image.Point) (image.Point, image.Point) {
	var (
		w, nb, delta,
		nl, rw int
		b   *Box
		tmp [TMPSIZE + 3]byte
	)
	//	log.Printf("bxscan: s=%s ppt=%s\n", s, ppt)
	frame.r = f.r
	frame.b = f.b
	frame.font = f.font
	frame.maxtab = f.maxtab
	frame.nbox = 0
	frame.nchars = 0

	frame.cols = append([]image.Image{}, f.cols...)
	delta = DELTA
	nl = 0

	for nb = 0; len(s) > 0 && nl <= f.maxlines; nb, frame.nbox = nb+1, frame.nbox+1 {
		if nb == frame.nalloc {
			frame.Grow(delta)
			if delta < 10000 {
				delta *= 2
			}
		}
		b = &frame.box[nb]
		c := s[0]
		if c == '\t' || c == '\n' {
			b.bc = c
			if len(b.ptr) == 0 {
				b.ptr = []byte{c}
			} else {
				b.ptr[0] = c
			}
			b.width = 5000
			if c == '\n' {
				b.minwidth = 0
				nl++
			} else {
				stringwidth(frame.font, " ")
			}
			b.nrune = -1
			frame.nchars++
			s = s[1:]
		} else {
			b.bc = c
			tp := 0 // index into tmp
			nr := 0
			w = 0
			for len(s) > 0 {
				c = s[0]
				if c == '\t' || c == '\n' {
					break
				}
				// TODO: runetochar: runes can be > 1 char
				tmp[tp] = c
				rw = 1
				if tp+rw >= len(tmp) {
					break
				}
				w += f.stringwidth([]byte(s[:1]))
				s = s[1:]
				tp += rw
				nr++
			}
			p := f.allocstr(uint(tp))
			b = &frame.box[nb]
			b.ptr = p
			copy(p, tmp[:tp])
			b.width = w
			b.nrune = nr
			frame.nchars += int64(nr)
		}
	}

	//	log.Printf("bxscan: ppt=%s\n",  ppt)
	ppt = f.LineWrap0(ppt, &frame.box[0])
	//	log.Printf("bxscan: ppt (wrap)=%s\n",  ppt)
	return ppt, (&frame).Draw(ppt)
}
