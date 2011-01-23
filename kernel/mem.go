package main

import "unsafe"

type Phys uint64

type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func cleartlb()

/*
	the whole address space               is 256 TB
	a page directory pointer table addresses 512 GB
	a page directory               addresses   1 GB
	a page table                   addresses   2 MB
	a page                         addresses   4 KB
*/

const (
	COREMAPSIZE   = 512
	max32         = 0xFFFFFFFF
	max64         = 0xFFFFFFFFFFFFFFFF
	kernelstart   uintptr = 0x100000
	E820FREE      uint64  = 1
	E820RSVD      uint64  = 2
	MAXE820       = 100   // maximum number of E820 entries
	PAGESIZE      = 4096
	LARGEPAGESIZE = 2097152
	PAGETABLESIZE = 512
	PAGEAVAIL     uint64 = 1
	PAGEWRITE     uint64 = 2
	PAGEUSER      uint64 = 4
	PAGELARGE     uint64 = 0x80
	ANTIPAGE      = max32 ^ (PAGESIZE - 1)
)

type coreMapEntry struct {
	start   uint64
	end     uint64
	virtual uintptr
}

var (
	e820map     [][3]uint64
	e820num     int
	coremap     [COREMAPSIZE]coreMapEntry
	cmsize      int
	memsize     uint64
	maxvirtaddr uintptr
	kernelpml4  []uint64
)

func pageroundup(n uint64) uint64 {
	return (n + PAGESIZE - 1) & (max64 ^ (PAGESIZE - 1))
}

func processe820() {
	var e820limits [2 * MAXE820]uint64
	k := 0
	for i := 0; i < e820num; i++ {
		e820limits[k] = pageroundup(e820map[i][0])
		e820limits[k+1] = pageroundup(e820map[i][0] + e820map[i][1])
		k += 2
	}
	swapped := true
	for swapped {
		swapped = false
		for i := 1; i < k; i++ {
			if e820limits[i] == e820limits[i-1] {
				e820limits[i], e820limits[k-1] = e820limits[k-1], e820limits[i]
				swapped = true
				k--
			}
			if e820limits[i] < e820limits[i-1] {
				e820limits[i], e820limits[i-1] = e820limits[i-1], e820limits[i]
				swapped = true
			}
		}
	}
	cmsize = 0
	memsize := uint64(0)
	for i := 0; i < k-1; i++ {
		l := e820limits[i]
		found := false
		for j := 0; j < e820num; j++ {
			if l >= e820map[j][0] && l < e820map[j][0]+e820map[j][1] {
				if e820map[j][2] != E820FREE {
					goto cont
				} else {
					found = true
				}
			}
		}
		if found {
			coremap[cmsize] = coreMapEntry{start: l, end: e820limits[i+1]}
			memsize += e820limits[i+1] - l
			cmsize++
		}
	cont:
	}
	coremap[cmsize] = coreMapEntry{}
	if memsize < 16777216 {
		fuck("Sorry, GOFY doesn't run on toasters")
	}
	print((memsize + 524288) / 1048576)
	print(" MB memory\n")
}

func mapmemory() {
	pmlo, pdpo, pdo := 0, 0, 1
	pml := (*[PAGETABLESIZE]uint64)(unsafe.Pointer(uintptr(0x1000)))
	pdp := (*[PAGETABLESIZE]uint64)(unsafe.Pointer(uintptr(0x2000)))
	pd := (*[PAGETABLESIZE]uint64)(unsafe.Pointer(uintptr(0x3000)))
	virtual := uintptr(LARGEPAGESIZE)
	physical := uint64(LARGEPAGESIZE)
	offset := uintptr(0x4000)
	k := 0
	for ; k < cmsize; k++ {
		if coremap[k].start <= physical && physical < coremap[k].end {
			goto found
		}
	}
	fuck("E820 error")
found:
	coremap[k].virtual = virtual - uintptr(physical - coremap[k].start)
	for {
		pd[pdo] = physical | PAGELARGE | PAGEWRITE | PAGEAVAIL
		pdo++
		if pdo == PAGETABLESIZE {
			pd = (*[PAGETABLESIZE]uint64)(unsafe.Pointer(offset))
			pdp[pdpo] = uint64(offset) | PAGEWRITE | PAGEAVAIL
			offset += PAGESIZE
			pdpo++
			if pdpo == PAGETABLESIZE {
				pdp = (*[PAGETABLESIZE]uint64)(unsafe.Pointer(offset))
				pml[pmlo] = uint64(offset) | PAGEWRITE | PAGEAVAIL
				offset += PAGESIZE
				pmlo++
			}
		}
		virtual += LARGEPAGESIZE
		physical += LARGEPAGESIZE
		if physical >= coremap[k].end {
		retry:
			k++
			if k >= cmsize {
				break
			}
			physical = coremap[k].start
			if physical+LARGEPAGESIZE > coremap[k].end {
				goto retry
			}
			coremap[k].virtual = virtual - uintptr(physical - coremap[k].start)
		}
	}
	maxvirtaddr = virtual
	cleartlb()
}

func initmem() {
	e820num = int(*(*uint32)(unsafe.Pointer(uintptr(0x600))))
	if e820num == 0 {
		fuck("E820 fucked up")
	}
	if e820num > MAXE820 {
		fuck("E820 map too large")
	}

	mh := (*SliceHeader)(unsafe.Pointer(&e820map))
	mh.Data = 0x608
	mh.Len = e820num
	mh.Cap = mh.Len
	processe820()
	mapmemory()
	h := (*SliceHeader)(unsafe.Pointer(&kernelpml4))
	h.Data = 0x1000
	h.Len = 512
	h.Cap = 512
}

func fuck(s string) {
	println("SHIT IS BROKEN")
	println(s)
	for {
	}
}

func DecodePhys(p Phys) uintptr {
	if p < LARGEPAGESIZE {
		return uintptr(p)
	}
	for i := 0; i < cmsize; i++ {
		if p >= Phys(coremap[i].start) && p < Phys(coremap[i].end) {
			return uintptr(p) - uintptr(coremap[i].start) + coremap[i].virtual
		}
	}
	fuck("DecodePhys: tried accessing unmapped address")
	return 0xDEADBEEFDEADBEEF
}

func DecodeVirt(v uintptr) Phys {
	if v < LARGEPAGESIZE {
		return Phys(v)
	}
	for i := 0; i < cmsize; i++ {
		if v >= coremap[i].virtual && v < uintptr(coremap[i].end-coremap[i].start)+coremap[i].virtual {
			return Phys(v) - Phys(coremap[i].virtual) + Phys(coremap[i].start)
		}
	}
	fuck("DecodeVirt: tried accessing unmapped address")
	return 0xDEADBEEFDEADBEEF
}

func MakeTableSlice(p Phys) (res []uint64) {
	h := (*SliceHeader)(unsafe.Pointer(&res))
	h.Data = DecodePhys(p)
	h.Len = 512
	h.Cap = 512
	return
}

func AddrDecode(v uintptr) (pdp, pd, pt, page, disp int) {
	pdp = int((v >> 39) & 0x1FF)
	pd = int((v >> 30) & 0x1FF)
	pt = int((v >> 21) & 0x1FF)
	page = int((v >> 12) & 0x1FF)
	disp = int(v & (PAGESIZE - 1))
	return
}
