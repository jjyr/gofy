#include "asm.h"
#include "msr.h"

TEXT syscallentry(SB), 7, $0
	SWAPGS
	MOVQ 0(GS), SI
	MOVQ 8(SI), BP
	MOVQ SP, 32(BP)
	MOVQ CX, 128(BP)
	MOVQ $0, 0220(BP)
	MOVQ tss+4(SB), SP

	MOVQ main·sysent(SB)(AX*8), AX
	PUSHQ BP
	CALL *AX
	POPQ BP

	MOVQ 0(BP), AX
	MOVQ 0x20(BP), SP
	MOVQ 128(BP), CX
	MOVQ 0220(BP), R11
	SWAPGS
	BYTE $0x48
	BYTE $0x0F
	BYTE $0x07
