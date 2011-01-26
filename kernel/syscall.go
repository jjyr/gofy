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

func syserror() {
}

func nosyscall(*Process) {
}

func syscall_write(p *Process) {
	print(p.string(*p.stack(2), *p.stack(3)))
}

func syscall_wtf(*Process) {
}
