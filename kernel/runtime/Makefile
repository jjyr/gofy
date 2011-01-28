GOOS:=gofykernel
GOARCH:=amd64

include $(GOROOT)/src/Make.inc

TARG=runtime

SIZE=64

# TODO(kaib): fix register allocation to honor extern register so we
# can enable optimizations again.
CFLAGS=-FVw

GOFILES=\
	console.go\
	cp437.go\
	debug.go\
	error.go\
	extern.go\
	sig.go\
	type.go\
	version.go\
	chan_defs.go\
	hashmap_defs.go\
	iface_defs.go\
	malloc_defs.go\
	mheapmap64_defs.go\
	runtime_defs.go\

OFILES=\
	rt0.$O\
	asm.$O\
	interrupt.$O\
	interrupt_s.$O\
	eia232.$O\
	syscall.$O\
	chan.$O\
	closure.$O\
	float.$O\
	complex.$O\
	hashmap.$O\
	iface.$O\
	malloc.$O\
	mcache.$O\
	mcentral.$O\
	mem.$O\
	memmove.$O\
	mfinal.$O\
	mfixalloc.$O\
	mgc0.$O\
	mheap.$O\
	mheapmap$(SIZE).$O\
	mprof.$O\
	msize.$O\
	print.$O\
	proc.$O\
	reflect.$O\
	rune.$O\
	runtime.$O\
	runtime1.$O\
	sema.$O\
	signal.$O\
	sigqueue.$O\
	slice.$O\
	string.$O\
	symtab.$O\
	thread.$O\
	traceback.$O\
	$(OFILES_$(GOARCH))\
	$(OFILES_$(GOOS))\

HFILES=\
	runtime.h\
	hashmap.h\
	malloc.h\
	asm.h\
	os.h\
	signals.h\
	defs.h\

GOFILES+=$(GOFILES_$(GOOS))

# special, out of the way compiler flag that means "add runtime metadata to output"
GC+= -+

%.$O:	%.c asm.h
	$(CC) $(CFLAGS) $<

%.$O:	%.s asm.h
	$(AS) $<


include $(GOROOT)/src/Make.pkg

$(pkgdir)/%.h: %.h
	@test -d $(QUOTED_GOROOT)/pkg && mkdir -p $(pkgdir)
	cp $< $@

clean: clean-local

clean-local:
	rm -f goc2c mkversion version.go asm.h runtime.acid.* $$(ls *.goc | sed 's/goc$$/c/')

asm.h: mkasmh.sh runtime.acid.$(GOARCH)
	./mkasmh.sh >$@.x
	mv -f $@.x $@

goc2c: goc2c.c
	quietgcc -o $@ $<

mkversion: mkversion.c
	quietgcc -o $@ -I "$(GOROOT)/include" $< "$(GOROOT)/lib/lib9.a"

version.go: mkversion
	./mkversion >version.go

%.c:	%.goc goc2c
	./goc2c `pwd`/$< > $@.tmp
	mv -f $@.tmp $@

# for discovering offsets inside structs when debugging
runtime.acid.$(GOARCH): runtime.h proc.c
	$(CC) $(CFLAGS) -a proc.c >$@

# 386 traceback is really amd64 traceback
ifeq ($(GOARCH),386)
traceback.$O:	amd64/traceback.c
	$(CC) $(CFLAGS) $<
endif
