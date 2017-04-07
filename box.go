package main

import (
	"fmt"
)

const SLOP = 25

type Box struct {
	bc       byte
	width    int
	minwidth int
	nrune    int
	ptr      []byte
}

// Add adds n boxes after box bn, the rest are shifted up
func (f *Frame) Add(bn, n int) {
	if bn > f.nbox {
		panic("Frame.Add")
	}
	if f.nbox+n > f.nalloc {
		f.Grow(n + SLOP)
	}
	for i := f.nbox - 1; i >= bn; i-- {
		f.box[i+n] = f.box[i]
	}
	f.nbox += n
}

// Close closess box n0-n1 inclusively. The rest are shifted down
func (f *Frame) Close(n0, n1 int) {
	if n0 >= f.nbox || n1 >= f.nbox || n1 < n0 {
		panic("Frame.Close")
	}
	n1++
	for i := n1; i < f.nbox; i++ {
		f.box[i-(n1-n0)] = f.box[i]
	}
	f.nbox -= n1 - n0
}

// Delete closes and deallocates n0-n1 inclusively
func (f *Frame) Delete(n0, n1 int) {
	if n0 >= f.nbox || n1 >= f.nbox || n1 < n0 {
		panic("Delete")
	}
	f.Free(n0, n1)
	f.Close(n0, n1)
}

// Free deallocates memory for boxes n0-n1 inclusively
func (f *Frame) Free(n0, n1 int) {
	if n1 < n0 {
		return
	}
	if n0 >= f.nbox || n1 >= f.nbox {
		panic("Free")
	}
	for i := n0; i < n1; i++ {
		if f.box[i].nrune >= 0 {
			f.box[i].ptr = nil
		}
	}
}

// Grow allocates memory for delta more boxes
func (f *Frame) Grow(delta int) {
	f.nalloc += delta
	f.box = append(f.box, make([]Box, delta)...)
}

// Dup copies the contents of box bn to box bn+1
func (f *Frame) Dup(bn int) {
	if f.box[bn].nrune < 0 {
		panic("Frame.Dup")
	}
	f.Add(bn, 1)
	if f.box[bn].nrune >= 0 {
		f.box[bn+1].ptr = append([]byte{}, f.box[bn].ptr...)
	}
}

// runeindex returns the index of the rune
// TODO: make it a rune, also it doesn't return a pointer number but index offset
func runeindex(s string, n int) int {
	return n
}

func (f *Frame) Truncate(b *Box, n int) {
	if b.nrune < 0 || b.nrune < n {
		panic("Truncate")
	}
	b.nrune -= n
	b.ptr = b.ptr[:b.nrune]
	b.width = f.stringwidth(b.ptr)
}

// Chop drops the first n chars in box b
func (f *Frame) Chop(b *Box, n int) {
	if b.nrune < 0 || b.nrune < n {
		panic("Chop")
	}
	copy(b.ptr, b.ptr[n:])
	b.nrune -= n
	b.ptr = b.ptr[:b.nrune]
	b.width = f.stringwidth(b.ptr)
}

// Split splits box bn into two boxes; bn and bn+1, at index n
func (f *Frame) Split(bn, n int) {
	f.Dup(bn)
	f.Truncate(&f.box[bn], (&f.box[bn]).nrune-n)
	f.Chop(&f.box[bn+1], n)
}

// Merge merges box bn and bn+1
func (f *Frame) Merge(bn int) {
	b0 := &f.box[bn]
	b1 := &f.box[bn+1]
	b0.ptr = append(b0.ptr, b1.ptr...)
	b0.width += b1.width
	b0.nrune += b1.nrune
	f.Delete(bn+1, bn+1)
}

//Find finds the box containing q starting from box bn index
// p and puts q at the start of the next box
func (f *Frame) Find(bn int, p, q int64) int {
	for ; bn < f.nbox; bn++ {
		b := &f.box[bn]
		if p+int64(NRUNE(b)) > q {
			break
		}
		p += int64(NRUNE(b))
	}
	if p != q {
		f.Split(bn, int(q-p))
		bn++
	}
	return bn
}

func (f *Frame) DumpBoxes() {
	fmt.Println("dumping boxes")
	fmt.Printf("nboxes: %d\n", f.nbox)
	fmt.Printf("nalloc: %d\n", f.nalloc)
	for i, b := range f.box {
		fmt.Printf("[%d] (%p) (nrune=%d l=%d w=%d mw=%d bc=%x): %q\n",
			i, &f.box[i], b.nrune, NRUNE(&b), b.width, b.minwidth, b.bc, b.ptr)
	}
}
