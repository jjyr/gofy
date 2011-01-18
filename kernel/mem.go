package main

import "unsafe"

type Phys uint64

type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}


/*
	the whole address space is 256 TB
	a page directory pointer table adresses 512 GB
	a page directory addresses 1 GB
	a page table addresses 2 MB
	a page addresses 4 KB
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
	PAGETABLESIZE = 512
	PAGEAVAIL     uint64 = 1
	PAGEWRITE     uint64 = 2
	ANTIPAGE      = max32 ^ (PAGESIZE - 1)
)

type coreMapEntry struct {
	addr uint64
	size uint64
}

var (
	e820map [][3]uint64
	e820num int
	coremap [COREMAPSIZE]coreMapEntry
	pml4    []uint64
	tmppt   []uint64
	gdt     []uint64
	heap    []uint32
	curstack *uint64
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
	m := 0
	size := uint64(0)
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
			coremap[m] = coreMapEntry{addr: l, size: e820limits[i+1] - l}
			size += e820limits[i+1] - l
			m++
		}
	cont:
	}
	if size < 16777216 {
		fuck("Sorry, GOFY doesn't run on toasters")
	}
	print((size+524288)/1048576)
	println(" MB memory\n")
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
	mh.Len = e820num + 2
	mh.Cap = mh.Len
	e820map[e820num] = [3]uint64{0, 0x1000, E820RSVD}
	highest := *(*uint64)(unsafe.Pointer(uintptr(0x502)))
	highest = pageroundup(highest)
	e820map[e820num+1] = [3]uint64{uint64(kernelstart), highest - uint64(kernelstart), E820RSVD}
	e820num += 2
	processe820()
}

func fuck(s string) {
	println("SHIT IS BROKEN")
	println(s)
	for {}
}
