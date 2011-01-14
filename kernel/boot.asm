bits 16
org 0x7C00
_start:
	cli
	xor bx, bx
	mov ds, bx
	mov ss, bx
	mov bx, 0x07E0
	mov es, bx
	mov sp, 0x7C00
	mov [drive], dl

	mov eax, 0x80000001
	cpuid
	test edx, 1<<29
	jz nolongmode

	mov byte [count], 1
	call readblock
	
	mov eax, [es:0]
	cmp eax, 0x978a0000
	jne wrongmagic

	mov eax, [es:4]
	bswap eax
	mov ebx, [es:8]
	bswap ebx
	add eax, ebx
	add eax, 511
	sar eax, 9
	cmp eax, 0x80
	jge toolarge
	mov byte [count], al
	call readblock

	in al, 0x92
	or al, 0x02
	out 0x92, al

	mov bx, 0x8000
	mov es, bx
	xor di, di
	xor al, al
	mov cx, 0x3000
	rep stosb
	
	mov dword [es:0x0000], firstPML4
	mov dword [es:0x1000], firstPDP
	mov dword [es:0x2000], firstPD

	mov eax, 0x80000
	mov cr3, eax

	mov eax, 0xA0
	mov cr4, eax

	mov ecx, 0xC0000080
	rdmsr
	or eax, 1<<8
	wrmsr

	mov eax, cr0
	or eax, (1<<31)|1
	mov cr0, eax
	lgdt [gdtptr]
	jmp 8:LongMode

readblock:
	mov si, 3
	movzx dx, byte [drive]
.doio:
	xor bx, bx
	mov cx, 2
	mov ah, 0x02
	mov al, [count]
	int 0x13
	jnc .iodone
	dec si
	jz .ioerror
	xor ah, ah
	int 0x13
	jmp .doio
.iodone:
	ret
.ioerror:
	mov si, iomsg
	jmp error


putstr:
	mov ah, 0xE
	mov bx, 0
	cld
.loop:
	lodsb
	test al, al
	jz .end
	int 0x10
	jmp .loop
.end:
	ret

nolongmode:
	mov si, nolongmsg
	jmp error

wrongmagic:
	mov si, nomagicmsg
	jmp error

toolarge:
	mov si, toolargemsg
	jmp error

error:
	call putstr
.L1:
	cli
	hlt
	jmp .L1


nolongmsg:	db "No AMD64", 13, 10, 0
nomagicmsg:	db "Magic", 13, 10, 0
toolargemsg:	db "Too large", 13, 10, 0
iomsg:		db "I/O", 13, 10, 0
drive equ 0x500
count equ 0x501
loc equ 0x502

; PML4 0x80000
; PDP 0x81000
; PD 0x82000

firstPML4 equ 0x8100F
firstPDP equ 0x8200F
firstPD equ 0x18F

gdt:
	dq 0
	dd 0
	dd (1<<15) | (1<<21) | (1<<12) | (1<<11)
	dd 0
	dd (1<<15) | (1<<12) | (1<<9)

gdtptr:
	dw 24
	dq gdt

bits 64
LongMode:
	cli
	mov ebx, 16
	mov ds, bx
	mov es, bx
	mov fs, bx
	mov gs, bx
	mov ss, bx
	mov rsp, 0x7C00

	mov ecx, [0x7E04] ; text size
	bswap ecx
	mov rsi, 0x7E28
	mov rdi, 0x10000
	rep movsb
	add rdi, 4095
	and rdi, ~4095
	mov ecx, [0x7E08] ; data size
	bswap ecx
	rep movsb
	mov ecx, [0x7E0C] ; bss size
	bswap ecx
	xor al, al
	rep stosb
	xor rax, rax
	mov eax, [0x7E14] ; entry point
	bswap eax
	jmp rax

bits 16
times 510-($-$$) db 0
dw 0xAA55
