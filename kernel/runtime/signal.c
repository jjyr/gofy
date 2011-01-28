#include "runtime.h"

int32 runtime·write(int32 fd, void* s, int32 len) {
	int8* ss = s;
	while(len--) {
		runtime·puteia232(*ss);
		runtime·putc(*ss++);
	}
}
