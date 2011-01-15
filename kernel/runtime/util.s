TEXT main·invlpg(SB), $0
	MOVQ 8(SP), AX
	BYTE $0x0F
	BYTE $0x01
	BYTE $0x38 // INVLPG (AX)
	RET

TEXT main·setcr3(SB), $0
	MOVQ 8(SP), AX
	MOVQ AX, CR3
	RET

TEXT main·lgdt(SB), $10
	MOVQ addr+0(FP), AX
	MOVQ AX, 2(SP)
	MOVW len+8(FP), AX
	SHLW $3, AX
	MOVW AX, 0(SP)
	BYTE $0x0F // LGDT (SP)
	BYTE $0x01
	BYTE $0x14
	BYTE $0x24
	RET

TEXT main·loadsegs(SB), $0
	MOVL $16, BX
	MOVW BX, DS
	MOVW BX, ES
	MOVW BX, SS
	SUBQ $8, SP
	MOVQ 8(SP), AX
	MOVQ AX, 0(SP)
	MOVL $8, 8(SP)
	BYTE $0x48 // RETF
	BYTE $0xCB

TEXT main·outb(SB), $0
	MOVB data+2(FP), AX
	MOVW addr+0(FP), DX
	BYTE $0xEE // OUTB AX, DX
	RET

TEXT main·halt(SB), $0
	HLT
