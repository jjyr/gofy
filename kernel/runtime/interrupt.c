#include "runtime.h"
#include "gdt.h"
#include "msr.h"
#include "proc.h"

typedef struct IntState IntState;
struct IntState {
	uint64 ax, cx, dx, bx, sp, bp, si, di,
	r8, r9, r10, r11, r12, r13, r14, r15,
	ds, cr2, 
	no, error, ip, cs, flags, usersp, ss;
};

typedef void (*InterruptHandler)(IntState st);

uint64 idt[512];
byte asmhandler[16*256];
InterruptHandler isr[256];

void common_isr();
void main·fuck(int8* s, int32 len);
void runtime·outb(uint16 addr, uint8 data);
void runtime·lidt(uint64* idt);
uint64 runtime·cr2(void);
void runtime·sti(void);
void runtime·cli(void);
uint64 runtime·readMSR(uint64 msr);
void runtime·writeMSR(uint64 msr, uint64 value);
uint64 runtime·getTSSSP(void);
void runtime·setTSSSP(uint64 sp);

#pragma textflag 7
static void
puthex(uint64 n, uint32 l, int8* t)
{
	static int8 hexdigits[] = "0123456789ABCDEF";
	while(l--) {
		*--t = hexdigits[n & 0xF];
		n >>= 4;
	}
}

#pragma textflag 7
static void
int_divbyzero(IntState)
{
	static int8 s[] = "Division by zero";
	main·fuck(s, sizeof(s));
}

#pragma textflag 7
static void
int_invalid(IntState st)
{
	static int8 s[] = "Invalid instruction at 0x0000000000000000";
	puthex(st.ip, 16, s+sizeof(s));
	main·fuck(s, sizeof(s));
}

#pragma textflag 7
static void
int_unknown(IntState st)
{
	static int8 s[] = "Unknown interrupt 0x00";
	puthex(st.no, 2, s+sizeof(s));
	main·fuck(s, sizeof(s));
}

#pragma textflag 7
static void
int_pagefault(IntState st)
{
	static int8 s[] = "Page fault at address 0x0000000000000000";
	puthex(runtime·cr2(), 16, s+sizeof(s));
	main·fuck(s, sizeof(s));
}

#pragma textflag 7
static void
int_gpf(IntState)
{
	static int8 s[] = "General protection fault";
	main·fuck(s, sizeof(s));
}

#pragma textflag 7
static void
resetpic(uint64 no)
{
	if(no >= 40) runtime·outb(0xA0, 0x20);
	runtime·outb(0x20, 0x20);
}

#pragma textflag 7
static void
int_timer(IntState st)
{
	resetpic(st.no);
	if(g && g->process)
		(*(uint64*)((uint8*)g->process + P_TIME))++;
	// we have been summoned from user mode
	if(st.cs & 3) {
		uint64 gs, ksp;
		gs = runtime·readMSR(KERNELGS);
		ksp = runtime·getTSSSP();
		runtime·writeMSR(KERNELGS, runtime·readMSR(GSBASE));
		runtime·gosched();
		runtime·cli();
		runtime·setTSSSP(ksp);
		runtime·writeMSR(KERNELGS, gs);
	}
}

#pragma textflag 7
static void
int_kb(IntState st)
{
	resetpic(st.no);
}

#pragma textflag 7
static void
int_ide(IntState st)
{
	resetpic(st.no);
}

void
runtime·initinterrupts(void)
{
	int32 i, j;
	uint64 off;
	for(j=0;j<256;j++) {
		i = j * 16;
		asmhandler[i] = 0xFA;
                if((i < 10) && (i != 8) || (i > 14)) {
                        asmhandler[i+1] = 0x6A;
                        asmhandler[i+2] = 0x00;
                } else {
                        asmhandler[i+1] = 0x90;
                        asmhandler[i+2] = 0x90;
                }
                asmhandler[i+3] = 0x6A;
                asmhandler[i+4] = j;
                asmhandler[i+5] = 0xB8;
                asmhandler[i+6] = (uint64) common_isr;
                asmhandler[i+7] = (uint64) common_isr >> 8;
                asmhandler[i+8] = (uint64) common_isr >> 16;
                asmhandler[i+9] = (uint64) common_isr >> 24;
                asmhandler[i+10] = 0xFF;
                asmhandler[i+11] = 0xE0;
		off = (uint64) asmhandler + i;
		idt[j*2] = (off & 0xFFFF) | 0x8e0000000000LL | (GDTKCODE << 16) | ((off & 0xFFFF0000) << 32LL);
		idt[j*2+1] = off >> 32;
		isr[j] = int_unknown;
	}
	isr[0x00] = int_divbyzero;
	isr[0x06] = int_invalid;
	isr[0x0D] = int_gpf;
	isr[0x0E] = int_pagefault;
	isr[0x20] = int_timer;
	isr[0x21] = int_kb;
	isr[0x2E] = int_ide;
	isr[0x2F] = int_ide;
        runtime·outb(0x20, 0x11);
        runtime·outb(0xA0, 0x11);
        runtime·outb(0x21, 0x20);
        runtime·outb(0xA1, 0x28);
        runtime·outb(0x21, 0x04);
        runtime·outb(0xA1, 0x02);
        runtime·outb(0x21, 0x01);
        runtime·outb(0xA1, 0x01);
        runtime·outb(0x21, 0x00);
        runtime·outb(0xA1, 0x00);
        runtime·lidt(idt);
	runtime·BeginCritical();
}

static int32 critical = 0;

void
runtime·BeginCritical(void)
{
	runtime·cli();
	critical++;
}

void
runtime·EndCritical(void)
{
	if(critical > 0)
		critical--;
	if(critical == 0)
		runtime·sti();
}
