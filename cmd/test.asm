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
	push 1
	push hellolen
	push hello
	mov rax, 4
	syscall
	jc prerror
	add rsp, 32

	push 128
	push buf
	push rax
	mov rax, 3
	syscall
	add rsp, 12

	push rax
	push buf
	push 0
	mov rax, 2
	syscall
	jc prerror

	jmp $

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

hello: db "/hello.txt"
hellolen equ $ - hello

textsize equ $ - text

data:
datasize equ $ - data

bss:

buf: resb 128

resb 4096
stack0:

bsssize equ $ - bss
