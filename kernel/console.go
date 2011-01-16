package main

import (
	"unsafe"
)

var (
	vram               []uint16
	consolex, consoley int
)

type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func outb(uint16, uint8)
func halt()

func initconsole() {
	h := (*SliceHeader)(unsafe.Pointer(&vram))
	h.Data = 0xB8000
	h.Len = 80 * 25
	h.Cap = h.Len
	for i := 0; i < 80*25; i++ {
		vram[i] = 0x0F00
	}
}

func putc(c int) {
	switch c {
	case 10:
		consolex = 0
		consoley++
		if consoley >= 25 {
			consoley--
			for i := 0; i < 80*24; i++ {
				vram[i] = vram[i+80]
			}
			for i := 80 * 24; i < 80*25; i++ {
				vram[i] = 0x0F00
			}
		}
	default:
		vram[consolex+consoley*80] = 0x0F00 | uint16(tocp437(c))
		consolex++
	}
	p := consolex + consoley*80
	outb(0x3D4, 0x0F)
	outb(0x3D5, byte(p))
	outb(0x3D4, 0x0E)
	outb(0x3D5, byte(p>>8))
}

func putnum_(l uint64, base int, first bool) {
	if l != 0 {
		putnum_(l/uint64(base), base, false)
	} else {
		if first {
			putc('0')
		}
		return
	}
	l %= uint64(base)
	if l > 9 {
		putc(int(l) + 'A' - 10)
	} else {
		putc(int(l) + '0')
	}
}

func putnum(l uint64, base int) {
	putnum_(l, base, true)
}

func puts(s string) {
	for _, v := range s {
		putc(v)
	}
}

func fuck(s string) {
	puts("SHIT IS BROKEN\n")
	puts(s)
	putc(10)
	cli()
	for {
		halt()
	}
}
