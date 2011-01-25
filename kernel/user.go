package main

import (
	"unsafe"
	"runtime"
)

const (
	AOUTMAGIC = 0x8A97
	PAGESIZE = 4096
	ANTIPAGE = 0xFFFFFFFFFFFFF000
	USERSTART = 512 * 512 * 4096 // 1 GB
	PAGEAVAIL = 1
	PAGEWRITE = 2
	PAGEUSER  = 4
)

// the layout of this structure is hardcoded into GoUser
type ProcState struct {
	ax, cx, dx, bx, sp, bp, si, di       uint64
	r8, r9, r10, r11, r12, r13, r14, r15 uint64
	ip, gs, rflags                       uint64
}

type Process struct {
	ProcState
	PML4 uint64
	mem [][2]uint64
	highest uintptr  // highest virtual address
}

func GoUser(ProcState)

func LE32(b []byte) (r uint32) {
	r |= uint32(b[0]) << 24
	r |= uint32(b[1]) << 16
	r |= uint32(b[2]) << 8
	r |= uint32(b[3])
	return
}

func pageroundup(p uint64) uint64 {
	return (p + PAGESIZE - 1) & ANTIPAGE
}

// free all associated memory
func (p *Process) CleanUp() {
	for _, v := range p.mem {
		runtime.Ffree(v[0], v[1])
	}
	p.PML4 = 0
	p.mem = nil
}

func (p *Process) KAllocate(size uint64) uint64 {
	phys := runtime.Falloc(size)
	p.mem = append(p.mem, [2]uint64{phys, size})
	return phys
}

func (p *Process) Allocate(size uint64) {
	size = (size + PAGESIZE - 1) / PAGESIZE
	phys := p.KAllocate(size)
	runtime.MapMem(p.PML4, phys, p.highest, uint32(size), PAGEUSER)
	p.highest += uintptr(size * PAGESIZE)
}

func Write64(n uintptr, v uint64) {
	*(*uint64)(unsafe.Pointer(n)) = v
}

func (p *Process) NewPML4() {
	p.PML4 = p.KAllocate(1)
	_pml := runtime.MapTmp(p.PML4)

	pdp := p.KAllocate(1)
	Write64(_pml, pdp | PAGEAVAIL | PAGEWRITE | PAGEUSER)
	_pdp := runtime.MapTmp(pdp)
	Write64(_pdp, runtime.KernelPD | PAGEAVAIL | PAGEWRITE | PAGEUSER)
	runtime.FreeTmp(_pdp)

	pdp = p.KAllocate(1)
	Write64(_pml + 511 * 8, pdp | PAGEAVAIL | PAGEWRITE | PAGEUSER)
	_pdp = runtime.MapTmp(pdp)
	pd := p.KAllocate(1)
	Write64(_pdp + 511 * 8, pd | PAGEAVAIL | PAGEWRITE | PAGEUSER)
	_pd := runtime.MapTmp(pd)
	Write64(_pd + 511 * 8, runtime.TmpPageTable | PAGEAVAIL | PAGEWRITE | PAGEUSER)
	runtime.FreeTmp(_pd)
	runtime.FreeTmp(_pdp)

	runtime.FreeTmp(_pml)
}

func (p *Process) Exec(f File) Error {
	var header [40]byte
	var v uintptr
	buf := make([]byte, PAGESIZE)
	bufh := (*runtime.SliceHeader)(unsafe.Pointer(&buf))

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

	cr3 := runtime.GetCR3()
	p.CleanUp()
	p.NewPML4()
	p.highest = USERSTART
	p.Allocate(procsize)
	runtime.SetCR3(p.PML4)
	off := uint64(40)
	for v = USERSTART; v < USERSTART + uintptr(textsize & ANTIPAGE); v += PAGESIZE {
		_, err := f.PRead(buf[:], off)
		if err != nil {
			return err
		}
		off += PAGESIZE
		runtime.Memmove(v, bufh.Data, PAGESIZE)
	}
	if v < USERSTART + uintptr(textsize) {
		n := USERSTART + uintptr(textsize) - v
		_, err := f.PRead(buf[:n], off)
		if err != nil {
			return err
		}
		off += uint64(n)
		runtime.Memmove(v, bufh.Data, uint32(n))
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
		runtime.Memmove(v, bufh.Data, PAGESIZE)
	}
	p.ProcState.ip = USERSTART
	runtime.SetCR3(cr3)
	return nil
}

func (p *Process) Run() {
	runtime.SetCR3(p.PML4)
	GoUser(p.ProcState)
}
