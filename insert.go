package main

import (
	"image"
	//	"fmt"
)

type Pts struct {
	pt0, pt1 image.Point
}

var (
	pts    []Pts
	nalloc int
)

func (f *Frame) Insert(s string, p0 int64) {
	var (
		pt0, pt1,
		ppt0, ppt1,
		opt0,
		pt image.Point

		b             *Box
		n, n0, nn0, y int
		cn0           int64
		back, text    image.Image

		r image.Rectangle
	)
	var (
		npts = 0
	)

	if p0 > f.nchars || len(s) == 0 || f.b == nil {
		return
	}

	// find p0, it's box, and its point in the box its in
	n0 = f.Find(0, 0, p0)
	cn0 = p0
	nn0 = n0
	pt0 = f.PtOfCharNBox(p0, n0)
	//	fmt.Printf("insert: f.PtOfCharNBox <- %d: %s\n", p0, pt0)
	ppt0 = pt0
	opt0 = pt0

	// find p1
	ppt0, pt1 = f.bxscan(s, ppt0)
	ppt1 = pt1
	//	fmt.Printf("insert: pt0, pt1: %s, %s\n", pt0, pt1)
	//	fmt.Printf("insert: ppt0, ppt1: %s, %s\n", ppt0, ppt1)
	// Line wrap
	if n0 < f.nbox {
		b = &f.box[n0]
		pt0 = f.LineWrap(pt0, b)
		ppt1 = f.LineWrap0(ppt1, b)
	}
	f.modified = true

	// pt0, pt1   - start and end of insertion (current; and without line wrap)
	// ppt0, ppt1 - start and end of insertion when its complete

	if f.p0 == f.p1 {
		f.tickat(f.PtOfChar(int64(f.p0)), false)
	}

	// Find the points where all the old x and new x line up
	// Invariants:
	//   pt0 is where the next box (b, n0) is now
	//   pt1 is where it will be after intsertion
	// If pt1 goes off the rect, toss everything from there on
	npts = 0
	if n0 < f.nbox {
		b = &f.box[n0]
	}
	for ; pt1.X != pt0.X && pt1.Y != f.r.Max.Y && n0 < f.nbox; n0, npts = n0+1, npts+1 {
		b = &f.box[n0]
		pt0 = f.LineWrap(pt0, b)
		pt1 = f.LineWrap0(pt1, b)

		if b.nrune > 0 {
			n = f.CanFit(pt1, b)
			//				fmt.Println("i can fit %d runes in this box from pt %s\n", n, pt1)
			if n == 0 {
				panic("f. ==0")
			}
			if n != b.nrune {
				f.Split(n0, n)
				b = &f.box[n0]
			}
		}

		if npts == nalloc {
			pts = append(pts, make([]Pts, DELTA)...)
			nalloc += DELTA
			b = &f.box[n0]
		}
		pts[npts].pt0 = pt0
		pts[npts].pt1 = pt1
		// check for text overflow off the frame
		if pt1.Y == f.r.Max.Y {
			break
		}
		pt0 = f.Advance(pt0, b)
		pt1.X += f.NewWid(pt1, b)
		cn0 += int64(NRUNE(b))
	}

	if pt1.Y > f.r.Max.Y {
		panic("Frame.Insert pt1 too far")
	}
	if pt1.Y == f.r.Max.Y && n0 < f.nbox {
		f.nchars -= f.strlen(n0)
		f.Delete(n0, f.nbox-1)
	}

	h := f.font.height
	if n0 == f.nbox {
		extra := 0
		if pt1.X > f.r.Min.X {
			extra = 1
		}
		f.nlines = (pt.Y-f.r.Min.Y)/h + extra
	} else if pt1.Y != pt0.Y {
		y = f.r.Max.Y
		q0 := pt0.Y + h
		q1 := pt1.Y + h
		f.nlines += (q1 - q0) / h
		if f.nlines > f.maxlines {
			//TODO: Theres a name collision with chop for the box version of chop
			f.ChopFrame(ppt1, p0, nn0)
		}
		if pt1.Y < y {
			r = f.r
			r.Min.Y = q1
			r.Max.Y = y
			if q1 < y {
				Draw(f.b, r, f.b, image.Pt(f.r.Min.X, q0))
			}
			r.Min = pt1
			r.Max.X = pt1.X + (f.r.Max.X - pt0.X)
			r.Max.Y = q1
			Draw(f.b, r, f.b, pt0)
		}
	}

	// Move the old stuff down to make rooms
	if pt1.Y == f.r.Max.Y {
		y = pt1.Y
	} else {
		y = 0
	}

	npts--
	for ctr := n0; npts >= 0; npts-- {
		ctr--
		b = &f.box[ctr]
		pt = pts[npts].pt1
		//		fmt.Printf("npts=%d selected point = %s\n", npts, pt)
		if b.nrune > 0 {
			r.Min = pt
			r.Max = r.Min
			r.Max.X += b.width
			r.Max.Y += f.font.height
			Draw(f.b, r, f.b, pts[npts].pt0)

			// clear bit hanging off right
			if npts == 0 && pt.Y > pt0.Y {
				// first new char bigger than first char displaced
				// so line wrap happened
				r.Min = opt0
				r.Max = opt0
				r.Max.X = f.r.Max.X
				r.Max.Y += f.font.height
				if f.p0 <= cn0 && cn0 < f.p1 { // b+1 is in selection
					back = f.cols[HIGH]
				} else {
					back = f.cols[BACK]
				}
				Draw(f.b, r, back, r.Min)
			} else if pt.Y < y {
				r.Min = pt
				r.Max = pt
				r.Min.X += b.width
				r.Max.Y += f.font.height
				if f.p0 <= cn0 && cn0 < f.p1 {
					back = f.cols[HIGH]
				} else {
					back = f.cols[BACK]
				}
				Draw(f.b, r, back, r.Min)
			}
			y = pt.Y
			cn0 -= int64(b.nrune)
		} else {
			r.Min = pt
			r.Max = pt
			r.Max.X += b.width
			r.Max.Y += f.font.height
			if r.Max.X >= f.r.Max.X {
				r.Max.X = f.r.Max.X
			}
			cn0--
			if f.p0 <= cn0 && cn0 < f.p1 { // box inside selection
				back = f.cols[HIGH]
			} else {
				back = f.cols[BACK]
			}
			Draw(f.b, r, back, r.Min)
			y = 0
			if pt.X == f.r.Min.X {
				y = pt.Y
			}
		}
	}

	// insertion can extend the selection; different condition
	if f.p0 < p0 && p0 <= f.p1 {
		text = f.cols[HIGH]
		back = f.cols[HTEXT]
	} else {
		text = f.cols[TEXT]
		back = f.cols[BACK]
	}

	f.SelectPaint(ppt0, ppt1, back)
	(&frame).DrawText(ppt0, text, back)
	f.Add(nn0, frame.nbox)
	for n = 0; n < frame.nbox; n++ {
		f.box[nn0+n] = frame.box[n]
	}
	if nn0 > 0 && f.box[nn0-1].nrune >= 0 && ppt0.X-f.box[nn0-1].width >= f.r.Min.X {
		nn0--
		ppt0.X -= f.box[nn0].width
	}

	n0 += frame.nbox
	if n0 < f.nbox-1 {
		f.Clean(ppt0, nn0, n0+1)
	} else {
		f.Clean(ppt0, nn0, n0)
	}
	f.nchars += frame.nchars
	if f.p0 >= p0 {
		f.p0 += frame.nchars
	}
	if f.p0 > f.nchars {
		f.p0 = f.nchars
	}
	if f.p1 >= p0 {
		f.p1 += frame.nchars
	}
	if f.p1 > f.nchars {
		f.p1 = f.nchars
	}
	if f.p0 == f.p1 {
		f.tickat(f.PtOfChar(f.p0), true)
	}
}
