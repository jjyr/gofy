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
	KERNELSTART   = 0x100000,
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
	ANTIPAGE      = ~(PAGESIZE - 1),
	PAGETABLESTART = 0xffffc0000000LL,
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
uint64 e820limits[2 * MAXE820];
uint64 runtime·highest;
uint64* pagetables[2];

#define pageroundup(n) (((n) + PAGESIZE - 1) & ~(PAGESIZE - 1))

#pragma textflag 7
void
runtime·processe820(void)
{
	uint32 i, j;
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
			if(cmsize && coremap[cmsize-1].end == *l) {
				coremap[cmsize-1].end = *(l+1);
				memsize += *(l+1) - *l;
			} else {
				coremap[cmsize].start = *l;
				coremap[cmsize].end = *(l+1);
				memsize += *(l+1) - *l;
				cmsize++;
			}
		} 
	cont: ;
	}
	coremap[cmsize].start = 0;
	coremap[cmsize].end = 0;
	runtime·printf("%d MB core\n", (uint32)((memsize + 524288) / 1048576));
	if(memsize < 16777216) {
		int8 s[] = "Sorry, GOFY doesn't run on toasters";
		main·fuck(s, sizeof(s));
	}
}

#pragma textflag 7
void
runtime·SysMemInit()
{
	e820num = *(uint32*)0x600;
	if(e820num == 0) {
		int8 s[] = "E820 fucked up";
		main·fuck(s, sizeof(s));
	}
	if(e820num > MAXE820) {
		int8 s[] = "E820 map too large";
		main·fuck(s, sizeof(s));
	}
	e820map[3*e820num+0] = KERNELSTART;
	e820map[3*e820num+1] = KERNELSTART - runtime·highest;
	e820map[3*e820num+2] = E820RSVD;
	e820map[3*e820num+3] = 0;
	e820map[3*e820num+4] = 0x10000;
	e820map[3*e820num+5] = E820RSVD;
	e820num += 2;
	e820map = (uint64*) 0x608;
	runtime·processe820();
}

#pragma textflag 7
void
runtime·SysFree()
{
}

#pragma textflag 7
uint64
falloc(uint64 n)
{
	uint32 i;
	uint64 p;

	if(!n) return 0;
        for(i = 0; i < cmsize; i++) {
                if(coremap[i].end - coremap[i].start >= PAGESIZE*n) {
                        p = coremap[i].start;
                        coremap[i].start += PAGESIZE * n;
                        if(coremap[i].start == coremap[i].end) {
                                i++;
				cmsize--;
                                for(; i < cmsize; i++) coremap[i-1] = coremap[i];
                        }
                        return p;
                }
        }
	int8 s[] = "out of memory";
	main·fuck(s, sizeof(s));
        return 0;
}

void
ffree(uint64 p, uint64 n) {
	uint32 i;

	if(!n) return;
	for(i = 0; i < cmsize && coremap[i].start <= i; i++);
	if(i && coremap[i-1].end == p) {
		coremap[i-1].end += n;
		if(p+n == coremap[i].start) {
			coremap[i-1].end = coremap[i].end;
			i++;
			cmsize--;
			for(; i < cmsize; i++) {
				coremap[i-1] = coremap[i];
			}
		}
	} else {
		if(p+n == coremap[i].start) {
			coremap[i].start -= n;
		} else {
			coreMapEntry *j;

			j = coremap + i;
			for(; i < cmsize; i++) {
				coremap[i+1] = coremap[i];
			}
			cmsize++;
			j->start = p;
			j->end = p+n;
		}
        }
}

#pragma textflag 7
void*
runtime·SysAlloc(uintptr n)
{
}

uint64*
runtime·AllocTable(void)
{
	uint64 p;

	p = falloc(1);
}
