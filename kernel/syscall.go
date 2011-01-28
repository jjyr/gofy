package main

import (
	"unsafe"
	"runtime"
)

type Syscall func(*Process)

const (
	SYS_INVALID = iota
	SYS_WTF
	SYS_PWRITE
	SYS_PREAD
	SYS_OPEN
	SYS_CLOSE
	SYS_FORK
	SYS_SETGS
	SYS_SBRK
	NUM_SYSCALLS
)

const (
	MEMREAD = 1
	MEMWRITE = 2
)

var sysent [NUM_SYSCALLS]Syscall = [NUM_SYSCALLS]Syscall {
	nosyscall,
	syscall_wtf,
	syscall_pwrite,
	syscall_pread,
	syscall_open,
	syscall_close,
	syscall_fork,
	syscall_setgs,
	syscall_sbrk,
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
	p.error = e.String()
	p.ProcState.flags |= 1
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

func syscall_wtf(p *Process) {
	var s string

	if p.IsInvalid(p.ProcState.sp, 020, MEMREAD) {
		return
	}
	buf := *p.stack(0)
	l := uint32(*p.stack(1))

	s = p.error
	if uint32(len(s)) > l {
		goto end
	}
	if p.IsInvalid(buf, uint64(l), MEMWRITE) {
		return
	}
	sh := (*runtime.StringHeader)(unsafe.Pointer(&s))
	runtime.Memmove(uintptr(buf), sh.Data, l)
end:
	p.ProcState.ax = uint64(len(s))
}

func (p *Process) getfd(n int) (f File) {
	if n > len(p.fd) {
		return nil
	}
	f = p.fd[n]
	if f == nil {
		p.syserror(SimpleError("no such fd"))
	}
	return
}

func syscall_pread(p *Process) {
	var buf []byte

	if p.IsInvalid(p.ProcState.sp, 040, MEMREAD) {
		return
	}
	fd := *p.stack(0)
	b := *p.stack(1)
	l := *p.stack(2)
	off := *p.stack(3)
	if p.IsInvalid(b, l, MEMWRITE) {
		return
	}
	f := p.getfd(int(fd))
	if f == nil {
		return
	}
	bufh := (*runtime.SliceHeader)(unsafe.Pointer(&buf))
	bufh.Data = uintptr(b)
	bufh.Len = int(l)
	bufh.Cap = int(l)
	n, err := f.PRead(buf, off)
	p.ProcState.ax = n
	if err != nil {
		p.syserror(err)
	}
}

func syscall_pwrite(p *Process) {
	if p.IsInvalid(p.ProcState.sp, 040, MEMREAD) {
		return
	}
	b := *p.stack(1)
	l := *p.stack(2)
	if p.IsInvalid(b, l, MEMREAD) {
		return
	}
	print(p.string(b, l))
}


func syscall_open(p *Process) {
	if p.IsInvalid(p.ProcState.sp, 040, MEMREAD) {
		return
	}
	nameb := *p.stack(0)
	namel := *p.stack(1)
	mode := *p.stack(2)
	perm := *p.stack(3)
	if p.IsInvalid(nameb, namel, MEMREAD) {
		return
	}
	name := p.string(nameb, namel)
	f, err := p.ns.Open(name, int(mode), uint32(perm))
	if err != nil {
		p.syserror(err)
		return
	}
	for k, v := range p.fd {
		if v == nil {
			p.fd[k] = v
			p.ProcState.ax = uint64(k)
			return
		}
	}
	p.ProcState.ax = uint64(len(p.fd))
	p.fd = append(p.fd, f)
}

func syscall_close(p *Process) {
	nosyscall(p)
}

func syscall_fork(p *Process) {
	if p.IsInvalid(p.ProcState.sp, 010, MEMREAD) {
		return
	}
	flags := *p.stack(0)
	p.Fork(flags)
	p.ProcState.ax = 1
}

func syscall_setgs(p *Process) {
	if p.IsInvalid(p.ProcState.sp, 010, MEMREAD) {
		return
	}
	p.ProcState.gs = *p.stack(0)
}

func syscall_sbrk(p *Process) {
	if p.IsInvalid(p.ProcState.sp, 010, MEMREAD) {
		return
	}
	size := *p.stack(0)
	p.Allocate(size)
	p.ProcState.ax = uint64(p.highest)
}
