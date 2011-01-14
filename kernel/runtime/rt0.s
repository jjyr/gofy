TEXT _rt0_amd64_gofykernel(SB), 7, $0
	PUSHQ $runtime·tls0(SB)
	CALL runtime·setgs(SB)
	ADDQ $8, SP

	CALL main·init(SB)
	CALL main·main(SB)
here:
	CLI
	HLT
	JMP here

TEXT runtime·setgs(SB), 7, $0
	MOVL 8(SP), AX
	MOVL 12(SP), DX
	MOVL $0xC0000101, CX
	WRMSR
	RET

GLOBL runtime·tls0(SB), $64

TEXT runtime·panicindex(SB), 7, $0
	HLT

TEXT runtime·throwreturn(SB), 7, $0
	HLT
