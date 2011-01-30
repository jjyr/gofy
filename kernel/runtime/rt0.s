#include "asm.h"
#include "msr.h"
#include "gdt.h"

#define HEADER 0x100000
#define PML4 0x1000
#define PDP0 0x2000
#define PD0 0x3000

TEXT _rt0_amd64_gofykernel(SB), 7, $0
	MODE $32
	CLI

	// parse a.out header and clear bss
	MOVL (HEADER+4), AX
	WORD $0xC80F
	ADDL $(4095+40), AX // header size
	ANDL $~4095, AX
	MOVL (HEADER+8), BX
	WORD $0xCB0F
	ADDL BX, AX
	ADDL $HEADER, AX
	MOVL AX, DI
	MOVL (HEADER+12), CX
	WORD $0xC90F
	XORL AX, AX
	REP; STOSB
	MOVL DI, runtime·highest(SB)
	MOVL $0, runtime·highest+4(SB)

	MOVL $PML4, DI
	MOVL $0, AX
	MOVL $0x4000, CX
	REP; STOSB

	MOVL $(PDP0|0xF), PML4
	MOVL $(PD0|0xF), PDP0
	MOVL $0x18F, PD0
	MOVL $0x200018F, (PD0+8)

	MOVL CR4, AX
	BTSL $5, AX
	MOVL AX, CR4

	MOVL $PML4, AX
	MOVL AX, CR3

	MOVL $EFER, CX
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
	WORD $GDTKCODE

TEXT now64(SB), 7, $0
	MODE $64
	MOVW $GDTKDATA, BX
	MOVW BX, DS
	MOVW BX, ES
	MOVW BX, FS
	MOVW BX, GS
	MOVW BX, SS

	MOVQ $tss(SB), AX
	SHLQ $16, AX
	ORQ AX, gdt+GDTTSS(SB)

	MOVQ $GDTTSS, BX
	BYTE $0x0F
	BYTE $0x00
	BYTE $0xDB

	MOVQ $stack0(SB), DI
	MOVQ $0, AX
	MOVQ $0x1000, CX
	REP; STOSB

	MOVQ $stack0(SB), SP
	ADDQ $4096, SP
	MOVL $stack0(SB), AX
	MOVL $0, DX
	MOVQ $GSBASE, CX
	WRMSR
	INCL CX
	WRMSR

	MOVQ $runtime·g0(SB), CX
	MOVQ CX, g(BX)
	MOVQ $runtime·m0(SB), AX
	MOVQ AX, m(BX)
	MOVQ CX, m_g0(AX)

	MOVQ SP, g_stackbase(CX)
	MOVQ $stack0(SB), g_stackguard(CX)

	CALL runtime·initconsole(SB)
	CALL runtime·initeia232(SB)
	CALL runtime·initinterrupts(SB)

	// set up SYSCALL
	MOVL $STAR, CX
	MOVL $0, AX
	MOVL $(GDTKCODE | ((GDTUCODE32|3) << 16)), DX
	WRMSR
	INCL CX
	MOVL $syscallentry(SB), AX
	MOVL $0, DX
	WRMSR
	INCL CX
	WRMSR
	INCL CX
	MOVL $0x200, AX
	WRMSR
	MOVL $EFER, CX
	RDMSR
	BTSL $0, AX
	WRMSR

	CALL runtime·schedinit(SB)
	PUSHQ $runtime·mainstart(SB)
	PUSHQ $0
	CALL runtime·newproc(SB)
	POPQ AX
	POPQ AX
	CALL runtime·mstart(SB)
	CALL runtime·notok(SB)
	RET

TEXT gdt(SB), 7, $0
	QUAD $0
	QUAD $0x20980000000000 // kernel code
	QUAD $0x00920000000000 // kernel data
	QUAD $0                // user 32bit (unused)
	QUAD $0x00F20000000000 // user data
	QUAD $0x20F80000000000 // user 64bit
	QUAD $0x0089000000006C // TSS (address is filled in by software)
	QUAD $0                // rest of TSS

TEXT gdtptr(SB), 7, $0
	WORD $0100
	QUAD $gdt(SB)

GLOBL stack0(SB), $4096

GLOBL tss(SB), $108

TEXT runtime·notok(SB), 7, $0
	HLT

TEXT runtime·gettime(SB), 7, $0
	RET

TEXT runtime·settls(SB), 7, $0
	HLT

TEXT runtime·signame(SB), 7, $0
	HLT

TEXT main·outb(SB), 7, $0
TEXT runtime·outb(SB), 7, $0
	MOVB data+2(FP), AX
	MOVW addr+0(FP), DX
	OUTB
	RET

TEXT runtime·inb(SB), 7, $0
	MOVW addr+0(FP), DX
	INB
	RET

TEXT main·inb(SB), 7, $0
	MOVW addr+0(FP), DX
	INB
	MOVB AX, data+8(FP)
	RET

TEXT main·outl(SB), 7, $0
	MOVL data+4(FP), AX
	MOVW addr+0(FP), DX
	OUTL
	RET

TEXT main·inl(SB), 7, $0
	MOVW addr+0(FP), DX
	INL
	MOVL AX, data+8(FP)
	RET

TEXT runtime·SetCR3(SB), 7, $0
	MOVQ cr3+0(FP), AX
	MOVQ AX, CR3
	RET

TEXT runtime·GetCR3(SB), 7, $0
	MOVQ CR3, AX
	MOVQ AX, cr3+0(FP)
	RET

TEXT runtime·FlushTLB(SB), 7, $0
	MOVQ CR3, AX
	MOVQ AX, CR3
	RET

TEXT runtime·InvlPG(SB), 7, $0
	MOVQ 8(SP), AX
	BYTE $0x0F
	BYTE $0x01
	BYTE $0x38
	RET

TEXT runtime·Halt(SB), 7, $0
	CLI
	HLT
	JMP runtime·Halt(SB)

TEXT main·GoUser(SB), 7, $40
	CLI
	MOVQ SP, tss+4(SB)
	// set GS
	SWAPGS
	MOVL $GSBASE, CX
	MOVL gs+136(FP), AX
	MOVL gs2+140(FP), DX
	WRMSR

	MOVQ cx+0010(FP), CX
	MOVQ dx+0020(FP), DX
	MOVQ bx+0030(FP), BX
	MOVQ bp+0050(FP), BP
	MOVQ si+0060(FP), SI
	MOVQ di+0070(FP), DI
	MOVQ r8+0100(FP), R8
	MOVQ r9+0110(FP), R9
	MOVQ r10+0120(FP), R10
	MOVQ r11+0130(FP), R11
	MOVQ r12+0140(FP), R12
	MOVQ r13+0150(FP), R13
	MOVQ r14+0160(FP), R14
	MOVQ r15+0170(FP), R15

	MOVQ ip+128(FP), AX
	MOVQ AX, 0(SP)
	MOVQ $(GDTUCODE | 3), 8(SP) // CS
	MOVQ flags+144(FP), AX
	MOVQ AX, 16(SP)
	MOVQ sp+32(FP), AX
	MOVQ AX, 24(SP)
	MOVQ $(GDTUDATA | 3), 32(SP) // SS

	MOVQ ax+0(FP), AX

	WORD $0147510

TEXT runtime·readMSR(SB), 7, $0
	MOVL msr+0(FP), CX
	RDMSR
	SHLQ $32, DX
	ORQ DX, AX
	RET

TEXT runtime·writeMSR(SB), 7, $0
	MOVL msr+0(FP), CX
	MOVL var+8(FP), AX
	MOVL var2+12(FP), DX
	WRMSR
	RET

TEXT runtime·getTSSSP(SB), 7, $0
	MOVQ tss+4(SB), AX
	RET

TEXT runtime·setTSSSP(SB), 7, $0
	MOVQ val+0(FP), AX
	MOVQ AX, tss+4(SB)
	RET

TEXT main·InPIO(SB), 7, $0
	MOVW port+0(FP), DX
	MOVQ buf+8(FP), DI
	MOVL len+16(FP), CX
	SHRQ $1, CX
	CLD
	REP
	INSW
	RET

TEXT main·OutPIO(SB), 7, $0
	MOVW port+0(FP), DX
	MOVQ buf+8(FP), SI
	MOVL len+16(FP), CX
	SHRQ $1, CX
	OUTSW
	DECL CX
	WORD $0xFA75
	RET
