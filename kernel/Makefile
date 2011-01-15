GOOS:=gofykernel
GOARCH:=amd64

include $(GOROOT)/src/Make.inc

DEPS=runtime
TARG=kernel
GOFILES=\
	main.go\
	console.go\
	cp437.go\
	frame.go\

LD+= -H 2 -s -T0x100000
PREREQ+=boot

image: boot kernel
	cat boot kernel > image

boot: boot.asm
	nasm -f bin $< -o $@

.PHONY: run
run: image
	qemu-system-x86_64 -fda image

include $(GOROOT)/src/Make.cmd
