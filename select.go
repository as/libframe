package main

import (
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/mouse"
	"image"
	//		"golang.org/x/mobile/event/paint"
	//	"fmt"
)

func region(a, b int64) int64 {
	if a < b {
		return -1
	}
	if a == b {
		return 0
	}
	return 1
}

func (f *Frame) Select(mp image.Point, ed screen.EventDeque, paintfn func()) {
	f.modified = false
	f.Drawsel(f.PtOfChar(f.p0), f.p0, f.p1, false)
	p1 := f.IndexOf(mp)
	p0 := p1
	pt0 := f.PtOfChar(p0)
	pt1 := f.PtOfChar(p1)
	f.Drawsel(pt0, p0, p1, true)

	reg := int64(0)
	for {
		q := f.IndexOf(mp)
		if p1 != q {
			if reg != region(q, p0) {
				if reg > 0 {
					f.Drawsel(pt0, p0, p1, false)
				} else if reg < 0 {
					f.Drawsel(pt1, p1, p0, false)
				}
				p1 = p0
				pt1 = pt0
				reg = region(q, p0)
				if reg == 0 {
					f.Drawsel(pt0, p0, p1, true)
				}
			}
			qt := f.PtOfChar(q)
			if reg > 0 {
				if q > p1 {
					f.Drawsel(pt1, p1, q, true)
				} else if q < p1 {
					f.Drawsel(qt, q, p1, false)
				}
			} else if reg < 0 {
				if q > p1 {
					f.Drawsel(pt1, p1, q, false)
				} else {
					f.Drawsel(qt, q, p1, true)
				}
			}
			p1 = q
			pt1 = qt
		}
		f.modified = false
		if p0 < p1 {
			f.p0 = p0
			f.p1 = p1
		} else {
			f.p0 = p1
			f.p1 = p0
		}
		e := ed.NextEvent()
		switch e := e.(type) {
		case mouse.Event:
			if e.Button == 1 && e.Direction == 2 {
				ed.SendFirst(e)
				return
			}
			mp = image.Pt(int(e.X), int(e.Y))
			paintfn()
			flushcache()
		case interface{}:
			ed.SendFirst(e)
			return
		}
	}
}

func (f *Frame) SelectPaint(p0, p1 image.Point, col image.Image) {
	if f.b == nil {
		panic("selectpaint: b == 0")
	}
	if f.r.Max.Y == p0.Y {
		return
	}
	h := f.font.height
	q0, q1 := p0, p1
	q0.Y += h
	q1.Y += h
	n := (p1.Y - p0.Y) / h

	if n == 0 { // one line
		Draw(f.b, image.Rectangle{p0, q1}, col, image.ZP)
	} else {
		if p0.X >= f.r.Max.X {
			p0.X = f.r.Max.X - 1
		}
		Draw(f.b, image.Rect(p0.X, p0.Y, f.r.Max.X, q0.Y), col, image.ZP)
		if n > 1 {
			Draw(f.b, image.Rect(f.r.Min.X, q0.Y, f.r.Max.X, p1.Y), col, image.ZP)
		}
		Draw(f.b, image.Rect(f.r.Min.X, p1.Y, q1.X, q1.Y), col, image.ZP)
	}
}
