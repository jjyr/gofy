GOOS:=gofykernel
GOARCH:=amd64

include $(GOROOT)/src/Make.inc

NM=/usr/x86_64/bin/x86_64-gnu-linux-nm
DEPS=runtime
TARG=kernel
GOFILES=\
	main.go\
	file.go\
	initrd.go\
	user.go\
	initrddata.go\
	syscall.go\
	bio.go\
	pci.go\
	ide.go\
	binary.go\

LD+= -H 2 -T0x100028 
CFLAGS+= -FVw 
PREREQ+=boot

all: image

kernel.elf: _go_.6
	6l -d -H 7 -T0x100028 -o $@ $<

kernel.sym: kernel.elf
	$(NM) kernel.elf | grep " T " | awk '{ print $$1 " " $$3 }' > kernel.sym

image: boot kernel
	cat boot kernel > image
	dd if=/dev/zero of=image seek=62730 count=0 2>/dev/null

boot: boot.asm
	nasm -f bin $< -o $@

.PHONY: gofmt
gofmt:
	( for i in $(GOFILES); do gofmt -w $$i; done )

.PHONY: run runng rund rundd
run: image
	qemu-system-x86_64 -hda image $@

runng: image
	qemu-system-x86_64 -hda image $@ -nographic

rund: image
	qemu-system-x86_64 -hda image -d in_asm,int -no-reboot

rundd: image kernel.elf
	qemu-system-x86_64 -hda image -s -S &
	gdb kernel.elf -ex 'target remote localhost:1234' -ex 'set architecture i386:x86-64:intel'

include $(GOROOT)/src/Make.cmd

.PHONY: initrddata.go
initrddata.go:
	make -C ../cmd ../cmd/foo
	echo package main > $@
	printf 'var testbinary [%d]byte = [%d]byte{\n' `stat -c %s ../cmd/test` `stat -c %s ../cmd/test` >> $@
	hexdump -v -e ' /1 "0x%.2x, " "\n"' ../cmd/foo >> $@
	echo '}' >> $@
