package main

import (
	"unsafe"
	"runtime"
)

const (
	AOUTMAGIC = 0x8A97
	USERSTART = PAGETABLESIZE * PAGETABLESIZE * PAGETABLESIZE * PAGESIZE
)

type ProcState struct {
	ax, cx, dx, bx, sp, bp, si, di       uint64
	r8, r9, r10, r11, r12, r13, r14, r15 uint64
	ip, gs                               uint64
}

type UserMemEntry struct {
	c []byte
	v uintptr // 0 for page tables etc.
}

type Process struct {
	ProcState
	PML4    [512]uint64
	mem     []UserMemEntry
	highest uintptr  // highest virtual address
}

func SetCR3(*[512]uint64)

func LE32(b []byte) (r uint32) {
	r |= uint32(b[0]) << 24
	r |= uint32(b[1]) << 16
	r |= uint32(b[2]) << 8
	r |= uint32(b[3])
	return
}

func (p *Process) KAllocate(size uint64) uintptr {
	size &= ANTIPAGE
	m := make([]byte, size)
	mh := (*SliceHeader)(unsafe.Pointer(&m))
	p.mem = append(p.mem, UserMemEntry{c: m})
	return mh.Data
}

func (p *Process) Allocate(size uint64) {
	pmlo, pdpo, pdo, pto, _ := AddrDecode(p.highest - 1)
	PDP := MakeTableSlice(Phys(p.PML4[pmlo] & ANTIPAGE))
	PD := MakeTableSlice(Phys(PDP[pdpo] & ANTIPAGE))
	PT := MakeTableSlice(Phys(PD[pdo] & ANTIPAGE))
	v := p.KAllocate(size)
	p.mem[len(p.mem)-1].v = p.highest
	p.highest += uintptr(size)
	vend := v + uintptr(size)
	for ; v < vend; v += PAGESIZE {
		pto++
		if pto == PAGETABLESIZE {
			pto = 0
			pdo++
			if pdo == PAGETABLESIZE {
				pdo = 0
				pdpo++
				if pdpo == PAGETABLESIZE {
					pdpo = 0
					pmlo++
					n := p.KAllocate(PAGESIZE)
					p.PML4[pmlo] = uint64(n) | PAGEAVAIL | PAGEWRITE | PAGEUSER
					PDP = MakeTableSlice(DecodeVirt(n))
				}
				n := p.KAllocate(PAGESIZE)
				PDP[pdpo] = uint64(n) | PAGEAVAIL | PAGEWRITE | PAGEUSER
				PD = MakeTableSlice(DecodeVirt(n))
			}
			n := p.KAllocate(PAGESIZE)
			PD[pdo] = uint64(n) | PAGEAVAIL | PAGEWRITE | PAGEUSER
			PT = MakeTableSlice(DecodeVirt(n))
		}
		PT[pto] = uint64(DecodeVirt(v)) | PAGEAVAIL | PAGEWRITE | PAGEUSER
	}
}

func (p *Process) DecodeAddr(a uintptr) uintptr {
	for _, v := range p.mem {
		if v.v == 0 {
			continue
		}
		sh := (*SliceHeader)(unsafe.Pointer(&v.c))
		if a >= v.v && a < v.v + uintptr(sh.Len) {
			return a - v.v + sh.Data
		}
	}
	fuck("Process.DecodeAddr: tried accessing unmapped address")
	return 0xDEADBEEFDEADBEEF
}

func (p *Process) Exec(f File) Error {
	var header [40]byte
	var v uintptr
	buf := make([]byte, PAGESIZE)
	bufh := (*SliceHeader)(unsafe.Pointer(&buf))

	_, err := f.PRead(header[:], 0)
	if err != nil {
		return err
	}
	magic := LE32(header[:])
	if magic != AOUTMAGIC {
		return SimpleError("invalid executable")
	}
	textsize := uint64(LE32(header[4:]))
	datasize := uint64(LE32(header[8:]))
	bsssize := uint64(LE32(header[12:]))
	procsize := pageroundup(textsize) + datasize + bsssize

	p.mem = []UserMemEntry{}
	p.PML4 = [512]uint64{}
	p.PML4[0] = kernelpml4[0]
	p.highest = USERSTART
	p.Allocate(procsize)
	off := uint64(40)
	for v = USERSTART; v < USERSTART + uintptr(textsize) & ANTIPAGE; v += PAGESIZE {
		_, err := f.PRead(buf[:], off)
		if err != nil {
			return err
		}
		off += PAGESIZE
		a := p.DecodeAddr(v)
		runtime.Memmove(a, bufh.Data, PAGESIZE)
	}
	if v < USERSTART + uintptr(textsize) {
		n := USERSTART + uintptr(textsize) - v
		_, err := f.PRead(buf[:n], off)
		if err != nil {
			return err
		}
		off += uint64(n)
		a := p.DecodeAddr(v)
		runtime.Memmove(a, bufh.Data, uint32(n))
		v += PAGESIZE
	}
	for ; v < USERSTART + uintptr(pageroundup(textsize) + datasize) ; v += PAGESIZE {
		_, err := f.PRead(buf[:], off)
		if err == EOF {
			break
		}
		if err != nil {
			return err
		}
		off += PAGESIZE
		a := p.DecodeAddr(v)
		runtime.Memmove(a, bufh.Data, PAGESIZE)
	}
	p.ProcState.ip = USERSTART + 40
	return nil
}

func (p *Process) Run() {
	SetCR3(&p.PML4)
}
