package main

import (
	"unsafe"
	"runtime"
)

type Syscall func(*Process)

const (
	SYS_INVALID = iota
	SYS_WTF
	SYS_WRITE
	NUM_SYSCALLS
)

const (
	MEMREAD = 1
	MEMWRITE = 2
)

var sysent [NUM_SYSCALLS]Syscall = [NUM_SYSCALLS]Syscall {
	nosyscall,
	syscall_wtf,
	syscall_write,
}

func (p *Process) string(s,l /* i'm not really stanley lieber */ uint64) string {
	r := make([]byte, l)
	rh := (*runtime.SliceHeader)(unsafe.Pointer(&r))
	runtime.Memmove(rh.Data, uintptr(s), uint32(l))
	return string(r)
}

func (p *Process) stack(i uint64) *uint64 {
	return (*uint64)(unsafe.Pointer(uintptr(p.ProcState.sp + i * 8)))
}

func (p *Process) syserror(e Error) {
	var kf int
	str := e.String()
	for k, v := range p.errors {
		if v == "" {
			p.errors[k] = v
			kf = k
			goto found
		}
	}
	kf = len(p.errors)
	p.errors = append(p.errors, str)
found:
	p.ProcState.ax = uint64(kf + 1)
	p.ProcState.rflags |= 1
}

func (p *Process) IsInvalid(a uint64, len uint64, t int) bool {
	if a < USERSTART {
		goto hitler
	}
	if a + len >= uint64(p.highest) {
		goto hitler
	}
	return false
hitler: // the user is presumably evil
	p.syserror(SimpleError("invalid memory reference"))
	return true
}

func nosyscall(p *Process) {
	p.syserror(SimpleError("No such syscall"))
}

func syscall_write(p *Process) {
	if p.IsInvalid(p.ProcState.sp, 030, MEMREAD) {
		return
	}
	b := *p.stack(1)
	l := *p.stack(2)
	if p.IsInvalid(b, l, MEMREAD) {
		return
	}
	print(p.string(b, l))
}

func syscall_wtf(p *Process) {
	var s string

	if p.IsInvalid(p.ProcState.sp, 030, MEMREAD) {
		return
	}
	k := *p.stack(0) - 1
	buf := *p.stack(1)
	l := uint32(*p.stack(2))

	if int(k) >= len(p.errors) {
		goto end
	}
	s = p.errors[k]
	if uint32(len(s)) > l {
		goto end
	}
	if p.IsInvalid(buf, uint64(l), MEMWRITE) {
		return
	}
	sh := (*runtime.StringHeader)(unsafe.Pointer(&s))
	runtime.Memmove(uintptr(buf), sh.Data, l)
	p.errors[k] = ""
end:
	p.ProcState.ax = uint64(len(s))
}
