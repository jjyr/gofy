TEXT	runtime·stringiter(SB), 7, $0
	XORL	AX, AX
	MOVQ	8(SP), SI // string address
	MOVL	24(SP), BX // position
	CMPL	BX, 16(SP) // string length
	JGE	stringiterout
	ADDQ	BX, SI
	MOVL	BX, AX
	INCL	AX
	TESTB	$0x80, (SI)
	JZ	stringiterout
	INCL	AX
	TESTB	$0x20, (SI)
	JZ	stringiterout
	INCL	AX
	TESTB	$0x10, (SI)
	JZ	stringiterout
	INCL	AX
stringiterout:
	MOVL	AX, 32(SP)
	RET

TEXT	runtime·stringiter2(SB), 7, $0
	XORL	AX, AX
	XORL	CX, CX
	XORL	DX, DX
	MOVQ	8(SP), SI // string address
	MOVL	24(SP), BX // position
	CMPL	BX, 16(SP) // string length
	JGE	stringiter2end
	ADDQ	BX, SI
	TESTB	$0x80, (SI)
	JZ	stringiter2ascii
	TESTB	$0x40, (SI)
	JZ	stringiter2invalid
	TESTB	$0x20, (SI)
	JZ	stringiter2len2
	TESTB	$0x10, (SI)
	JZ	stringiter2len3
stringiter2len4:
	MOVL	$4, AX

	MOVB	3(SI), DX
	ANDB	$0x3F, DX
	ORL	DX, CX

	MOVB	2(SI), DX
	ANDB	$0x3F, DX
	SHLL	$6, DX
	ORL	DX, CX

	MOVB	1(SI), DX
	ANDB	$0x3F, DX
	SHLL	$12, DX
	ORL	DX, CX

	MOVB	0(SI), DX
	ANDB	$0x07, DX
	SHLL	$18, DX
	ORL	DX, CX
	JMP	stringiter2out
stringiter2ascii:
	MOVL	$1, AX
	MOVB	(SI), CX
	JMP	stringiter2out
stringiter2len2:
	MOVL	$2, AX

	MOVB	1(SI), DX
	ANDB	$0x3F, DX
	ORL	DX, CX

	MOVB	(SI), DX
	ANDB	$0x1F, DX
	SHLL	$6, DX
	ORL	DX, CX

	JMP	stringiter2out
stringiter2len3:
	MOVL	$3, AX

	MOVB	2(SI), DX
	ANDB	$0x3F, DX
	ORL	DX, CX

	MOVB	1(SI), DX
	ANDB	$0x3F, DX
	SHLL	$6, DX
	ORL	DX, CX

	MOVB	(SI), DX
	ANDB	$0x0F, DX
	SHLL	$12, DX
	ORL	DX, CX
	JMP	stringiter2out
stringiter2invalid:
	MOVL	$1, AX
	MOVL	$0xFFFE, CX
stringiter2out:
	ADDL	BX, AX
stringiter2end:
	MOVL	AX, 32(SP)
	MOVL	CX, 36(SP)
	RET
