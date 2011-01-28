GOOS:=gofy
GOARCH:=amd64

include $(GOROOT)/src/Make.inc

DEPS=../runtime
TARG=test
GOFILES=\
	test.go\

LD+= -H 2 -T0x40000028

include $(GOROOT)/src/Make.cmd

%: %.asm
	nasm $< -o $@
