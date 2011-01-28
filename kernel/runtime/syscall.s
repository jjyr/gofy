#include "asm.h"
#include "msr.h"
#include "proc.h"

TEXT syscallentry(SB), 7, $0
	SWAPGS
	MOVQ g(CX), SI
	MOVQ g_process(SI), BP
	MOVQ SP, P_SP(BP)
	MOVQ CX, P_IP(BP)
	BTRQ $0, P_FLAGS(BP)
	MOVQ tss+4(SB), SP
	PUSHQ AX
	MOVL $KERNELGS, CX
	RDMSR
	MOVL AX, P_GS(BP)
	MOVL DX, (P_GS+4)(BP)
	DECL CX
	RDMSR
	INCL CX
	WRMSR
	POPQ AX
	PUSHQ P_TIME(BP)
	STI

	MOVQ main·sysent(SB)(AX*8), AX
	PUSHQ BP
	CALL *AX
	POPQ BP

	POPQ AX
	CMPQ AX, P_TIME(BP) // check if a timer interrupt occured
	JEQ noswitch

	PUSHQ BP
	CALL runtime·gosched(SB)
	POPQ BP
	MOVQ SP, tss+4(SB)

noswitch:
	CLI
	MOVL $KERNELGS, CX
	MOVL P_GS(BP), AX
	MOVL (P_GS+4)(BP), DX
	WRMSR
	SWAPGS

	MOVQ P_AX(BP), AX
	MOVQ P_SP(BP), SP
	MOVQ P_IP(BP), CX
	MOVQ P_FLAGS(BP), R11
	BYTE $0x48 // SYSRETQ
	BYTE $0x0F
	BYTE $0x07
