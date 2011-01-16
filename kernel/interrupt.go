package main

import "unsafe"

type intstate struct {
	ax, cx, dx, bx, sp, bp, si, di uint64
	r8, r9, r10, r11, r12, r13, r14, r15 uint64
	ds, cr2, gs uint64
	no, error, ip, cs, flags, usersp, ss uint64
}

type InterruptHandler func(intstate)

var (
	idt        []uint64
	asmhandler [256 * 16]byte
	inthandler [256]InterruptHandler
)

func lidt([]uint64)
func cli()
func sti()
func commonisraddr() uintptr

type StringHeader struct {
	Data uintptr
	Len  int
}


func int_unknown(st intstate) {
	putnum(st.no, 10)
	putc(10)
	fuck("unknown interrupt")
}

func int_pagefault(st intstate) {
	fuck("page fault")
}

func int_gpf(st intstate) {
	fuck("general protection fault")
}

func int_invalid(st intstate) {
	fuck("invalid instruction")
}

func initinterrupts() {
	idth := (*SliceHeader)(unsafe.Pointer(&idt))
	idth.Data = uintptr(falloc(1))
	idth.Len = 512
	idth.Cap = 512

	isr := uint64(commonisraddr())
	for j := 0; j < 256; j++ {
		i := j * 16
		asmhandler[i] = 0xFA
		if (i < 10) && (i != 8) || (i > 14) {
			asmhandler[i+1] = 0x6A
			asmhandler[i+2] = 0x00
		} else {
			asmhandler[i+1] = 0x90
			asmhandler[i+2] = 0x90
		}
		asmhandler[i+3] = 0x6A
		asmhandler[i+4] = byte(j)
		asmhandler[i+5] = 0xB8
		asmhandler[i+6] = byte(isr)
		asmhandler[i+7] = byte(isr >> 8)
		asmhandler[i+8] = byte(isr >> 16)
		asmhandler[i+9] = byte(isr >> 24)
		asmhandler[i+10] = 0xFF
		asmhandler[i+11] = 0xE0
		p := unsafe.Pointer(&asmhandler[i])
		off := uint64(uintptr(p))
		idt[j*2] = (off & 0xFFFF) | (8 << 16) | (0xE << 40) | (1 << 47) | ((off & 0xFFFF0000) << 32)
		idt[j*2+1] = off >> 32
		inthandler[j] = int_unknown
	}

	inthandler[0x6] = int_invalid
	inthandler[0xD] = int_gpf
	inthandler[0xE] = int_pagefault

	// remap IRQs
	outb(0x20, 0x11)
	outb(0xA0, 0x11)
	outb(0x21, 0x20)
	outb(0xA1, 0x28)
	outb(0x21, 0x04)
	outb(0xA1, 0x02)
	outb(0x21, 0x01)
	outb(0xA1, 0x01)
	outb(0x21, 0x00)
	outb(0xA1, 0x00)

	lidt(idt)
}
