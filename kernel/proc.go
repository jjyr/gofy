package main

import "unsafe"

const (
	PROCRUN = 1
	PROCWAIT = 2

	MAXPROC = 100
)

type Gstruct struct {
	stackguard, stackbase uintptr
}

type TLS struct {
	gp *Gstruct
	mp uintptr
	g Gstruct
}

type Context struct {
	sp uintptr
	gs Phys
	stack Phys
}

type Proc struct {
	status int
	pid uint64
	ppid uint64
	c Context
}

var (
	proc [MAXPROC]Proc
	lastpid uint64
	curproc int
	tls *TLS
)

func initproc() {
	proc[0] = Proc{status: PROCRUN, pid: 1}
	curproc = 0
	lastpid++
	tls = (*TLS)(unsafe.Pointer(STACK))
	tls.gp = &tls.g
	tls.g.stackguard = STACK - PAGESIZE
	tls.g.stackbase = STACK
}

func savu(*Context)
func retu(*Context)

func procreate() uint64 {
	var slot int
start:
	for k := range proc {
		if proc[k].pid == lastpid {
			lastpid++
			goto start
		}
	}
	for k := range proc {
		if proc[k].pid == 0 {
			slot = k
			goto found
		}
	}
	fuck("no free process slot")
found:
	proc[slot] = Proc{status: PROCRUN, pid: lastpid, ppid: proc[curproc].pid}
	proc[slot].c.stack = falloc(1)
	proc[slot].c.gs = falloc(1)
	v := tmpmap(proc[slot].c.stack)
	memcpy(v, STACK - PAGESIZE, PAGESIZE)

	w := tmpmap(proc[slot].c.gs)
	memcpy(w, STACK, PAGESIZE)
	tmpfree(v)
	tmpfree(w)
	return ssavu(&proc[slot].c)
}

func ssavu(c *Context) (ret uint64) {
	ret = 0
	savu(c)
	return
}

func sretu(c *Context) (ret uint64) {
	ret = 1
	retu(c)
	return
}

func procrastinate() {
	halt()
}

func swtch() uint64 {
	var nextproc, i int
	ssavu(&proc[curproc].c)
start:
	for i = (curproc+1)%MAXPROC; i < MAXPROC; i++ {
		if proc[i].status == PROCRUN {
			nextproc = i
			goto found
		}
	}
	for i = 0; i <= curproc; i++ {
		if proc[i].status == PROCRUN {
			nextproc = i
			goto found
		}
	}
	procrastinate()
	goto start
found:
	curproc = nextproc
	sretu(&proc[curproc].c)
	return 1
}
