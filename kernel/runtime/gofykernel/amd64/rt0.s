#include "amd64/asm.h"

#define HEADER 0x100000
#define PML4 0x1000
#define PDP0 0x2000
#define PD0 0x3000

TEXT _rt0_amd64_gofykernel(SB), 7, $0
	MODE $32
	MOVL $PML4, DI
	MOVL $0, AX
	MOVL $0x4000, CX
	REP; STOSB

	MOVL $(PDP0|0xF), PML4
	MOVL $(PD0|0xF), PDP0
	MOVL $0x18F, PD0

	MOVL CR4, AX
	BTSL $5, AX
	MOVL AX, CR4

	MOVL $PML4, AX
	MOVL AX, CR3

	MOVL $0xC0000080, CX
	RDMSR
	ORL $0x100, AX
	WRMSR

	MOVL CR0, AX
	BTSL $31, AX
	MOVL AX, CR0

	LONG $0x2514010F // LGDT
	LONG $gdtptr(SB)

	BYTE $0xEA // LJMP
	LONG $now64(SB)
	WORD $8

TEXT now64(SB), 7, $0
	MODE $64
	MOVW $16, BX
	MOVW BX, DS
	MOVW BX, ES
	MOVW BX, FS
	MOVW BX, GS
	MOVW BX, SS

	MOVQ $stack0(SB), DI
	MOVQ $0, AX
	MOVQ $0x1000, CX
	REP; STOSB

	MOVQ $runtime·g0(SB), g(BX)
	MOVQ $runtime·m0(SB), m(BX)
	
	MOVQ $stack0(SB), SP
	ADDQ $4096, SP
	MOVQ $stack0(SB), AX
	MOVQ AX, DX
	SHRQ $32, DX
	MOVQ $0xC0000101, CX
	WRMSR

	CALL runtime·initconsole(SB)
	CALL main·initmem(SB)
	HLT

TEXT gdt(SB), 7, $0
	QUAD $0
	QUAD $0x20980000000000
	QUAD $0x00920000000000

TEXT gdtptr(SB), 7, $0
	WORD $24
	QUAD $gdt(SB)

GLOBL stack0(SB), $4096

TEXT runtime·notok(SB), 7, $0
	HLT

TEXT runtime·gettime(SB), 7, $0
	HLT

TEXT runtime·settls(SB), 7, $0
	HLT

TEXT runtime·signame(SB), 7, $0
	HLT

TEXT runtime·outb(SB), 7, $0
	MOVB data+2(FP), AX
	MOVW addr+0(FP), DX
	BYTE $0xEE
	RET
