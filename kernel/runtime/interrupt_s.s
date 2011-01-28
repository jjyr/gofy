#include "msr.h"

TEXT runtime·lidt(SB), 7, $10
	MOVQ addr+0(FP), AX
	MOVQ AX, 2(SP)
	MOVW $4096, 0(SP)
	BYTE $0x0F // LIDT (SP)
	BYTE $0x01
	BYTE $0x1C
	BYTE $0x24
	RET

TEXT common_isr(SB), 7, $0
        SUBQ $0220, SP
        MOVQ AX,  0000(SP)
        MOVQ CX,  0010(SP)
        MOVQ DX,  0020(SP)
        MOVQ BX,  0030(SP)
        MOVQ SP,  0040(SP)
        MOVQ BP,  0050(SP)
        MOVQ SI,  0060(SP)
        MOVQ DI,  0070(SP)
        MOVQ R8,  0100(SP)
        MOVQ R9,  0110(SP)
        MOVQ R10, 0120(SP)
        MOVQ R11, 0130(SP)
        MOVQ R12, 0140(SP)
        MOVQ R13, 0150(SP)
        MOVQ R14, 0160(SP)
        MOVQ R15, 0170(SP)
        MOVW DS, BX
        MOVQ BX,  0200(SP)
        MOVQ CR2, BX
        MOVQ BX,  0210(SP)
	SWAPGS

        MOVQ 0220(SP), AX
        SHLQ $3, AX
        ADDQ $isr(SB), AX
        CALL *(AX)

	CLI // if interrupt got enabled for whatever reason, disable them before shit happens w.r.t. GS
	SWAPGS
        MOVQ 128(SP), BX
        MOVW BX, DS
        MOVW BX, ES
        MOVQ 0(SP), AX
        MOVQ 8(SP), CX
        MOVQ 16(SP), DX
        MOVQ 24(SP), BX
        MOVQ 40(SP), BP
        MOVQ 48(SP), SI
        MOVQ 56(SP), DI
        MOVQ 64(SP), R8
        MOVQ 72(SP), R9
        MOVQ 80(SP), R10
        MOVQ 88(SP), R11
        MOVQ 96(SP), R12
        MOVQ 104(SP), R13
        MOVQ 112(SP), R14
        MOVQ 120(SP), R15
        ADDQ $(144+16), SP
        WORD $0xCF48

TEXT runtime·cli(SB), 7, $0
	CLI
	RET

TEXT runtime·sti(SB), 7, $0
	STI
	RET

TEXT runtime·cr2(SB), 7, $0
	MOVQ CR2, AX
	RET
