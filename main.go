package main

import (
	"sync"
	//	"github.com/as/clip"
	//
	//	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var wg sync.WaitGroup
var winSize = image.Pt(1920, 1080)

var (
	Red    = image.NewUniform(color.RGBA{255, 0, 0, 255})
	Green  = image.NewUniform(color.RGBA{0, 255, 0, 255})
	Blue   = image.NewUniform(color.RGBA{0, 192, 192, 255})
	Cyan   = image.NewUniform(color.RGBA{0xAA, 0xAA, 0xFF, 255})
	White  = image.NewUniform(color.RGBA{255, 255, 255, 255})
	Yellow = image.NewUniform(color.RGBA{255, 255, 224, 255})
	Gray   = image.NewUniform(color.RGBA{66, 66, 66, 255})
	Mauve  = image.NewUniform(color.RGBA{128, 66, 193, 255})
)

func mkfont(size int) Font {
	f, err := truetype.Parse(gomono.TTF)
	if err != nil {
		panic(err)
	}
	return Font{
		Face: truetype.NewFace(f, &truetype.Options{
			Size: float64(size),
		}),
		height: size,
	}
}

func file(s string) []byte {
	p, err := ioutil.ReadFile(s)
	if err != nil {
		log.Println(err)
	}
	return p
}
func main() {
	driver.Main(func(src screen.Screen) {
		win, _ := src.NewWindow(&screen.NewWindowOptions{winSize.X, winSize.Y})
		focused := false
		focused = focused

		buf, err := src.NewBuffer(winSize)
		if err != nil {
			panic(err)
		}
		draw.Draw(buf.RGBA(), buf.RGBA().Bounds(), Yellow, image.ZP, draw.Src)
		tx, err := src.NewTexture(winSize)
		if err != nil {
			panic(err)
		}
		cols := []image.Image{
			Gray,   // Text
			Yellow, // Back
			Gray,   // HTEXT
			Cyan,   // HIGH
			Green,  // ????
		}
		ft := mkfont(40)
		r := image.Rectangle{image.ZP, winSize}
		fr := NewFrame(r, ft, buf.RGBA(), cols)
		if len(os.Args) > 1 {
			fr.Insert(string(file(strings.Join(os.Args[1:], " "))), fr.p0)
		}
		paintfn := func() {
			tx.Upload(image.ZP, buf, buf.Bounds())
			win.Copy(buf.Bounds().Min, tx, tx.Bounds(), screen.Src, nil)
			win.Publish()
		}

		// lambda to paint only rectangles changed during a sweep of the mouse
		paintcache := func() {
			wg.Add(len(cache))
			for _, r := range cache {
				go func(r image.Rectangle) {
					tx.Upload(r.Min, buf, r)
					wg.Done()
				}(r)
			}
			wg.Wait()
			wg.Add(len(cache))
			for _, r := range cache {
				go func(r image.Rectangle) {
					pt := r.Min
					pt.X += r.Max.X - (r.Max.X - r.Min.X)
					pt.Y += r.Max.Y - (r.Max.Y - r.Min.Y)
					win.Copy(pt, tx, r, screen.Src, nil)
					wg.Done()
				}(r)
			}
			wg.Wait()
			win.Publish()
			flushcache()

		}
		var drawdot bool
		for {
			switch e := win.NextEvent().(type) {
			case mouse.Event:
				pt := image.Pt(int(e.X), int(e.Y))
				if e.Button == 2 || drawdot {
					if e.Direction == mouse.DirRelease {
						drawdot = false
					} else {
						drawdot = true
						Draw(fr.b, image.Rect(-5, -2, 5, 2).Add(pt), Mauve, image.ZP)
						win.Send(paint.Event{})
					}
				}
				if e.Button == 1 {
					//				//	i := fr.IndexOf(pt)
					switch e.Direction {
					case mouse.DirPress:
						fr.Select(pt, win, paintcache)
						win.Send(paint.Event{})
					case mouse.DirRelease:
						//							fr.p1 = i
						//							fr.Drawsel(pt, fr.p0, fr.p1, true)
					}
				}
			case key.Event:
				if e.Direction == key.DirRelease {
					continue
				}
				if e.Rune == 'M' {
					fr.Insert("Mink", fr.p0)
					fr.p1 += int64(len("Mink"))
					win.Send(paint.Event{})
					continue
				}
				if e.Rune == '\r' {
					e.Rune = '\n'
				}
				if e.Code == key.CodeLeftArrow {
					fr.p0--
					fr.p1--
					win.Send(paint.Event{})
					continue
				}

				if e.Code == key.CodeRightArrow {
					fr.p0++
					fr.p1++
					win.Send(paint.Event{})
					continue
				}
				if e.Rune == '\x08' {
					if fr.p1 == fr.p0 {
						fr.p0--
					}
					fr.DeleteBytes(fr.p0, fr.p1)
					win.Send(paint.Event{})
					fr.p0 = fr.p1
					continue
				}

				if e.Rune == -1 {
					continue
				}
				fr.Insert(string(e.Rune), fr.p1)
				fr.p0 = fr.p1
				win.SendFirst(paint.Event{})
			case size.Event:
				fr.Redraw()
				paintfn()
				flushcache()
			case paint.Event:
				paintfn()
				flushcache()
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
				// NT doesn't repaint the window if another window covers it
				if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOff {
					focused = false
				} else if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOn {
					focused = true
				}
			}
		}
	})
}

func drawBorder(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point, thick int) {
	draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+thick), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Min.X, r.Max.Y-thick, r.Max.X, r.Max.Y), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Min.X+thick, r.Max.Y), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Max.X-thick, r.Min.Y, r.Max.X, r.Max.Y), src, sp, draw.Src)
}
