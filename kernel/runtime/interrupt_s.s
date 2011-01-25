#define GSBASE 0xC0000101

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
        SUBQ $152, SP
        MOVQ AX, 0(SP)
        MOVQ CX, 8(SP)
        MOVQ DX, 16(SP)
        MOVQ BX, 24(SP)
        MOVQ SP, 32(SP)
        MOVQ BP, 40(SP)
        MOVQ SI, 48(SP)
        MOVQ DI, 56(SP)
        MOVQ R8, 64(SP)
        MOVQ R9, 72(SP)
        MOVQ R10, 80(SP)
        MOVQ R11, 88(SP)
        MOVQ R12, 96(SP)
        MOVQ R13, 104(SP)
        MOVQ R14, 112(SP)
        MOVQ R15, 120(SP)
        MOVW DS, BX
        MOVQ BX, 128(SP)
        MOVQ CR2, BX
        MOVQ BX, 136(SP)
        MOVL $GSBASE, CX
        RDMSR
        MOVL AX, 144(SP)
        MOVL DX, 148(SP)
	MOVQ $stack0(SB), AX
	MOVQ AX, DX
	SHRQ $32, DX
	WRMSR
        MOVQ 152(SP), AX
        SHLQ $3, AX
        ADDQ $isr(SB), AX
        CALL *(AX)
        MOVL $GSBASE, CX
        MOVL 144(SP), AX
        MOVL 148(SP), DX
        WRMSR
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
        ADDQ $152, SP
	ADDQ $16, SP // error code and interrupt number
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
