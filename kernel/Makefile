GOOS:=gofykernel
GOARCH:=amd64

include $(GOROOT)/src/Make.inc

DEPS=runtime
TARG=kernel
GOFILES=\
	main.go\
	console.go\
	cp437.go\
	mem.go\
	interrupt.go\

LD+= -H 2 -s -T0x100000
PREREQ+=boot

image: boot kernel
	cat boot kernel > image

boot: boot.asm
	nasm -f bin $< -o $@

.PHONY: gofmt
gofmt:
	( for i in $(GOFILES); do gofmt $$i > /tmp/a; mv /tmp/a $$i; done )

.PHONY: run rund
run: image
	qemu-system-x86_64 -fda image

rund: image
	qemu-system-x86_64 -fda image -d in_asm,int,cpu -no-reboot

include $(GOROOT)/src/Make.cmd
