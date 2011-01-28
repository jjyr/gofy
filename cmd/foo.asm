bits 64
org 0x40000000
%define be(x) ((x & 0xFF) << 24) | ((x & 0xFF00) << 8) | ((x & 0xFF0000) >> 8) | ((x & 0xFF000000) >> 24)
dd be(0x8a97)
dd be(textsize)
dd be(datasize)
dd be(bsssize)
dd be(0)
dd be(0x40000028)
dd be(0)
dd be(0)
dd 0
dd 0

text:
	mov rsp, stack0

	push 0
	mov rax, 6
	syscall
	jc prerror
	add rsp, 8

	add al, '0'
	mov [hello], al

print:
	push 0
	push hellolen
	push hello
	push 2
	mov rax, 2
	syscall
	jc prerror
	add rsp, 32

	mov rcx, 100000
	loop $

	jmp print

prerror:
	push 128
	push buf
	mov rax, 1
	syscall
	add rsp, 8

	push rax
	push buf
	push 1
	mov rax, 2
	syscall
	add rsp, 12

	jmp $

hello: db "Hello, World", 10
hellolen equ $ - hello

textsize equ $ - text

data:
datasize equ $ - data

section .bss
bss:

buf: resb 128

resb 4096
stack0:

bsssize equ $ - bss
