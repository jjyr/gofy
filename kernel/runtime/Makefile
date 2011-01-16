GOOS:=gofykernel
GOARCH:=amd64

include $(GOROOT)/src/Make.inc

TARG=runtime
GOFILES=\
	runtime.go\
	type.go\
	panics.go\
	runtime_defs.go\
	hashmap_defs.go\
	chan_defs.go\

OFILES=\
	rt0.6\
	morestack.6\
	string.6\
	util.6\

GC+= -+

include $(GOROOT)/src/Make.pkg
