#include "runtime.h"

int32 runtime·write(int32 fd, void* s, int32 len) {
	int8* ss = s;
	while(len--) runtime·putc(*ss++);
}
