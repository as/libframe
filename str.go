package main

import "fmt"

const CHUNK = 16

func ROUNDUP(n uint) uint {
	return (n + CHUNK) &^ (CHUNK - 1)
}
func NBYTE(b *Box) int {
	if b.nrune < 0 {
		return 1
	}
	return strlen(b.ptr)
}

func NRUNE(b *Box) int {
	return NBYTE(b)
}
func (f *Frame) allocstr(n uint) []byte {
	return make([]byte, n)
	//return make([]byte, ROUNDUP(n))
}

func (f *Frame) insure(bn int, n uint) {
	return
	b := &f.box[bn]
	if b.nrune < 0 {
		panic("frinsure")
	}
	if ROUNDUP(uint(b.nrune)) > n {
		return
	}
	p := f.allocstr(n)
	b = &f.box[bn]
	copy(p, b.ptr[:NBYTE(b)+1])
	b.ptr = p
}

func strlen(s []byte) int {
	return len(s)
	fmt.Printf("string: %q\n", s)
	i := 0
	for i = range s {
		if s[i] == '\x00' {
			return i
		}
	}
	panic("strlen")
}

func strcpy(dst, src []byte) {
	copy(dst, src[:strlen(src)])
}

func (f *Frame) stringwidth(p []byte) int {
	return f.measure(p[:strlen(p)])
}
