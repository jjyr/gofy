TEXT _rt0_amd64_gofy(SB), 7, $0
	MOVL $runtime·stack0+4096(SB), SP
	SUBL $24, SP
	MOVQ $0, 0(SP)
	MOVQ $0, 8(SP)
	MOVQ $0, 16(SP)
	JMP _rt0_amd64(SB)

GLOBL runtime·stack0(SB), $4096
