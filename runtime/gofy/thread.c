#include "runtime.h"

extern void* runtime·schedp;

int8 *goos = "gofy";

void
runtime·minit(void)
{
}

void
runtime·goenvs(void)
{
}

void
runtime·initsig(int32)
{
}

void
runtime·exit(int32)
{
	for(;;);
}

// single processor, no interrupts,
// so no need for real concurrency or atomicity

void
runtime·newosproc(M *m, G *g, void *stk, void (*fn)(void))
{
	USED(m, g, stk, fn);
	runtime·throw("newosproc");
}

void
runtime·lock(Lock *l)
{
	if(m->locks < 0)
		runtime·throw("lock count");
	m->locks++;
	if(l->key != 0)
		runtime·throw("deadlock");
	l->key = 1;
}

void
runtime·unlock(Lock *l)
{
	m->locks--;
	if(m->locks < 0)
		runtime·throw("lock count");
	if(l->key != 1)
		runtime·throw("unlock of unlocked lock");
	l->key = 0;
}

void 
runtime·destroylock(Lock*)
{
	// nothing
}

void
runtime·noteclear(Note *n)
{
	n->lock.key = 0;
}

void
runtime·notewakeup(Note *n)
{
	n->lock.key = 1;
}

void
runtime·notesleep(Note *n)
{
	if(n->lock.key != 1)
		runtime·throw("notesleep");
}

void
runtime·SysMemInit()
{
}

void
runtime·SysFree()
{
}

void
runtime·SysAlloc()
{
}

void
runtime·osinit()
{
}
