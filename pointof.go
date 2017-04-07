package main

import (
	"image"
)

func (f *Frame) PtOfCharPtBox(p int64, pt image.Point, bn int) (x image.Point) {
	var (
		b    *Box
		l, w int
		//r rune
	)
	for ; bn < f.nbox; bn++ {
		b = &f.box[bn]
		pt = f.LineWrap(pt, b)
		l = NRUNE(b)
		if p < int64(l) {
			if b.nrune > 0 {
				for s := b.ptr; p > 0; p, s = p-1, s[w:] {
					// TODO: runes
					w = 1
					pt.X += stringwidth(f.font, string(s[:1]))
					if pt.X > f.r.Max.X {
						panic("PtOfCharPtBox")
					}
				}
			}
			break
		}
		p -= int64(l)
		pt = f.Advance(pt, b)
	}
	return pt
}
func (f *Frame) PtOfCharNBox(p int64, nb int) (pt image.Point) {
	nbox := f.nbox
	f.nbox = nb
	pt = f.PtOfCharPtBox(p, f.r.Min, 0)
	f.nbox = nbox
	return pt
}

func (f *Frame) PtOfChar(p int64) image.Point {
	return f.PtOfCharPtBox(p, f.r.Min, 0)

}
func (f *Frame) Grid(pt image.Point) image.Point {
	pt.Y -= f.r.Min.Y
	pt.Y -= pt.Y % f.font.height
	pt.Y += f.r.Min.Y
	if pt.X > f.r.Max.X {
		pt.X = f.r.Max.X
	}
	return pt
}
func (f *Frame) IndexOf(pt image.Point) int64 {
	pt = f.Grid(pt)
	qt := f.r.Min
	p := int64(0)
	bn := 0
	for ; bn < f.nbox && qt.Y < pt.Y; bn++ {
		b := &f.box[bn]
		qt = f.LineWrap(qt, b)
		if qt.Y >= pt.Y {
			break
		}
		qt = f.Advance(qt, b)
		p += int64(NRUNE(b))
	}

	for ; bn < f.nbox && qt.X <= pt.X; bn++ {
		b := &f.box[bn]
		qt = f.LineWrap(qt, b)
		if qt.Y > pt.Y {
			break
		}
		if qt.X+b.width > pt.X {
			if b.nrune < 0 {
				qt = f.Advance(qt, b)
			} else {
				s := b.ptr
				for {
					r := s[0]
					//TODO: rune
					w := 1
					if r == 0 {
						//println("calm panic: nul in string")
					}
					qt.X += stringwidth(f.font, string(s[:1]))
					s = s[w:]
					if qt.X > pt.X {
						break
					}
					p++
				}
			}
		} else {
			p += int64(NRUNE(b))
			qt = f.Advance(qt, b)
		}
	}
	return p
}
