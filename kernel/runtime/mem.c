#include "runtime.h"

/*
	the whole address space               is 256 TB
	a page directory pointer table addresses 512 GB
	a page directory               addresses   1 GB
	a page table                   addresses   2 MB
	a page                         addresses   4 KB
*/

enum {
	COREMAPSIZE   = 512,
	kernelstart   = 0x100000,
	E820FREE      = 1,
	E820RSVD      = 2,
	MAXE820       = 100,   // maximum number of E820 entries
	PAGESIZE      = 4096,
	LARGEPAGESIZE = 2097152,
	PAGETABLESIZE = 512,
	PAGEAVAIL     = 1,
	PAGEWRITE     = 2,
	PAGEUSER      = 4,
	PAGELARGE     = 0x80,
	ANTIPAGE      = ~(PAGESIZE - 1)
};

void main·fuck(int8*, uint32);

typedef struct coreMapEntry coreMapEntry;

struct coreMapEntry {
	uint64 start, end;
};

uint64* e820map;
uint32 e820num;
coreMapEntry coremap[COREMAPSIZE];
uint32 cmsize;
uint64 memsize;
uint64* kernelpml4;
uint64 e820limits[2 * MAXE820];

#define pageroundup(n) (((n) + PAGESIZE - 1) & ~(PAGESIZE - 1))

#pragma textflag 7
void
runtime·processe820(void)
{
	uint32 i, j, k;
	uint64 *l, *lk, t;
	bool swapped, found;

	lk = e820limits;
	for(i = 0; i < e820num; i++) {
		*lk++ = pageroundup(e820map[3*i]);
		*lk++ = (e820map[3*i] + e820map[3*i+1]) & ANTIPAGE;
	}
	swapped = true;
	while(swapped) {
		swapped = false;
		for(l=e820limits + 1; l < lk; l++) {
			if(*l == *(l-1)) {
				lk--;
				t = *l;
				*l = *lk;
				*lk = t;
				swapped = true;
			}
			if(*l < *(l-1)) {
				t = *l;
				*l = *(l-1);
				*(l-1) = t;
				swapped = true;
			}
		}
	}
	cmsize = 0;
	memsize = 0;
	for(l = e820limits; l < lk-1; l++) {
		found = false;
		runtime·printf("%p\n", *l);
		for(j = 0; j < e820num; j++) {
			if(*l >= e820map[3*j] && *l < e820map[3*j]+e820map[3*j+1]) {
				if(e820map[3*j+2] != E820FREE) {
					goto cont;
				} else {
					found = true;
				}
			}
		}
		if(found) {
			coremap[cmsize].start = *l;
			coremap[cmsize].end = *(l+1);
			memsize += *(l+1) - *l;
			cmsize++;
		} 
	cont: ;
	}
	coremap[cmsize].start = 0;
	coremap[cmsize].end = 0;
	runtime·printf("%d MB core\n", (memsize + 524288) / 1048576);
	if(memsize < 16777216) {
		int8 s[] = "Sorry, GOFY doesn't run on toasters";
		main·fuck(s, sizeof(s));
	}
}

#pragma textflag 7
void
runtime·initmem()
{
	e820num = *(uint32*)(0x600);
	if(e820num == 0) {
		int8 s[] = "E820 fucked up";
		main·fuck(s, sizeof(s));
	}
	if(e820num > MAXE820) {
		int8 s[] = "E820 map too large";
		main·fuck(s, sizeof(s));
	}
	e820map = (uint64*) 0x608;
	runtime·processe820();
}

/*
func DecodePhys(p Phys) uintptr {
	if p < LARGEPAGESIZE {
		return uintptr(p)
	}
	for i = 0; i < cmsize; i++ {
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
	for i = 0; i < cmsize; i++ {
		if v >= coremap[i].virtual && v < uintptr(coremap[i].end-coremap[i].start)+coremap[i].virtual {
			return Phys(v) - Phys(coremap[i].virtual) + Phys(coremap[i].start)
		}
	}
	fuck("DecodeVirt: tried accessing unmapped address")
	return 0xDEADBEEFDEADBEEF
}

func MakeTableSlice(p Phys) (res []uint64) {
	h = (*SliceHeader)(unsafe.Pointer(&res))
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
*/
