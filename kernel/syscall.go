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

var sysent [NUM_SYSCALLS]Syscall = [NUM_SYSCALLS]Syscall {
	nosyscall,
	syscall_wtf,
	syscall_write,
}

func (p *Process) stack(i uint64) *uint64 {
	return (*uint64)(unsafe.Pointer(uintptr(p.ProcState.sp + i * 8)))
}

func (p *Process) string(s,l /* i'm not really stanley lieber */ uint64) string {
	r := make([]byte, l)
	rh := (*runtime.SliceHeader)(unsafe.Pointer(&r))
	runtime.Memmove(rh.Data, uintptr(s), uint32(l))
	return string(r)
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

func nosyscall(p *Process) {
	p.syserror(SimpleError("No such syscall"))
}

func syscall_write(p *Process) {
	print(p.string(*p.stack(1), *p.stack(2)))
}

func syscall_wtf(p *Process) {
	var s string
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
	sh := (*runtime.StringHeader)(unsafe.Pointer(&s))
	runtime.Memmove(uintptr(buf), sh.Data, l)
	p.errors[k] = ""
end:
	p.ProcState.ax = uint64(len(s))
}
