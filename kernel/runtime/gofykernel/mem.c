#include "runtime.h"
#include "malloc.h"

uint64 runtime·highest;
extern uint64 main·maxvirtaddr;

void*
runtime·SysAlloc(uintptr ask)
{
	void* ret;

	ret = (void*) runtime·highest;
	runtime·highest += ask;
	return ret;
}

void
runtime·SysFree(void *v, uintptr n)
{
}

void
runtime·SysUnused(void *v, uintptr n)
{
}

void
runtime·SysMemInit(void)
{
}
