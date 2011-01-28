TEXT runtime·write(SB), 7, $32
	LEAQ fd+0(FP), SI
	MOVQ SP, DI
	MOVSQ ; MOVSQ ; MOVSQ
	MOVQ $0, 24(SP)
	MOVQ $2, AX
	SYSCALL
	RET

TEXT runtime·gettime(SB), 7, $0
	RET

TEXT runtime·settls(SB), 7, $8
	MOVQ DI, (SP)
	MOVQ $7, AX
	SYSCALL
	RET

TEXT runtime·notok(SB), 7, $0
	HLT

TEXT runtime·signame(SB), 7, $0
	HLT
