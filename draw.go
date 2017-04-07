package main

import (
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
)

func (f *Frame) tickat(pt image.Point, ticked bool) {
	//TODO
	if f.ticked == ticked || f.tick == nil || !pt.In(f.r) {
		return
	}
	//pt.X--
	r := image.Rect(pt.X, pt.Y, pt.X+FRTICKW, pt.Y+f.font.height)
	if r.Max.X > f.r.Max.X {
		r.Max.X = f.r.Max.X
	}
	if ticked {
		Draw(f.tickback, f.tickback.Bounds(), f.b, pt)
		Draw(f.b, r, f.tick, image.ZP)
	} else {
		Draw(f.b, r, f.tickback, image.ZP)
	}
	f.ticked = ticked
}

func (f *Frame) Redraw() {
	if f.p0 == f.p1 {
		ticked := f.ticked
		if ticked {
			f.tickat(f.PtOfChar(f.p0), false)
		}
		f.drawsel(f.PtOfChar(0), 0, f.nchars, f.cols[BACK], f.cols[TEXT])
		if ticked {
			f.tickat(f.PtOfChar(f.p0), true)
		}
		return
	}
	pt := f.PtOfChar(0)
	pt = f.drawsel(pt, 0, f.p0, f.cols[BACK], f.cols[TEXT])
	pt = f.drawsel(pt, f.p0, f.p1, f.cols[HIGH], f.cols[HTEXT])
	pt = f.drawsel(pt, f.p1, f.nchars, f.cols[BACK], f.cols[TEXT])
}

func (f *Frame) Draw(pt image.Point) image.Point {
	n := 0
	for nb := 0; nb < f.nbox; nb++ {
		b := &f.box[nb]
		pt = f.LineWrap0(pt, b)
		if pt.Y == f.r.Max.Y {
			f.nchars -= f.strlen(nb)
			f.Delete(nb, f.nbox-1)
			break
		}

		if b.nrune > 0 {
			n = f.CanFit(pt, b)
			if n == 0 {
				panic("frame: draw: cant fit shit")
			}
			if n != b.nrune {
				f.Split(nb, n)
				b = &f.box[nb]
			}
			pt.X += b.width
		} else {
			if b.bc == '\n' {
				pt.X = f.r.Min.X
				pt.Y += f.font.height
			} else {
				pt.X += f.NewWid(pt, b)
			}
		}
	}
	return pt
}

func (f *Frame) DrawText(pt image.Point, text, back image.Image) {
	nb := 0
	for ; nb < f.nbox; nb++ {
		b := &f.box[nb]
		pt = f.LineWrap(pt, b)
		//if !f.noredraw && b.nrune >= 0 {
		if b.nrune >= 0 {
			stringbg(f.b, pt, text, image.ZP, f.font, b.ptr, back, image.ZP)
		}
		pt.X += b.width
	}
}

func (f *Frame) Drawsel(pt image.Point, p0, p1 int64, issel bool) {
	var back, text image.Image
	if f.ticked {
		f.tickat(f.PtOfChar(f.p0), false)
	}

	if p0 == p1 {
		f.tickat(pt, issel)
		return
	}

	if issel {
		back = f.cols[HIGH]
		text = f.cols[HTEXT]
	} else {
		back = f.cols[BACK]
		text = f.cols[TEXT]
	}

	f.drawsel(pt, p0, p1, back, text)
}

func (f *Frame) drawsel(pt image.Point, p0, p1 int64, back, text image.Image) image.Point {
	p := int64(0)
	nr := p
	w := 0
	trim := false
	qt := image.ZP
	var b *Box
	nb := 0
	x := 0
	var ptr []byte
	for ; nb < f.nbox && p < p1; nb++ {
		b = &f.box[nb]
		nr = int64(b.nrune)
		if nr < 0 {
			nr = 1
		}
		if p+nr <= p0 {
			goto Continue
		}
		if p >= p0 {
			qt = pt
			pt = f.LineWrap(pt, b)
			// fill in the end of a wrapped line
			if pt.Y > qt.Y {
				//	cache = append(cache, image.Rect(qt.X, qt.Y, f.r.Max.X, pt.Y))
				Draw(f.b, image.Rect(qt.X, qt.Y, f.r.Max.X, pt.Y), back, qt)
			}
		}
		ptr = b.ptr
		if p < p0 {
			ptr = ptr[p0-p:] // todo: runes
			nr -= p0 - p
			p = p0
		}

		trim = false
		if p+nr > p1 {
			nr -= (p + nr) - p1
			trim = true
		}

		if b.nrune < 0 || nr == int64(b.nrune) {
			w = b.width
		} else {
			// TODO: put stringwidth back
			w = f.stringwidth(ptr[:nr])
		}
		x = pt.X + w
		if x > f.r.Max.X {
			x = f.r.Max.X
		}
		//cache = append(cache, image.Rect(pt.X, pt.Y, x, pt.Y+f.font.height))
		Draw(f.b, image.Rect(pt.X, pt.Y, x, pt.Y+f.font.height), back, pt)
		if b.nrune >= 0 {
			//TODO: must be stringnbg....
			stringbg(f.b, pt, text, image.ZP, f.font, ptr[:nr], back, image.ZP)
		}
		pt.X += w
	Continue:
		b = &f.box[nb+1]
		p += nr
	}

	if p1 > p0 && nb != 0 && nb != f.nbox && (&f.box[nb-1]).nrune > 0 && !trim {
		qt = pt
		pt = f.LineWrap(pt, b)
		if pt.Y > qt.Y {
			//cache =append(cache, image.Rect(qt.X, qt.Y, f.r.Max.X, pt.Y))
			Draw(f.b, image.Rect(qt.X, qt.Y, f.r.Max.X, pt.Y), back, qt)
		}
	}
	return pt
}

var Rainbow = color.RGBA{255, 0, 0, 255}

func next() {
	Rainbow = nextcolor(Rainbow)
}

// nextcolor steps through a gradient
func nextcolor(c color.RGBA) color.RGBA {
	switch {
	case c.R == 255 && c.G == 0 && c.B == 0:
		c.G += 25
	case c.R == 255 && c.G != 255 && c.B == 0:
		c.G += 25
	case c.G == 255 && c.R != 0:
		c.R -= 25
	case c.R == 0 && c.B != 255:
		c.B += 25
	case c.B == 255 && c.G != 0:
		c.G -= 25
	case c.G == 0 && c.R != 255:
		c.R += 25
	default:
		c.B -= 25
	}
	return c
}

func stringbg(dst draw.Image, p image.Point, src image.Image,
	sp image.Point, font Font, s []byte, bg image.Image, bgp image.Point) int {
	h := font.height
	h = int(float64(h) - float64(h)/float64(5))
	for _, v := range s {
		fp := fixed.P(p.X, p.Y)
		dr, mask, maskp, advance, ok := font.Glyph(fp, rune(v))
		if !ok {
			break
		}
		dr.Min.Y += h
		dr.Max.Y += h
		//src = image.NewUniform(Rainbow)
		draw.Draw(dst, dr, bg, bgp, draw.Src)
		draw.DrawMask(dst, dr, src, sp, mask, maskp, draw.Over)
		//next()
		p.X += int(advance >> 6)
	}
	return int(p.X)
}
