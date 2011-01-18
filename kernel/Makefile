GOOS:=gofykernel
GOARCH:=amd64

include $(GOROOT)/src/Make.inc

DEPS=runtime
TARG=kernel
GOFILES=\
	main.go\
	mem.go\

LD+= -H 2 -T0x100028
PREREQ+=boot

all: image

kernel.elf: _go_.6
	6l -H 7 -T0x100028 -o $@ $<

image: boot kernel
	cat boot kernel > image
	dd if=/dev/zero count=100 >> image 2>/dev/null

boot: boot.asm
	nasm -f bin $< -o $@

.PHONY: gofmt
gofmt:
	( for i in $(GOFILES); do gofmt -w $$i; done )

.PHONY: run rund
run: image
	qemu-system-x86_64 -hda image

rund: image
	qemu-system-x86_64 -hda image -d in_asm,int,cpu -no-reboot

include $(GOROOT)/src/Make.cmd
