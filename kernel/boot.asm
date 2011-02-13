drive equ 0x500 ; byte
count equ 0x504 ; dword
memmap equ 0x600
header equ 0x1000
; 64 KB buffer
buf equ 0x10000
lba equ 1
AOUTMAGIC equ 0x978a0000
E820MAGIC equ 0x534D4150
%define toseg(addr) ((addr & 0xFFFF0)<<12) | (addr & 0xF)

bits 16
org 0x7C00
_start:
	cli
	xor bx, bx
	mov ds, bx
	mov ss, bx
	mov sp, 0x7C00
	push loop
	mov [drive], dl

a20:
	in al, 0x92
	or al, 0x02
	out 0x92, al

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
	xor bx, bx
	mov es, bx

readheader:
	mov byte [bno], 1
	call readblock
	mov si, header
	cmp dword [si], AOUTMAGIC
	jnz invalidmagic
	mov eax, [si+4]
	bswap eax
	add eax, 4095 + 40
	and eax, ~4095
	mov ebx, [si+8]
	bswap ebx
	add eax, ebx
	add eax, 511
	shr eax, 9
	mov [count], eax
	mov dword [dst], toseg(buf)
	
readin:
	mov al, 3
	call putchr
	mov ecx, [count]
	cmp ecx, 0x80
	jl .go
	mov cl, 0x80
.go:
	mov [bno], cl
	call readblock

protected:
	lgdt [gdt]
	mov eax, cr0
	or al, 1
	mov cr0, eax
	jmp 8:.here
bits 32
.here:
	mov bx, 16
	mov ds, bx
	mov es, bx
	mov ss, bx

copy:
	mov edi, [target]
	mov esi, buf
	mov ecx, 65536/4
	rep movsd
	movzx ecx, byte [packet+2]
	add dword [target], 65536
	add dword [src], ecx
	adc dword [src+4], 0
	sub dword [count], ecx
	jle bootit

realmode:
	jmp 24:.here2
bits 16
.here2:
	mov eax, cr0
	and eax, ~1
	mov cr0, eax
	jmp 0:.here3
.here3:
	xor bx, bx
	mov ds, bx
	mov es, bx
	mov ss, bx
	jmp readin

bits 32
bootit:
	mov eax, [header+0x14]
	bswap eax
	jmp eax

bits 16

readblock:
	mov si, packet
	mov ah, 0x42
	mov dl, [drive]
	int 0x13
	jc ioerror
	test ah, ah
	jnz ioerror
	ret

loop:
	jmp $
ioerror:
	xchg al, ah
	add al, '0'
	call putchr
	mov al, 'I'
	call putchr
	jmp $
e820error:
	mov al, '8'
	jmp putchr
invalidmagic:
	mov al, 'M'
putchr:
	mov ah, 0xE
	mov bx, 0x7
	int 0x10
	ret


target:
	dd 0x100000

gdt:
	dw 32
	dd gdt
	dw 0
	dd 0xFFFF
	dd (0xF<<16) | (1<<15) | (1<<12) | (1<<11) | (1<<22) | (1<<23)
	dd 0xFFFF
	dd (0xF<<16) | (1<<15) | (1<<12) | (1<<9) | (1<<23)
	dd 0xFFFF
	dd (0xF<<16) | (1<<15) | (1<<12) | (1<<11) | (1<<23)

packet:
	dw 0x10
bno:
	dw 0
dst:
	dd header
src:
	dq lba

times 446-($-$$) db 0
db 0x80, 0, 1, 0, 0x83, 5, 0x91, 0x67
dd 8193
dd 54537
times 510-($-$$) db 0
dw 0xAA55
