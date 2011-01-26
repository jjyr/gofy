#include "asm.h"
#include "msr.h"

TEXT syscallentry(SB), 7, $0
	SWAPGS
	MOVQ 0(GS), SI
	MOVQ 8(SI), BP
	MOVQ SP, 32(BP)
	MOVQ CX, 128(BP)
	MOVQ tss+4(SB), SP

	MOVQ 32(BP), AX
	MOVQ (AX), AX
	MOVQ main·sysent(SB)(AX*8), AX
	PUSHQ BP
	CALL *AX
	POPQ BP

	MOVQ 32(BP), SP
	MOVQ 128(BP), CX
	SWAPGS
	SYSRET
