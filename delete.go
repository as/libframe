package main

import (
	"image"
)

func (f *Frame) DeleteBytes(p0, p1 int64) int {
	var (
		b              *Box
		pt0, pt1, ppt0 image.Point
		n0, n1, n      int
		cn1            int64
		r              image.Rectangle
		nn0            int
		col            image.Image
	)

	if p0 >= f.nchars || p0 == p1 || f.b == nil {
		return 0
	}

	if p1 > f.nchars {
		p1 = f.nchars
	}
	n0 = f.Find(0, 0, p0)
	if n0 == f.nbox {
		panic("delete")
	}
	n1 = f.Find(n0, p0, p1)
	pt0 = f.PtOfCharNBox(p0, n0)
	pt1 = f.PtOfChar(p1)
	if f.p0 == f.p1 {
		f.tickat(f.PtOfChar(int64(f.p0)), false)
	}

	nn0 = n0
	ppt0 = pt0
	f.Free(n0, n1-1)
	f.modified = true

	// pt0, pt1 - beginning, end
	// n0 - has beginning of deletion
	// n1, b - first box kept after deletion
	// cn1 char pos of n1
	//
	// adjust f.p0 and f.p1 after deletion is finished

	if n1 > f.nbox {
		panic("DeleteBytes: Split bug: nul terminators removed")
	}

	b = &f.box[n1]
	cn1 = int64(p1)

	for pt1.X != pt0.X && n1 < f.nbox {
		pt0 = f.LineWrap0(pt0, b)
		pt1 = f.LineWrap(pt1, b)
		r.Min = pt0
		r.Max = pt0
		r.Max.Y += f.font.height

		if b.nrune > 0 { // non-newline
			n = f.CanFit(pt0, b)
			if n == 0 {
				panic("delete: canfit==0")
			}
			if n != b.nrune {
				f.Split(n1, n)
				b = &f.box[n1]
			}
			r.Max.X += b.width
			Draw(f.b, r, f.b, pt1)
			//drawBorder(f.b, r.Add(pt1).Inset(-4), Red, image.ZP, 8)
			//drawBorder(f.b, r.Inset(-4), Green, image.ZP, 8)
			cn1 += int64(b.nrune)
		} else {
			r.Max.X += f.NewWid0(pt0, b)
			if r.Max.X > f.r.Max.X {
				r.Max.X = f.r.Max.X
			}
			col = f.cols[BACK]
			if f.p0 <= cn1 && cn1 < f.p1 {
				col = f.cols[HIGH]
			}
			Draw(f.b, r, col, pt0)
			cn1++
		}
		pt1 = f.Advance(pt1, b)
		pt0.X += f.NewWid(pt0, b)
		f.box[n0] = f.box[n1]
		n0++
		n1++
		b = &f.box[n1]
	}

	if n1 == f.nbox && pt0.X != pt1.X {
		f.SelectPaint(pt0, pt1, f.cols[BACK])
	}

	if pt1.Y != pt0.Y {
		pt2 := f.PtOfCharPtBox(32767, pt1, n1)
		if pt2.Y > f.r.Max.Y {
			panic("delete: PtOfCharPtBox")
		}
		if n1 < f.nbox {
			h := f.font.height
			q0 := pt0.Y + h
			q1 := pt1.Y + h
			q2 := pt2.Y + h
			if q2 > f.r.Max.Y {
				q2 = f.r.Max.Y
			}
			Draw(f.b, image.Rect(pt0.X, pt0.Y, pt0.X+(f.r.Max.X-pt1.X), q0), f.b, pt1)
			Draw(f.b, image.Rect(f.r.Min.X, q0, f.r.Max.X, q0+(q2-q1)), f.b, image.Pt(f.r.Min.X, q1))
			f.SelectPaint(image.Pt(pt2.X, pt2.Y-(pt1.Y-pt0.Y)), pt2, f.cols[BACK])
		} else {
			f.SelectPaint(pt0, pt2, f.cols[BACK])
		}
	}

	f.Close(n0, n1-1)
	if nn0 > 0 && f.box[nn0-1].nrune >= 0 && ppt0.X-f.box[nn0-1].width >= f.r.Min.X {
		nn0--
		ppt0.X -= f.box[nn0].width
	}

	if n0 < f.nbox-1 {
		f.Clean(ppt0, nn0, n0+1)
	} else {
		f.Clean(ppt0, nn0, n0)
	}

	if f.p1 > p1 {
		f.p1 -= p1 - p0
	} else if f.p1 > p0 {
		f.p1 = p0
	}

	if f.p0 > p1 {
		f.p0 -= p1 - p0
	} else if f.p0 > p0 {
		f.p0 = p0
	}

	f.nchars -= p1 - p0
	if f.p0 == f.p1 {
		f.tickat(f.PtOfChar(f.p0), true)
	}
	pt0 = f.PtOfChar(f.nchars)
	n = f.nlines
	extra := 0
	if pt0.X > f.r.Min.X {
		extra = 1
	}
	f.nlines = (pt0.Y-f.r.Min.Y)/f.font.height + extra
	return n - f.nlines
}
