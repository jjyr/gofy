GOOS:=gofykernel
GOARCH:=amd64

include $(GOROOT)/src/Make.inc

TARG=runtime
GOFILES=\
	runtime.go\
	type.go\

OFILES=\
	rt0.6\
	morestack.6\
	string.6\

GC+= -+

include $(GOROOT)/src/Make.pkg
