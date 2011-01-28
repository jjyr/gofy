package runtime

import "unsafe"

var (
	vram               []uint16
	consolex, consoley int
)

type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

type StringHeader struct {
	Data uintptr
	Len int
}

func outb(uint16, uint8)

func initconsole() {
	consolex, consoley = 0, 0
	h := (*SliceHeader)(unsafe.Pointer(&vram))
	h.Data = 0xB8000
	h.Len = 80 * 25
	h.Cap = h.Len
	for i := 0; i < 80*25; i++ {
		vram[i] = 0x0F00
	}
}

var putting bool

func putc(c int) {
	BeginCritical()
	switch c {
	case 10:
		consolex = 0
		consoley++
	default:
		vram[consolex+consoley*80] = 0x0F00 | uint16(tocp437(c))
		consolex++
		if consolex >= 80 {
			consolex = 0
			consoley++
		}
	}
	if consoley >= 25 {
		consoley--
		for i := 0; i < 80*24; i++ {
			vram[i] = vram[i+80]
		}
		for i := 80 * 24; i < 80*25; i++ {
			vram[i] = 0x0F00
		}
	}
	p := consolex + consoley*80
	outb(0x3D4, 0x0F)
	outb(0x3D5, byte(p))
	outb(0x3D4, 0x0E)
	outb(0x3D5, byte(p>>8))
	EndCritical()
}
