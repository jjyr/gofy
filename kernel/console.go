package main

import (
	"unsafe"
)

var (
	vram []uint16
	consolex, consoley int
)

type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func consoleinit() {
	h := (*SliceHeader)(unsafe.Pointer(&vram))
	h.Data = 0xB8000
	h.Len = 80*25
	h.Cap = h.Len
	for i := 0; i < 80*25; i++ {
		vram[i] = 0
	}
}

func putc(c int) {
	switch c {
	case 10:
		consolex = 0
		consoley++
	default:
		vram[consolex+consoley*80] = 0x0F00 | uint16(tocp437(c))
		consolex++
	}
}

func putnum(l uint64) {
	if l != 0 {
		putnum(l>>4)
	}
	l &= 15
	if l > 9 {
		putc(int(l) + 'A' - 10)
	} else {
		putc(int(l) + '0')
	}
}

func puts(s string) {
	for _, v := range s {
		putc(v)
	}
}

func fuck(s string) {
	puts("SHIT IS BROKEN\n")
	puts(s)
	for {}
}
