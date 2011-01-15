; PML4 0x80000
; PDP 0x81000
; PD 0x82000
; GDT 0x83000

drive equ 0x500
count equ 0x501
highest equ 0x502
memmap equ 0x600
loader equ 0x7C00
kernel equ 0x10000

AOUTMAGIC equ 0x978a0000
E820MAGIC equ 0x534D4150

firstPML4 equ 0x8100F
firstPDP equ 0x8200F
firstPD equ 0x18F
codeseg equ (1<<15) | (1<<21) | (1<<12) | (1<<11)
dataseg equ (1<<15) | (1<<12) | (1<<9)

bits 16
org loader
_start:
	cli
	xor bx, bx
	mov ds, bx
	mov ss, bx
	mov bx, kernel/0x10
	mov es, bx
	mov sp, loader
	mov [drive], dl

	mov byte [count], 1
	call readblock
	
	mov eax, [es:0]
	cmp eax, AOUTMAGIC 
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

	mov bx, memmap/0x10
	mov es, bx
	xor ebx, ebx
	mov dword [es:0], ebx
	mov di, 8
	mov edx, E820MAGIC 
	mov eax, 0xE820
	mov ecx, 24
	int 0x15
	jc e820error
	cmp eax, E820MAGIC 
	jnz e820error
e820loop:
	add di, 24
	mov eax, 0xE820
	mov ecx, 24
	int 0x15
	jc e820end
	inc dword [es:0]
	test ebx, ebx
	jnz e820loop
e820end:

	in al, 0x92
	or al, 0x02
	out 0x92, al

	mov bx, 0x8000
	mov es, bx
	xor di, di
	xor al, al
	mov cx, 0x4000
	rep stosb
	
	mov dword [es:0x0000], firstPML4
	mov dword [es:0x1000], firstPDP
	mov dword [es:0x2000], firstPD
	mov dword [es:0x300C], codeseg
	mov dword [es:0x3014], dataseg

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
	xchg al, ah
	add al, '0'
	mov ah, 0xE
	mov bx, 0x7
	int 0x10
	mov al, 'I'
	jmp error

wrongmagic:
	mov al, 'M'
	jmp error

toolarge:
	mov al, 'L'
	jmp error

e820error:
	mov al, 'E'
	jmp error

error:
	mov ah, 0xE
	mov bx, 0x7
	int 0x10
	jmp $

gdtptr:
	dw 24
	dq 0x83000

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

	mov ecx, [kernel + 0x04] ; text size
	bswap ecx
	mov rsi, kernel + 0x28
	mov rdi, 0x100000
	rep movsb
	add rdi, 4095
	and rdi, ~4095
	mov ecx, [kernel + 0x08] ; data size
	bswap ecx
	rep movsb
	mov ecx, [kernel + 0x0C] ; bss size
	bswap ecx
	xor al, al
	rep stosb
	mov [highest], rdi
	xor rax, rax
	mov eax, [kernel + 0x14] ; entry point
	bswap eax
	jmp rax

bits 16
times 510-($-$$) db 0
dw 0xAA55
