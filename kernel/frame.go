package main

import "unsafe"

type Phys uint64

const (
	COREMAPSIZE = 512
	max32 uint64 = 0xFFFFFFFF
	max64 uint64 = 0xFFFFFFFFFFFFFFFF
	kernelstart uint64 = 0x100000
	E820FREE uint64 = 1
	E820RSVD uint64 = 2
	MAXE820 = 100 // maximum number of E820 entries
)

type coreMapEntry struct {
	addr Phys
	len uint64
}

var (
	e820map [][3]uint64
	e820num int
	coremap [COREMAPSIZE]coreMapEntry
	coremaplen int
)

func processe820() {
	var e820limits [2*MAXE820]uint64
	k := 0
	for i := 0; i < e820num; i++ {
		e820limits[k] = e820map[i][0]
		e820limits[k+1] = e820map[i][0] + e820map[i][1]
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
	for i := 0; i < k-1; i++ {
		l := e820limits[i]
		found := false
		for j := 0; j < e820num; j++ {
			if l >= e820map[j][0] && l < e820map[j][0] + e820map[j][1] {
				if e820map[j][2] != E820FREE {
					goto cont
				} else {
					found = true
				}
			}
		}
		if found {
			coremap[coremaplen] = coreMapEntry{addr: Phys(l), len: e820limits[i+1] - l}
			coremaplen++
		}
		cont:
	}
}

func frameinit() {
	e820num = int(*(*uint32)(unsafe.Pointer(uintptr(0x600))))
	if e820num == 0 {
		fuck("E820 fucked up")
	}
	if e820num > MAXE820 {
		fuck("E820 map too large")
	}
	
	mh := (*SliceHeader)(unsafe.Pointer(&e820map))
	mh.Data = 0x608	
	mh.Len = e820num+2
	mh.Cap = mh.Len
	e820map[e820num] = [3]uint64{0, 0x1000, E820RSVD}
	highest := *(*uint64)(unsafe.Pointer(uintptr(0x502)))
	highest = (highest + 4095) & (max64 ^ 4095)
	e820map[e820num+1] = [3]uint64{kernelstart, highest - kernelstart, E820RSVD}
	e820num += 2
	processe820()
}

func falloc(n int) Phys {
	if n <= 0 {
		return 0
	}
	return 0
}
