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

// the layout of this structure is hardcoded into GoUser, syscall, etc.
// don't fuck with it
type ProcState struct {
	ax, cx, dx, bx, sp, bp, si, di       uint64
	r8, r9, r10, r11, r12, r13, r14, r15 uint64
	ip, gs, flags                       uint64
}

type Process struct {
	ProcState
	time uint64
	PML4 uint64
	mem [][2]uint64
	highest uintptr  // highest virtual address
	error string
	ns Namespace
	fd []File
}

func GoUser(ProcState)
func SetProc(*Process)

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
	runtime.Memclr(_pml, PAGESIZE)
	pdp := p.KAllocate(1)
	Write64(_pml, pdp | PAGEAVAIL | PAGEWRITE | PAGEUSER)
	_pdp := runtime.MapTmp(pdp)
	runtime.Memclr(_pdp, PAGESIZE)
	Write64(_pdp, runtime.KernelPD | PAGEAVAIL | PAGEWRITE | PAGEUSER)
	runtime.FreeTmp(_pdp)
	runtime.FreeTmp(_pml)
	runtime.SetLocalCR3(p.PML4)
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
	magic := BE32(header[:])
	if magic != AOUTMAGIC {
		return SimpleError("invalid executable")
	}
	textsize := uint64(BE32(header[4:]))
	datasize := uint64(BE32(header[8:]))
	bsssize := uint64(BE32(header[12:]))
	procsize := pageroundup(textsize+40) + datasize + bsssize

	p.CleanUp()
	p.NewPML4()
	p.highest = USERSTART
	p.Allocate(procsize)
	off := uint64(0)
	v = USERSTART
	end := USERSTART + uintptr(pageroundup(textsize+40)+datasize)
	for v < end {
		n, err := f.PRead(buf[:], off)
		if err != nil {
			if v+uintptr(n) >= end && err == EOF {
				break
			}
			return err
		}
		runtime.Memmove(v, bufh.Data, uint32(n))
		off += n
		v += uintptr(n)
	}
	runtime.Memclr(v, bsssize)
	p.ProcState.ip = uint64(BE32(header[20:]))
	return nil
}

func (p *Process) Run() {
	SetProc(p)
	GoUser(p.ProcState)
}

func (q *Process) Dawn(flags uint64, c chan bool) {
	highest := q.highest
	pml4 := q.PML4
	q.highest = USERSTART
	q.NewPML4()
	q.Allocate(uint64(highest - USERSTART))
	buf := make([]byte, PAGESIZE)
	bufh := (*runtime.SliceHeader)(unsafe.Pointer(&buf))
	for v := uintptr(USERSTART); v < highest; v += PAGESIZE {
		runtime.SetLocalCR3(pml4)
		runtime.Memmove(bufh.Data, v, PAGESIZE)
		runtime.SetLocalCR3(q.PML4)
		runtime.Memmove(v, bufh.Data, PAGESIZE)
	}
	q.ProcState.ax = 0
	c<-false
	q.Run()
}

func (p *Process) Fork(flags uint64) (q *Process) {
	c := make(chan bool)
	q = new(Process)
	*q = *p
	go q.Dawn(flags, c)
	<-c
	return
}
