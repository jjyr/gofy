#include "runtime.h"

enum { COM1 = 0x3F8 };

void runtime·outb(uint16 port, uint8 data);
uint8 runtime·inb(uint16 port);

void
runtime·initeia232()
{
	runtime·outb(COM1 + 1, 0x00);
	runtime·outb(COM1 + 3, 0x80);
	runtime·outb(COM1 + 0, 0x03);
	runtime·outb(COM1 + 1, 0x00);
	runtime·outb(COM1 + 3, 0x03);
	runtime·outb(COM1 + 2, 0xC7);
	runtime·outb(COM1 + 4, 0x0B);
}

void
runtime·puteia232(int8 c)
{
	if(c == 10) runtime·puteia232(13);
	while(!(runtime·inb(COM1 + 5) & 0x20));
	runtime·outb(COM1, c);
}
